package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
)

type Hub struct {
	config     Config
	rooms      map[string]*Room
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub(config Config) *Hub {
	return &Hub{
		config:     config,
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.RLock()
			room, exists := h.rooms[client.RoomCode]
			h.mu.RUnlock()

			if exists {
				room.AddClient(client)
				log.Printf("Client %s joined room %s", client.Id, client.RoomCode)

				// Notify everyone in the room
				room.Broadcast(newMessage(MsgPlayerJoined, PlayerInfo{
					Id:    client.Id,
					Name:  client.Name,
					Score: 0,
				}))
			}

		case client := <-h.unregister:
			h.mu.RLock()
			room, exists := h.rooms[client.RoomCode]
			h.mu.RUnlock()

			if exists {
				room.RemoveClient(client.Id)
				log.Printf("Client %s left room %s", client.Id, client.RoomCode)

				room.Broadcast(newMessage(MsgPlayerLeft, map[string]string{
					"playerId": client.Id,
				}))

				// Delete empty rooms
				if len(room.Clients) == 0 {
					h.mu.Lock()
					delete(h.rooms, client.RoomCode)
					h.mu.Unlock()
					log.Printf("Room %s deleted (empty)", client.RoomCode)
				}
			}
		}
	}
}

func (h *Hub) CreateRoom(playerName string) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	code := fmt.Sprintf("ROOM%04d", rand.Intn(9000)+1000)
	room := NewRoom(code)
	h.rooms[code] = room

	log.Printf("Room %s created by %s", code, playerName)
	return code, nil
}

func (h *Hub) GetRoom(code string) (*Room, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	room, exists := h.rooms[code]
	return room, exists
}
