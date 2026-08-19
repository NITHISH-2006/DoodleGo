package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func sendError(conn *websocket.Conn, msg string) {
	conn.WriteJSON(newMessage(MsgError, map[string]string{"error": msg}))
	conn.Close()
}

func handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	// Read the first message (must be CREATE_ROOM or JOIN_ROOM)
	var initMsg Message
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	err = conn.ReadJSON(&initMsg)
	if err != nil {
		log.Printf("Initial message error: %v", err)
		conn.Close()
		return
	}

	var client *Client
	var roomCode string
	var playerName string

	switch initMsg.Type {
	case MsgCreateRoom:
		var data CreateRoomData
		if err := json.Unmarshal(initMsg.Data, &data); err != nil {
			sendError(conn, "Invalid CREATE_ROOM data")
			return
		}
		playerName = data.PlayerName

		code, err := hub.CreateRoom(playerName)
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
		playerName = data.PlayerName
		roomCode = data.RoomCode

		// Check if room exists
		if _, exists := hub.GetRoom(roomCode); !exists {
			sendError(conn, "Room not found")
			return
		}
		// TODO: validate that the room exists

	default:
		sendError(conn, "First message must be CREATE_ROOM or JOIN_ROOM")
		return
	}

	// Create unique player ID
	playerId := "player-" + randomString(8)

	// Create the client
	client = NewClient(playerId, playerName, roomCode, conn, hub)

	// Register with Hub and wait until the room membership is visible.
	hub.register <- client
	<-client.registered

	room, _ := hub.GetRoom(roomCode)
	players := room.GetPlayers()
	// Send ROOM_JOINED response to the client
	response := newMessage(MsgRoomJoined, RoomJoinedData{
		RoomCode: roomCode,
		PlayerId: playerId,
		Players:  players,
		Status:   "lobby",
	})
	client.send <- response

	// This keeps the connection alive (read + write pumps)
	client.Serve()
}
