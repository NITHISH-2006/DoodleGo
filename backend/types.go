package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"
)

type MessageType string

const (
	// Client -> Server
	MsgCreateRoom  MessageType = "CREATE_ROOM"
	MsgJoinRoom    MessageType = "JOIN_ROOM"
	MsgStartGame   MessageType = "START_GAME"
	MsgDrawStroke  MessageType = "DRAW_STROKE"
	MsgUndoStroke  MessageType = "UNDO_STROKE"
	MsgClearCanvas MessageType = "CLEAR_CANVAS"
	MsgGuess       MessageType = "GUESS"
	MsgLeaveRoom   MessageType = "LEAVE_ROOM"
	MsgPong        MessageType = "PONG"

	// Server -> Client
	MsgRoomJoined    MessageType = "ROOM_JOINED"
	MsgPlayerJoined  MessageType = "PLAYER_JOINED"
	MsgPlayerLeft    MessageType = "PLAYER_LEFT"
	MsgGameStarted   MessageType = "GAME_STARTED"
	MsgStroke        MessageType = "STROKE"
	MsgCanvasState   MessageType = "CANVAS_STATE"
	MsgGuessReceived MessageType = "GUESS_RECEIVED"
	MsgRoundEnd      MessageType = "ROUND_END"
	MsgError         MessageType = "ERROR"
	MsgPing          MessageType = "PING"
	MsgScoreUpdate   MessageType = "SCORE_UPDATE"
)

// Message is the universal envelope for all WebSocket communication.
type Message struct {
	Type      MessageType     `json:"type"`
	PlayerId  string          `json:"playerId,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
}

// Point is a single normalized canvas coordinate.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// StrokeData describes one chunk of a drawing gesture. All chunks belonging to
// a single pointer gesture share the same StrokeId so undo can remove the whole
// gesture at once.
type StrokeData struct {
	StrokeId string  `json:"strokeId,omitempty"`
	OwnerId  string  `json:"ownerId,omitempty"`
	Points   []Point `json:"points"`
	Color    string  `json:"color"`
	Size     int     `json:"size"`
	Tool     string  `json:"tool,omitempty"`
	Opacity  float64 `json:"opacity,omitempty"`
}

type JoinRoomData struct {
	RoomCode   string `json:"roomCode"`
	PlayerName string `json:"playerName"`
}

type CreateRoomData struct {
	PlayerName    string `json:"playerName"`
	Mode          string `json:"mode"`
	RoundDuration int    `json:"roundDuration"` // seconds
}

type PlayerInfo struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

type GameStartedData struct {
	DrawerId    string `json:"drawerId"`
	DrawerName  string `json:"drawerName"`
	Word        string `json:"word"` // real word for drawer, empty for everyone else
	WordLength  int    `json:"wordLength"`
	MaskedWord  string `json:"maskedWord"`
	RoundNum    int    `json:"roundNum"`
	TimeLeft    int    `json:"timeLeft"`
	RoundEndsAt int64  `json:"roundEndsAt"`
}

type RoomJoinedData struct {
	RoomCode      string           `json:"roomCode"`
	PlayerId      string           `json:"playerId"`
	Players       []PlayerInfo     `json:"players"`
	Status        string           `json:"status"`
	Mode          string           `json:"mode"`
	RoundDuration int              `json:"roundDuration"`
	MaxPlayers    int              `json:"maxPlayers"`
	Game          *GameStartedData `json:"game,omitempty"`
	Strokes       []StrokeData     `json:"strokes,omitempty"`
}

type RoundEndData struct {
	Word    string       `json:"word"`
	Players []PlayerInfo `json:"players"`
}

type GuessData struct {
	Text string `json:"text"`
}

type GuessReceivedData struct {
	GuesserId   string `json:"guesserId"`
	GuesserName string `json:"guesserName"`
	Text        string `json:"text"`
	IsCorrect   bool   `json:"isCorrect"`
	Message     string `json:"message"`
}

type PlayerScoreUpdate struct {
	PlayerId string `json:"playerId"`
	Score    int    `json:"score"`
}

type CanvasStateData struct {
	Strokes []StrokeData `json:"strokes"`
}

// Helper to create timestamped messages.
func newMessage(msgType MessageType, data interface{}) *Message {
	b, _ := json.Marshal(data)
	return &Message{
		Type:      msgType,
		Timestamp: time.Now().UnixMilli(),
		Data:      b,
	}
}

func randomString(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	cryptoBytes := make([]byte, n)
	if _, err := rand.Read(cryptoBytes); err != nil {
		for i := range b {
			b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		}
		return string(b)
	}
	for i := range b {
		b[i] = letters[int(cryptoBytes[i])%len(letters)]
	}
	return string(b)
}

func randomID(prefix string) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return prefix + randomString(8)
	}
	return prefix + hex.EncodeToString(b)
}
