package main

import (
	"errors"
	"log"
	"sync"
	"time"
)

var (
	errRoomNotFound = errors.New("room not found")
	errRoomFull     = errors.New("room is full")
	errRoomClosed   = errors.New("room is closed")
)

type RoomOptions struct {
	Mode          string
	RoundDuration time.Duration
	MaxPlayers    int
}

type Hub struct {
	config Config
	rooms  map[string]*Room
	mu     sync.RWMutex
}

func NewHub(config Config) *Hub {
	return &Hub{
		config: config,
		rooms:  make(map[string]*Room),
	}
}

func (h *Hub) CreateRoom(options RoomOptions) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if options.MaxPlayers <= 1 {
		options.MaxPlayers = h.config.MaxPlayers
	}
	if options.RoundDuration <= 0 {
		options.RoundDuration = h.config.RoundDuration
	}
	if options.Mode != modeTogether {
		options.Mode = modeClassic
	}

	code := ""
	for attempt := 0; attempt < 200; attempt++ {
		candidate := "ROOM" + randomString(4)
		if _, exists := h.rooms[candidate]; !exists {
			code = candidate
			break
		}
	}
	if code == "" {
		return "", errors.New("could not allocate a room code")
	}

	h.rooms[code] = NewRoom(code, options)
	log.Printf("Room %s created (mode=%s, round=%s)", code, options.Mode, options.RoundDuration)
	return code, nil
}

// Register atomically checks the room exists and adds the client. This removes
// the old check-then-register race that could hang a joining connection.
func (h *Hub) Register(client *Client) (*Room, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[client.RoomCode]
	if !exists {
		return nil, errRoomNotFound
	}
	if err := room.AddClient(client); err != nil {
		return nil, err
	}
	return room, nil
}

// Unregister removes the client and deletes the room once it is empty.
func (h *Hub) Unregister(client *Client) (*Room, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[client.RoomCode]
	if !exists {
		return nil, false
	}

	removed := room.RemoveClient(client.Id)
	if removed {
		room.BroadcastExcept(client.Id, newMessage(MsgPlayerLeft, map[string]string{
			"playerId": client.Id,
		}))
		log.Printf("Client %s left room %s", client.Id, client.RoomCode)
	}
	if room.Count() == 0 {
		room.Close()
		delete(h.rooms, client.RoomCode)
		log.Printf("Room %s deleted (empty)", client.RoomCode)
	}
	return room, removed
}

func (h *Hub) GetRoom(code string) (*Room, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	room, exists := h.rooms[code]
	return room, exists
}
