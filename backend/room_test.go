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
		send: make(chan *Message, 64),
	}
}

func testRoom(mode string, duration time.Duration) *Room {
	return NewRoom("ROOM1234", RoomOptions{
		Mode:          mode,
		RoundDuration: duration,
		MaxPlayers:    20,
	})
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

func decodeGuess(t *testing.T, message *Message) GuessReceivedData {
	t.Helper()
	var guess GuessReceivedData
	if err := json.Unmarshal(message.Data, &guess); err != nil {
		t.Fatal(err)
	}
	return guess
}

func TestHandleGuessAwardsScoreOnlyOncePerRound(t *testing.T) {
	room := testRoom(modeClassic, time.Minute)
	drawer := testClient("drawer", "Alice")
	guesser := testClient("guesser", "Bob")
	room.AddClient(drawer)
	room.AddClient(guesser)
	room.mu.Lock()
	room.State = statePlaying
	room.CurrentDrawer = drawer.Id
	room.CurrentWord = "apple"
	room.mu.Unlock()

	room.HandleGuess(guesser.Id, " Apple ")
	room.HandleGuess(guesser.Id, "apple")

	if room.Scores[guesser.Id] != 10 {
		t.Fatalf("expected score 10, got %d", room.Scores[guesser.Id])
	}

	guessMessage := readMessage(t, drawer)
	if guessMessage.Type != MsgGuessReceived {
		t.Fatalf("expected GUESS_RECEIVED, got %s", guessMessage.Type)
	}
	guess := decodeGuess(t, guessMessage)
	if !guess.IsCorrect {
		t.Fatal("expected guess to be correct")
	}
	if guess.Text != "" {
		t.Fatalf("correct guess leaked the word: %q", guess.Text)
	}

	scoreMessage := readMessage(t, drawer)
	if scoreMessage.Type != MsgScoreUpdate {
		t.Fatalf("expected SCORE_UPDATE, got %s", scoreMessage.Type)
	}

	// The duplicate correct guess must not generate another score update.
	select {
	case message := <-drawer.send:
		t.Fatalf("unexpected extra message after duplicate guess: %s", message.Type)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestWrongGuessTextIsStillBroadcast(t *testing.T) {
	room := testRoom(modeClassic, time.Minute)
	drawer := testClient("drawer", "Alice")
	guesser := testClient("guesser", "Bob")
	room.AddClient(drawer)
	room.AddClient(guesser)
	room.mu.Lock()
	room.State = statePlaying
	room.CurrentDrawer = drawer.Id
	room.CurrentWord = "apple"
	room.mu.Unlock()

	room.HandleGuess(guesser.Id, "orange")

	guess := decodeGuess(t, readMessage(t, drawer))
	if guess.IsCorrect || guess.Text != "orange" {
		t.Fatalf("unexpected guess broadcast: %+v", guess)
	}
	if room.Scores[guesser.Id] != 0 {
		t.Fatalf("expected no score, got %d", room.Scores[guesser.Id])
	}
}

func TestRoundUsesConfiguredDuration(t *testing.T) {
	room := testRoom(modeClassic, 30*time.Second)
	first := testClient("first", "Alice")
	second := testClient("second", "Bob")
	room.AddClient(first)
	room.AddClient(second)
	room.StartGame()

	message := readMessage(t, first)
	if message.Type != MsgGameStarted {
		t.Fatalf("expected GAME_STARTED, got %s", message.Type)
	}
	var started GameStartedData
	if err := json.Unmarshal(message.Data, &started); err != nil {
		t.Fatal(err)
	}
	if started.TimeLeft < 29 || started.TimeLeft > 30 {
		t.Fatalf("expected ~30s remaining, got %d", started.TimeLeft)
	}
	if started.RoundEndsAt == 0 {
		t.Fatal("expected roundEndsAt timestamp")
	}
}

func TestDrawerSeesWordAndOthersOnlySeeMask(t *testing.T) {
	room := testRoom(modeClassic, time.Minute)
	first := testClient("first", "Alice")
	second := testClient("second", "Bob")
	room.AddClient(first)
	room.AddClient(second)
	room.StartGame()

	var drawerMsg, guesserMsg GameStartedData
	for _, client := range []*Client{first, second} {
		message := readMessage(t, client)
		var started GameStartedData
		if err := json.Unmarshal(message.Data, &started); err != nil {
			t.Fatal(err)
		}
		if started.DrawerId == client.Id {
			drawerMsg = started
		} else {
			guesserMsg = started
		}
	}

	if drawerMsg.Word == "" || guesserMsg.Word != "" {
		t.Fatalf("word visible to wrong player: drawer=%q guesser=%q", drawerMsg.Word, guesserMsg.Word)
	}
	if guesserMsg.WordLength == 0 || guesserMsg.MaskedWord == "" {
		t.Fatal("expected word length and mask for non-drawer")
	}
}

func TestRoundEndsAndRevealsWord(t *testing.T) {
	room := testRoom(modeClassic, 10*time.Millisecond)
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
	if state != stateLobby {
		t.Fatalf("expected lobby after round end, got %s", state)
	}
}

func TestTogetherModeAllowsBothPlayersToDrawAndUndo(t *testing.T) {
	room := testRoom(modeTogether, time.Minute)
	first := testClient("first", "Alice")
	second := testClient("second", "Bob")
	room.AddClient(first)
	room.AddClient(second)

	stroke := StrokeData{
		StrokeId: "gesture-1",
		Points:   []Point{{X: 1, Y: 2}, {X: 3, Y: 4}},
		Color:    "#d64f3b",
		Size:     6,
		Tool:     "pen",
	}
	room.HandleStroke(first.Id, stroke)
	room.HandleStroke(second.Id, StrokeData{
		StrokeId: "gesture-2",
		Points:   []Point{{X: 5, Y: 6}, {X: 7, Y: 8}},
		Color:    "#2b6cb0",
		Size:     10,
		Tool:     "marker",
	})

	room.mu.RLock()
	strokeCount := len(room.Strokes)
	room.mu.RUnlock()
	if strokeCount != 2 {
		t.Fatalf("expected 2 strokes, got %d", strokeCount)
	}

	room.HandleUndo(first.Id)
	room.mu.RLock()
	strokeCount = len(room.Strokes)
	remainingOwner := room.Strokes[0].OwnerId
	room.mu.RUnlock()
	if strokeCount != 1 || remainingOwner != second.Id {
		t.Fatalf("undo removed the wrong stroke: count=%d owner=%s", strokeCount, remainingOwner)
	}
}

func TestUndoRemovesAllChunksOfAGesture(t *testing.T) {
	room := testRoom(modeTogether, time.Minute)
	first := testClient("first", "Alice")
	second := testClient("second", "Bob")
	room.AddClient(first)
	room.AddClient(second)

	for i := 0; i < 3; i++ {
		room.HandleStroke(first.Id, StrokeData{
			StrokeId: "same-gesture",
			Points:   []Point{{X: float64(i), Y: float64(i)}, {X: float64(i + 1), Y: float64(i + 1)}},
			Color:    "#162521",
			Size:     6,
		})
	}
	room.HandleStroke(first.Id, StrokeData{
		StrokeId: "new-gesture",
		Points:   []Point{{X: 9, Y: 9}, {X: 10, Y: 10}},
		Color:    "#162521",
		Size:     6,
	})
	room.HandleUndo(first.Id)

	room.mu.RLock()
	defer room.mu.RUnlock()
	if len(room.Strokes) != 3 {
		t.Fatalf("expected 3 remaining chunks, got %d", len(room.Strokes))
	}
	for _, stroke := range room.Strokes {
		if stroke.StrokeId != "same-gesture" {
			t.Fatalf("unexpected remaining gesture %q", stroke.StrokeId)
		}
	}
}

func TestSnapshotGivesMidGameJoinerMaskedGameAndStrokes(t *testing.T) {
	room := testRoom(modeClassic, time.Minute)
	first := testClient("first", "Alice")
	second := testClient("second", "Bob")
	room.AddClient(first)
	room.AddClient(second)
	room.StartGame()
	readMessage(t, first)
	readMessage(t, second)

	room.mu.RLock()
	drawerId := room.CurrentDrawer
	room.mu.RUnlock()

	room.HandleStroke(drawerId, StrokeData{
		StrokeId: "g1",
		Points:   []Point{{X: 2, Y: 3}, {X: 4, Y: 5}},
		Color:    "#162521",
		Size:     6,
	})

	lateClient := testClient("late", "Cara")
	room.AddClient(lateClient)
	snapshot := room.Snapshot(lateClient.Id)

	if snapshot.Status != statePlaying {
		t.Fatalf("expected playing status, got %s", snapshot.Status)
	}
	if snapshot.Game == nil || snapshot.Game.Word != "" {
		t.Fatal("expected masked game data for late joiner")
	}
	if len(snapshot.Strokes) != 1 {
		t.Fatalf("expected stroke history, got %d", len(snapshot.Strokes))
	}
}

func TestCreateRoomGeneratesUniqueCodes(t *testing.T) {
	hub := NewHub(DefaultConfig())
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, err := hub.CreateRoom(RoomOptions{Mode: modeClassic, RoundDuration: time.Minute})
		if err != nil {
			t.Fatal(err)
		}
		if seen[code] {
			t.Fatalf("duplicate room code generated: %s", code)
		}
		seen[code] = true
	}
}
