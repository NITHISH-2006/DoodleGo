package main

import (
	"log"
	"sync"
)

type Room struct {
	Code    string
	Clients map[string]*Client
	mu      sync.RWMutex
}

func NewRoom(code string) *Room {
	return &Room{
		Code:    code,
		Clients: make(map[string]*Client),
	}
}

func (r *Room) AddClient(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Clients[c.Id] = c
}

func (r *Room) RemoveClient(clientId string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Clients, clientId)
}

func (r *Room) Broadcast(msg *Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, client := range r.Clients {
		select {
		case client.send <- msg:
		default:
			log.Printf("Client %s send buffer full", client.Id)
		}
	}
}

func (r *Room) GetPlayers() []PlayerInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	players := make([]PlayerInfo, 0, len(r.Clients))
	for _, c := range r.Clients {
		players = append(players, PlayerInfo{
			Id:    c.Id,
			Name:  c.Name,
			Score: 0,
		})
	}
	return players
}
