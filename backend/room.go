package main

import (
	"log"
	"math/rand"
	"sync"
)

type Room struct {
	Code          string
	Clients       map[string]*Client
	mu            sync.RWMutex
	State         string // "lobby" or "playing"
	CurrentDrawer string
	CurrentWord   string
	RoundNum      int
}

func NewRoom(code string) *Room {
	return &Room{
		Code:    code,
		Clients: make(map[string]*Client),
		State:   "lobby",
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
		players = append(players, PlayerInfo{Id: c.Id, Name: c.Name, Score: 0})
	}
	return players
}

func (r *Room) StartGame() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Clients) < 2 {
		return
	}

	// Choose random drawer
	ids := make([]string, 0, len(r.Clients))
	for id := range r.Clients {
		ids = append(ids, id)
	}
	drawerId := ids[rand.Intn(len(ids))]

	// Simple word list for now
	words := []string{"apple", "banana", "cat", "dog", "house", "tree", "car", "moon", "star", "book"}
	word := words[rand.Intn(len(words))]

	r.State = "playing"
	r.CurrentDrawer = drawerId
	r.CurrentWord = word
	r.RoundNum++

	// Send different messages to drawer vs others
	for _, client := range r.Clients {
		wordToSend := "*****"
		if client.Id == drawerId {
			wordToSend = word
		}

		msg := newMessage(MsgGameStarted, GameStartedData{
			DrawerId: drawerId,
			Word:     wordToSend,
			RoundNum: r.RoundNum,
			TimeLeft: 90,
		})
		client.send <- msg
	}

	log.Printf("Room %s started. Drawer: %s, Word: %s", r.Code, drawerId, word)
}
