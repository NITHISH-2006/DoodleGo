package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type Room struct {
	Code          string
	Clients       map[string]*Client
	mu            sync.RWMutex
	State         string // "lobby" or "playing"
	CurrentDrawer string
	CurrentWord   string
	RoundNum      int
	Scores        map[string]int
	RoundDuration time.Duration
	roundTimer    *time.Timer
}

func NewRoom(code string, roundDuration time.Duration) *Room {
	return &Room{
		Code:          code,
		Clients:       make(map[string]*Client),
		Scores:        make(map[string]int),
		State:         "lobby",
		RoundDuration: roundDuration,
	}
}

func (r *Room) AddClient(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Clients[c.Id] = c
	r.Scores[c.Id] = 0
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
	return r.playersLocked()
}

func (r *Room) playersLocked() []PlayerInfo {
	players := make([]PlayerInfo, 0, len(r.Clients))
	for _, c := range r.Clients {
		players = append(players, PlayerInfo{Id: c.Id, Name: c.Name, Score: r.Scores[c.Id]})
	}
	return players
}

func (r *Room) StartGame() {
	r.mu.Lock()

	if len(r.Clients) < 2 || r.State == "playing" {
		r.mu.Unlock()
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
	r.startRoundTimerLocked()

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
	r.mu.Unlock()

	log.Printf("Room %s started. Drawer: %s, Word: %s", r.Code, drawerId, word)
}

func (r *Room) startRoundTimerLocked() {
	if r.roundTimer != nil {
		r.roundTimer.Stop()
	}
	duration := r.RoundDuration
	if duration <= 0 {
		duration = 90 * time.Second
	}
	r.roundTimer = time.AfterFunc(duration, r.endRound)
}

func (r *Room) endRound() {
	r.mu.Lock()
	if r.State != "playing" {
		r.mu.Unlock()
		return
	}

	r.State = "lobby"
	r.CurrentDrawer = ""
	word := r.CurrentWord
	r.CurrentWord = ""
	players := r.playersLocked()
	r.mu.Unlock()

	r.Broadcast(newMessage(MsgRoundEnd, RoundEndData{Word: word, Players: players}))
	time.AfterFunc(5*time.Second, func() {
		r.mu.RLock()
		canContinue := len(r.Clients) >= 2 && r.State == "lobby"
		r.mu.RUnlock()
		if canContinue {
			r.StartGame()
		}
	})
}

func (r *Room) HandleGuess(clientId, text string) {
	r.mu.Lock()

	if r.State != "playing" {
		r.mu.Unlock()
		return
	}

	// Drawer cannot guess
	if clientId == r.CurrentDrawer {
		r.mu.Unlock()
		return
	}

	client, exists := r.Clients[clientId]
	if !exists {
		r.mu.Unlock()
		return
	}

	isCorrect := strings.EqualFold(strings.TrimSpace(text), r.CurrentWord)

	if isCorrect {
		r.Scores[clientId] += 10
	}

	// Broadcast the guess result
	msg := newMessage(MsgGuessReceived, GuessReceivedData{
		GuesserId:   clientId,
		GuesserName: client.Name,
		Text:        text,
		IsCorrect:   isCorrect,
		Message:     fmt.Sprintf("%s guessed '%s'", client.Name, text),
	})

	var scoreMsg *Message
	if isCorrect {
		scoreMsg = newMessage(MsgScoreUpdate, PlayerScoreUpdate{
			PlayerId: clientId,
			Score:    r.Scores[clientId],
		})
		log.Printf("Room %s: %s guessed correctly! Score: %d", r.Code, client.Name, r.Scores[clientId])
	}
	r.mu.Unlock()

	// Broadcast after releasing the room lock because Broadcast acquires a read lock.
	r.Broadcast(msg)
	if scoreMsg != nil {
		r.Broadcast(scoreMsg)
	}
}
