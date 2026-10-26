package main

import (
	"testing"
	"time"
)

func togetherRoom(activity string) *Room {
	room := NewRoom("ROOMTEST", RoomOptions{
		Mode:              modeTogether,
		TogetherActivity:  activity,
		RelayTurnDuration: time.Minute,
		MaxPlayers:        20,
	})
	return room
}

func addTestClients(room *Room, clients ...*Client) {
	for _, client := range clients {
		room.AddClient(client)
	}
}

func TestMissionDeckIsValid(t *testing.T) {
	if len(missionDeck) < 40 {
		t.Fatalf("expected at least 40 missions, got %d", len(missionDeck))
	}
	seen := make(map[string]bool)
	for _, mission := range missionDeck {
		if mission.Id == "" || mission.Title == "" || mission.Brief == "" {
			t.Fatalf("mission is missing required text: %+v", mission)
		}
		if seen[mission.Id] {
			t.Fatalf("duplicate mission id %q", mission.Id)
		}
		seen[mission.Id] = true
		if len(mission.Tasks) < 4 || len(mission.Tasks) > 6 {
			t.Fatalf("mission %q has %d tasks, want 4-6", mission.Id, len(mission.Tasks))
		}
		for _, task := range mission.Tasks {
			if task == "" {
				t.Fatalf("mission %q has an empty task", mission.Id)
			}
		}
	}
}

func TestMissionToggleAndAdvance(t *testing.T) {
	room := togetherRoom(activityMissions)
	t.Cleanup(room.Close)
	first := testClient("first", "Alice")
	second := testClient("second", "Bob")
	addTestClients(room, first, second)

	state := room.Snapshot(first.Id).Mission
	if state == nil {
		t.Fatal("expected mission state in snapshot")
	}
	if state.Number != 1 || state.Completed != 0 {
		t.Fatalf("unexpected initial mission state: %+v", state)
	}

	room.HandleMissionToggle(0)
	room.HandleNextMission()
	state = room.Snapshot(first.Id).Mission
	if state.Completed != 0 || state.Number != 1 {
		t.Fatalf("mission advanced before all tasks were checked: %+v", state)
	}

	room.HandleMissionToggle(0)
	for index := range state.Mission.Tasks {
		room.HandleMissionToggle(index)
	}
	room.HandleStroke(first.Id, StrokeData{
		StrokeId: "mission-doodle",
		Points:   []Point{{X: 1, Y: 2}, {X: 3, Y: 4}},
		Color:    "#162521",
		Size:     6,
	})
	room.HandleNextMission()

	state = room.Snapshot(first.Id).Mission
	if state.Completed != 1 || state.Number != 2 {
		t.Fatalf("unexpected advanced mission state: %+v", state)
	}
	for _, checked := range state.Checked {
		if checked {
			t.Fatal("new mission should start with an unchecked checklist")
		}
	}
	if strokes := room.Snapshot(first.Id).Strokes; len(strokes) != 0 {
		t.Fatalf("expected mission advance to clear the canvas, got %d strokes", len(strokes))
	}
}

func TestRelayStartRequiresTwoPlayers(t *testing.T) {
	room := togetherRoom(activityRelay)
	t.Cleanup(room.Close)
	addTestClients(room, testClient("first", "Alice"))

	room.StartRelay()
	state := room.Snapshot("first").Relay
	if state == nil || state.Status != relayStatusLobby {
		t.Fatalf("relay should stay in lobby with one player: %+v", state)
	}
}

func TestRelayStartsWithScheduleAndSnapshot(t *testing.T) {
	room := togetherRoom(activityRelay)
	t.Cleanup(room.Close)
	first := testClient("first", "Alice")
	second := testClient("second", "Bob")
	third := testClient("third", "Cara")
	addTestClients(room, first, second, third)

	if room.MaxPlayers != maxRelayPlayers {
		t.Fatalf("relay max players = %d, want %d", room.MaxPlayers, maxRelayPlayers)
	}
	room.StartRelay()

	snapshot := room.Snapshot(first.Id)
	if snapshot.Relay == nil {
		t.Fatal("expected relay state in snapshot")
	}
	if snapshot.Relay.Status != relayStatusPlaying {
		t.Fatalf("expected playing relay, got %s", snapshot.Relay.Status)
	}
	if snapshot.Relay.Prompt == "" || snapshot.Relay.DrawerId == "" {
		t.Fatalf("expected prompt and active drawer: %+v", snapshot.Relay)
	}
	if snapshot.Relay.TotalTurns != 6 || snapshot.Relay.Turn != 1 {
		t.Fatalf("unexpected relay turn schedule: %+v", snapshot.Relay)
	}
	if len(snapshot.Strokes) != 0 {
		t.Fatal("expected relay start to clear the canvas")
	}
}

func TestRelayOnlyActivePlayerCanDrawAndClearIsDisabled(t *testing.T) {
	room := togetherRoom(activityRelay)
	t.Cleanup(room.Close)
	first := testClient("first", "Alice")
	second := testClient("second", "Bob")
	addTestClients(room, first, second)
	room.StartRelay()

	state := room.Snapshot(first.Id).Relay
	activeID := state.DrawerId
	otherID := first.Id
	if activeID == first.Id {
		otherID = second.Id
	}
	stroke := StrokeData{
		StrokeId: "relay-stroke",
		Points:   []Point{{X: 1, Y: 2}, {X: 3, Y: 4}},
		Color:    "#162521",
		Size:     6,
	}
	room.HandleStroke(otherID, stroke)
	if strokes := room.Snapshot(first.Id).Strokes; len(strokes) != 0 {
		t.Fatal("non-active relay player should not be able to draw")
	}

	room.HandleStroke(activeID, stroke)
	if strokes := room.Snapshot(first.Id).Strokes; len(strokes) != 1 {
		t.Fatalf("active relay player should be able to draw, got %d strokes", len(strokes))
	}

	room.HandleClear(activeID)
	if strokes := room.Snapshot(first.Id).Strokes; len(strokes) != 1 {
		t.Fatal("clear should be disabled during a relay")
	}
}

func TestRelayClearsCanvasBetweenTurns(t *testing.T) {
	room := togetherRoom(activityRelay)
	t.Cleanup(room.Close)
	addTestClients(room, testClient("first", "Alice"), testClient("second", "Bob"))
	room.StartRelay()

	firstState := room.Snapshot("first").Relay
	room.HandleStroke(firstState.DrawerId, StrokeData{
		StrokeId: "relay-drawing",
		Points:   []Point{{X: 2, Y: 3}, {X: 4, Y: 5}},
		Color:    "#162521",
		Size:     6,
	})
	if strokes := room.Snapshot("first").Strokes; len(strokes) != 1 {
		t.Fatalf("expected active drawer stroke, got %d strokes", len(strokes))
	}

	room.handleRelayTimeout()
	state := room.Snapshot("first").Relay
	if state.Status != relayStatusPlaying || state.Turn != 2 {
		t.Fatalf("expected relay to advance to turn 2, got %+v", state)
	}
	if strokes := room.Snapshot("first").Strokes; len(strokes) != 0 {
		t.Fatalf("expected next relay turn to clear the canvas, got %d strokes", len(strokes))
	}
}

func TestRelayAdvancesTurnsAndFinishes(t *testing.T) {
	room := togetherRoom(activityRelay)
	t.Cleanup(room.Close)
	addTestClients(room, testClient("first", "Alice"), testClient("second", "Bob"))
	room.StartRelay()

	totalTurns := room.Snapshot("first").Relay.TotalTurns
	for i := 0; i < totalTurns; i++ {
		room.handleRelayTimeout()
	}

	state := room.Snapshot("first").Relay
	if state.Status != relayStatusFinished {
		t.Fatalf("expected finished relay, got %s", state.Status)
	}
	if state.DrawerId != "" || state.Prompt != "" {
		t.Fatalf("finished relay should not have an active drawer or prompt: %+v", state)
	}
}

func TestRelayLateJoinerReceivesCurrentTurn(t *testing.T) {
	room := togetherRoom(activityRelay)
	t.Cleanup(room.Close)
	first := testClient("first", "Alice")
	second := testClient("second", "Bob")
	addTestClients(room, first, second)
	room.StartRelay()

	drawerID := room.Snapshot(first.Id).Relay.DrawerId
	room.HandleStroke(drawerID, StrokeData{
		StrokeId: "relay-drawing",
		Points:   []Point{{X: 2, Y: 3}, {X: 4, Y: 5}},
		Color:    "#162521",
		Size:     6,
	})

	late := testClient("late", "Drew")
	room.AddClient(late)
	snapshot := room.Snapshot(late.Id)
	if snapshot.Relay == nil || snapshot.Relay.Status != relayStatusPlaying {
		t.Fatalf("late joiner should receive the active relay: %+v", snapshot.Relay)
	}
	if snapshot.Relay.Prompt == "" || snapshot.Relay.DrawerId == "" {
		t.Fatalf("late joiner is missing relay turn details: %+v", snapshot.Relay)
	}
	if len(snapshot.Strokes) != 1 {
		t.Fatalf("late joiner should receive canvas history, got %d strokes", len(snapshot.Strokes))
	}
}

func TestRelayPausesAndResumesAfterPlayerChange(t *testing.T) {
	room := togetherRoom(activityRelay)
	t.Cleanup(room.Close)
	first := testClient("first", "Alice")
	second := testClient("second", "Bob")
	addTestClients(room, first, second)
	room.StartRelay()

	room.RemoveClient(first.Id)
	state := room.Snapshot(second.Id).Relay
	if state.Status != relayStatusPaused {
		t.Fatalf("relay should pause with one player left, got %s", state.Status)
	}

	third := testClient("third", "Cara")
	room.AddClient(third)
	room.TryResumeRelay()
	state = room.Snapshot(second.Id).Relay
	if state.Status != relayStatusPlaying {
		t.Fatalf("relay should resume with two players, got %s", state.Status)
	}
	if state.TotalTurns != 4 {
		t.Fatalf("resumed relay should include both players twice, got %d turns", state.TotalTurns)
	}
}
