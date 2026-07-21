package main

import (
	"encoding/json"
	"time"
)

type MessageType string

const (
	// Client → Server
	MsgCreateRoom MessageType = "CREATE_ROOM"
	MsgJoinRoom   MessageType = "JOIN_ROOM"
	MsgStartGame  MessageType = "START_GAME"
	MsgDrawStroke MessageType = "DRAW_STROKE"
	MsgGuess      MessageType = "GUESS"
	MsgLeaveRoom  MessageType = "LEAVE_ROOM"
	MsgPong       MessageType = "PONG"

	// Server → Client
	MsgRoomJoined    MessageType = "ROOM_JOINED"
	MsgPlayerJoined  MessageType = "PLAYER_JOINED"
	MsgPlayerLeft    MessageType = "PLAYER_LEFT"
	MsgGameStarted   MessageType = "GAME_STARTED"
	MsgStroke        MessageType = "STROKE"
	MsgGuessReceived MessageType = "GUESS_RECEIVED"
	MsgRoundEnd      MessageType = "ROUND_END"
	MsgError         MessageType = "ERROR"
	MsgPing          MessageType = "PING"
)

// Message is the universal envelope for all WebSocket communication
type Message struct {
	Type      MessageType     `json:"type"`
	PlayerId  string          `json:"playerId,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"` // Flexible payload
	Timestamp int64           `json:"timestamp,omitempty"`
}

// Common data structures

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type StrokeData struct {
	Points []Point `json:"points"`
	Color  string  `json:"color"`
	Size   int     `json:"size"`
}

type JoinRoomData struct {
	RoomCode   string `json:"roomCode"`
	PlayerName string `json:"playerName"`
}

type CreateRoomData struct {
	PlayerName string `json:"playerName"`
}

type RoomJoinedData struct {
	RoomCode string       `json:"roomCode"`
	PlayerId string       `json:"playerId"`
	Players  []PlayerInfo `json:"players"`
	Status   string       `json:"status"`
}

type PlayerInfo struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

// Helper to create timestamped messages
func newMessage(msgType MessageType, data interface{}) *Message {
	b, _ := json.Marshal(data)
	return &Message{
		Type:      msgType,
		Timestamp: time.Now().UnixMilli(),
		Data:      b,
	}
}
