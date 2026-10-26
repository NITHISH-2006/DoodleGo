package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

var allowedRoundDurations = map[int]bool{
	30:  true,
	60:  true,
	90:  true,
	120: true,
	180: true,
}

func buildUpgrader(config Config) websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return originAllowed(config, r)
		},
	}
}

func originAllowed(config Config, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	for _, allowed := range config.AllowedOrigins {
		if allowed == "*" || strings.EqualFold(allowed, origin) {
			return true
		}
	}
	return false
}

func sendError(conn *websocket.Conn, msg string) {
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	conn.WriteJSON(newMessage(MsgError, map[string]string{"error": msg}))
	conn.Close()
}

func normalizeName(name string) (string, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", false
	}
	if utf8.RuneCountInString(name) > 24 {
		runes := []rune(name)
		name = string(runes[:24])
	}
	return name, true
}

func normalizeRoundDuration(seconds int) time.Duration {
	if !allowedRoundDurations[seconds] {
		return 90 * time.Second
	}
	return time.Duration(seconds) * time.Second
}

func handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	upgrader := buildUpgrader(hub.config)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	var initMsg Message
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	if err := conn.ReadJSON(&initMsg); err != nil {
		log.Printf("Initial message error: %v", err)
		conn.Close()
		return
	}

	var roomCode string
	var playerName string

	switch initMsg.Type {
	case MsgCreateRoom:
		var data CreateRoomData
		if err := json.Unmarshal(initMsg.Data, &data); err != nil {
			sendError(conn, "Invalid CREATE_ROOM data")
			return
		}
		name, ok := normalizeName(data.PlayerName)
		if !ok {
			sendError(conn, "Please enter a name")
			return
		}
		playerName = name

		mode := modeClassic
		if data.Mode == modeTogether {
			mode = modeTogether
		}
		code, err := hub.CreateRoom(RoomOptions{
			Mode:              mode,
			Difficulty:        normalizeDifficulty(data.Difficulty),
			TogetherActivity:  normalizeTogetherActivity(data.TogetherActivity),
			RoundDuration:     normalizeRoundDuration(data.RoundDuration),
			RelayTurnDuration: normalizeRelayTurnDuration(time.Duration(data.RelayTurnDuration) * time.Second),
			MaxPlayers:        hub.config.MaxPlayers,
		})
		if err != nil {
			sendError(conn, err.Error())
			return
		}
		roomCode = code

	case MsgJoinRoom:
		var data JoinRoomData
		if err := json.Unmarshal(initMsg.Data, &data); err != nil {
			sendError(conn, "Invalid JOIN_ROOM data")
			return
		}
		name, ok := normalizeName(data.PlayerName)
		if !ok {
			sendError(conn, "Please enter a name")
			return
		}
		roomCode = strings.ToUpper(strings.TrimSpace(data.RoomCode))
		if roomCode == "" {
			sendError(conn, "Please enter a room code")
			return
		}
		playerName = name

	default:
		sendError(conn, "First message must be CREATE_ROOM or JOIN_ROOM")
		return
	}

	playerId := randomID("player-")
	client := NewClient(playerId, playerName, roomCode, conn, hub)

	room, err := hub.Register(client)
	if err != nil {
		switch err {
		case errRoomNotFound:
			sendError(conn, "Room not found")
		case errRoomFull:
			sendError(conn, "That room is full")
		default:
			sendError(conn, "Could not join the room")
		}
		return
	}

	snapshot := room.Snapshot(client.Id)
	response := newMessage(MsgRoomJoined, RoomJoinedData{
		RoomCode:          roomCode,
		PlayerId:          playerId,
		Players:           snapshot.Players,
		Status:            snapshot.Status,
		Mode:              snapshot.Mode,
		Difficulty:        snapshot.Difficulty,
		TogetherActivity:  snapshot.TogetherActivity,
		RoundDuration:     snapshot.RoundDuration,
		RelayTurnDuration: snapshot.RelayTurnDuration,
		MaxPlayers:        snapshot.MaxPlayers,
		Game:              snapshot.Game,
		Mission:           snapshot.Mission,
		Relay:             snapshot.Relay,
		Strokes:           snapshot.Strokes,
	})
	client.send <- response

	room.BroadcastExcept(client.Id, newMessage(MsgPlayerJoined, PlayerInfo{
		Id:    client.Id,
		Name:  client.Name,
		Score: 0,
	}))
	room.TryResumeRelay()

	log.Printf("Client %s (%s) joined room %s", client.Id, client.Name, client.RoomCode)
	client.Serve()
}
