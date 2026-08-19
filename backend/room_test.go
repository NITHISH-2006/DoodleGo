package main

import (
	"encoding/json"
	"testing"
	"time"
)

func testClient(id, name string) *Client {
	return &Client{
		Id:   id,
		Name: name,
		send: make(chan *Message, 16),
	}
}

func readMessage(t *testing.T, client *Client) *Message {
	t.Helper()
	select {
	case message := <-client.send:
		return message
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
		return nil
	}
}

func TestHandleGuessAwardsScoreAndBroadcasts(t *testing.T) {
	room := NewRoom("ROOM1234", time.Minute)
	drawer := testClient("drawer", "Alice")
	guesser := testClient("guesser", "Bob")
	room.AddClient(drawer)
	room.AddClient(guesser)
	room.mu.Lock()
	room.State = "playing"
	room.CurrentDrawer = drawer.Id
	room.CurrentWord = "apple"
	room.mu.Unlock()

	room.HandleGuess(guesser.Id, " Apple ")

	guessMessage := readMessage(t, drawer)
	if guessMessage.Type != MsgGuessReceived {
		t.Fatalf("expected GUESS_RECEIVED, got %s", guessMessage.Type)
	}
	var guess GuessReceivedData
	if err := json.Unmarshal(guessMessage.Data, &guess); err != nil {
		t.Fatal(err)
	}
	if !guess.IsCorrect {
		t.Fatal("expected guess to be correct")
	}

	scoreMessage := readMessage(t, drawer)
	if scoreMessage.Type != MsgScoreUpdate {
		t.Fatalf("expected SCORE_UPDATE, got %s", scoreMessage.Type)
	}
	if room.Scores[guesser.Id] != 10 {
		t.Fatalf("expected score 10, got %d", room.Scores[guesser.Id])
	}
}

func TestRoundEndsAndRevealsWord(t *testing.T) {
	room := NewRoom("ROOM1234", 10*time.Millisecond)
	first := testClient("first", "Alice")
	second := testClient("second", "Bob")
	room.AddClient(first)
	room.AddClient(second)
	room.StartGame()

	readMessage(t, first)
	readMessage(t, second)
	endMessage := readMessage(t, first)
	if endMessage.Type != MsgRoundEnd {
		t.Fatalf("expected ROUND_END, got %s", endMessage.Type)
	}
	var end RoundEndData
	if err := json.Unmarshal(endMessage.Data, &end); err != nil {
		t.Fatal(err)
	}
	if end.Word == "" {
		t.Fatal("expected round-end message to reveal the word")
	}
	room.mu.RLock()
	state := room.State
	room.mu.RUnlock()
	if state != "lobby" {
		t.Fatalf("expected lobby after round end, got %s", state)
	}
}
