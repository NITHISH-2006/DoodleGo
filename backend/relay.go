package main

import (
	"math"
	"math/rand"
	"sort"
	"time"
)

const (
	relayStatusLobby    = "lobby"
	relayStatusPlaying  = "playing"
	relayStatusPaused   = "paused"
	relayStatusFinished = "finished"

	relayTurnsPerPlayer = 2
)

var relayPrompts = []string{
	"a cat riding a bicycle",
	"a dog baking cookies",
	"a rocket landing on a pizza",
	"an astronaut walking a dinosaur",
	"a snowman at the beach",
	"a robot watering flowers",
	"a shark wearing sunglasses",
	"a pirate searching for treasure",
	"a dragon reading a book",
	"a penguin on a skateboard",
	"a monkey painting a rainbow",
	"a turtle crossing a bridge",
	"a whale spouting confetti",
	"a bear camping under stars",
	"a fish in a spacesuit",
	"a superhero rescuing a cat",
	"a giant cupcake in a city",
	"a fox flying a kite",
	"a jellyfish playing a guitar",
	"a dinosaur at a birthday party",
	"a llama in a hot air balloon",
	"an octopus juggling fruit",
	"a snail racing a car",
	"a ghost drinking coffee",
	"a frog on a throne",
	"a squirrel storing acorns",
	"a cat sleeping on a laptop",
	"a giraffe wearing a scarf",
	"a mouse building a robot",
	"a panda eating spaghetti",
	"an owl delivering mail",
	"a crab playing beach volleyball",
	"a rabbit pulling a carrot",
	"a lion getting a haircut",
	"a bee carrying a picnic basket",
	"a unicorn jumping over a rainbow",
	"a crocodile brushing its teeth",
	"a hedgehog with an umbrella",
	"a robot walking a dog",
	"a flamingo dancing in the rain",
}

func normalizeRelayTurnDuration(duration time.Duration) time.Duration {
	switch int(duration / time.Second) {
	case 30, 60, 90:
		return duration
	default:
		return 60 * time.Second
	}
}

func newRelayDeck(lastPrompt string) []string {
	deck := append([]string(nil), relayPrompts...)
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
	if len(deck) > 1 && deck[0] == lastPrompt {
		deck[0], deck[1] = deck[1], deck[0]
	}
	return deck
}

func (r *Room) relayStateLocked() *RelayStateData {
	status := r.relayStatus
	if status == "" {
		status = relayStatusLobby
	}
	timeLeft := 0
	if status == relayStatusPlaying {
		timeLeft = int(math.Ceil(time.Until(r.relayEndsAt).Seconds()))
		if timeLeft < 0 {
			timeLeft = 0
		}
	}
	drawerName := ""
	if drawer, exists := r.Clients[r.CurrentDrawer]; exists {
		drawerName = drawer.Name
	}
	turn := r.relayTurn
	if status == relayStatusLobby {
		turn = 0
	}
	var endsAt int64
	if !r.relayEndsAt.IsZero() {
		endsAt = r.relayEndsAt.UnixMilli()
	}
	return &RelayStateData{
		Status:         status,
		Prompt:         r.relayPrompt,
		DrawerId:       r.CurrentDrawer,
		DrawerName:     drawerName,
		Turn:           turn,
		TotalTurns:     r.relayTotalTurns,
		TimeLeft:       timeLeft,
		TurnEndsAt:     endsAt,
		TurnDuration:   int(r.RelayTurnDuration / time.Second),
		TurnsPerPlayer: relayTurnsPerPlayer,
	}
}

func (r *Room) relayPlayerIDsLocked() []string {
	ids := make([]string, 0, len(r.Clients))
	for id := range r.Clients {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		left := r.Clients[ids[i]]
		right := r.Clients[ids[j]]
		if left.Name == right.Name {
			return left.Id < right.Id
		}
		return left.Name < right.Name
	})
	return ids
}

func (r *Room) buildRelayScheduleLocked() {
	ids := r.relayPlayerIDsLocked()
	schedule := make([]string, 0, len(ids)*relayTurnsPerPlayer)
	for round := 0; round < relayTurnsPerPlayer; round++ {
		roundIDs := append([]string(nil), ids...)
		if round%2 == 1 {
			for left, right := 0, len(roundIDs)-1; left < right; left, right = left+1, right-1 {
				roundIDs[left], roundIDs[right] = roundIDs[right], roundIDs[left]
			}
		}
		schedule = append(schedule, roundIDs...)
	}
	r.relayOrder = schedule
	r.relayTotalTurns = len(schedule)
	r.relayTurnIndex = 0
}

func (r *Room) StartRelay() {
	r.mu.Lock()
	if r.Mode != modeTogether || r.TogetherActivity != activityRelay || len(r.Clients) < 2 {
		r.mu.Unlock()
		return
	}
	if r.relayStatus == relayStatusPlaying {
		r.mu.Unlock()
		return
	}
	if r.relayTimer != nil {
		r.relayTimer.Stop()
		r.relayTimer = nil
	}
	r.relayDeck = newRelayDeck(r.relayLastPrompt)
	r.relayDeckIndex = 0
	r.buildRelayScheduleLocked()
	r.Strokes = nil
	r.startRelayTurnLocked()
	state := r.relayStateLocked()
	r.mu.Unlock()

	r.Broadcast(newMessage(MsgRelayState, state))
	r.Broadcast(newMessage(MsgCanvasState, CanvasStateData{Strokes: []StrokeData{}}))
}

func (r *Room) startRelayTurnLocked() {
	if r.relayTimer != nil {
		r.relayTimer.Stop()
		r.relayTimer = nil
	}
	if len(r.Clients) < 2 {
		r.relayStatus = relayStatusPaused
		r.CurrentDrawer = ""
		r.relayPrompt = ""
		r.relayEndsAt = time.Time{}
		return
	}

	for r.relayTurnIndex < len(r.relayOrder) {
		drawerID := r.relayOrder[r.relayTurnIndex]
		if _, exists := r.Clients[drawerID]; exists {
			break
		}
		r.relayOrder = append(r.relayOrder[:r.relayTurnIndex], r.relayOrder[r.relayTurnIndex+1:]...)
	}
	if r.relayTurnIndex >= len(r.relayOrder) {
		r.finishRelayLocked()
		return
	}

	r.relayTotalTurns = len(r.relayOrder)
	r.CurrentDrawer = r.relayOrder[r.relayTurnIndex]
	r.relayTurn = r.relayTurnIndex + 1
	r.relayPrompt = r.pickRelayPromptLocked()
	r.relayStatus = relayStatusPlaying
	r.relayEndsAt = time.Now().Add(r.RelayTurnDuration)
	r.relayTimer = time.AfterFunc(r.RelayTurnDuration, r.handleRelayTimeout)
}

func (r *Room) pickRelayPromptLocked() string {
	if len(r.relayDeck) == 0 || r.relayDeckIndex >= len(r.relayDeck) {
		r.relayDeck = newRelayDeck(r.relayLastPrompt)
		r.relayDeckIndex = 0
	}
	if len(r.relayDeck) == 0 {
		return "a mystery doodle"
	}
	prompt := r.relayDeck[r.relayDeckIndex]
	r.relayDeckIndex++
	r.relayLastPrompt = prompt
	return prompt
}

func (r *Room) finishRelayLocked() {
	if r.relayTimer != nil {
		r.relayTimer.Stop()
		r.relayTimer = nil
	}
	r.relayStatus = relayStatusFinished
	r.CurrentDrawer = ""
	r.relayPrompt = ""
	r.relayEndsAt = time.Time{}
	if r.relayTotalTurns == 0 {
		r.relayTotalTurns = r.relayTurn
	}
}

func (r *Room) handleRelayTimeout() {
	r.mu.Lock()
	if r.relayStatus != relayStatusPlaying {
		r.mu.Unlock()
		return
	}
	r.relayTurnIndex++
	if r.relayTurnIndex >= len(r.relayOrder) {
		r.finishRelayLocked()
	} else {
		r.Strokes = nil
		r.startRelayTurnLocked()
	}
	state := r.relayStateLocked()
	strokes := append([]StrokeData(nil), r.Strokes...)
	r.mu.Unlock()

	r.Broadcast(newMessage(MsgRelayState, state))
	r.Broadcast(newMessage(MsgCanvasState, CanvasStateData{Strokes: strokes}))
}

func (r *Room) handleRelayPlayerLeft(clientID string) {
	r.mu.Lock()
	if r.Mode != modeTogether || r.TogetherActivity != activityRelay {
		r.mu.Unlock()
		return
	}

	wasScheduled := false
	for _, id := range r.relayOrder {
		if id == clientID {
			wasScheduled = true
			break
		}
	}
	if !wasScheduled {
		r.mu.Unlock()
		return
	}

	currentID := r.CurrentDrawer
	oldIndex := r.relayTurnIndex
	oldOrder := r.relayOrder
	newOrder := make([]string, 0, len(oldOrder))
	removedBefore := 0
	for index, id := range oldOrder {
		if id == clientID {
			if index < oldIndex {
				removedBefore++
			}
			continue
		}
		newOrder = append(newOrder, id)
	}
	newIndex := oldIndex - removedBefore
	if newIndex < 0 {
		newIndex = 0
	}
	r.relayOrder = newOrder
	r.relayTotalTurns = len(newOrder)

	cleared := false
	switch {
	case len(r.Clients) < 2:
		if r.relayTimer != nil {
			r.relayTimer.Stop()
			r.relayTimer = nil
		}
		r.relayStatus = relayStatusPaused
		r.CurrentDrawer = ""
		r.relayPrompt = ""
		r.relayEndsAt = time.Time{}
	case r.relayStatus == relayStatusPlaying && currentID == clientID:
		r.relayTurnIndex = newIndex
		if newIndex >= len(r.relayOrder) {
			r.finishRelayLocked()
		} else {
			r.startRelayTurnLocked()
			cleared = true
		}
	case r.relayStatus == relayStatusPlaying:
		found := -1
		for index := newIndex; index < len(r.relayOrder); index++ {
			if r.relayOrder[index] == currentID {
				found = index
				break
			}
		}
		if found < 0 {
			for index := 0; index < len(r.relayOrder); index++ {
				if r.relayOrder[index] == currentID {
					found = index
					break
				}
			}
		}
		if found >= 0 {
			r.relayTurnIndex = found
		} else {
			if r.relayTimer != nil {
				r.relayTimer.Stop()
				r.relayTimer = nil
			}
			r.relayStatus = relayStatusPaused
			r.CurrentDrawer = ""
			r.relayPrompt = ""
			r.relayEndsAt = time.Time{}
		}
	}

	state := r.relayStateLocked()
	var strokes []StrokeData
	if cleared {
		strokes = []StrokeData{}
	} else {
		strokes = append([]StrokeData(nil), r.Strokes...)
	}
	r.mu.Unlock()

	r.Broadcast(newMessage(MsgRelayState, state))
	if cleared {
		r.Broadcast(newMessage(MsgCanvasState, CanvasStateData{Strokes: strokes}))
	}
}

func (r *Room) TryResumeRelay() {
	r.mu.Lock()
	if r.Mode != modeTogether || r.TogetherActivity != activityRelay || r.relayStatus != relayStatusPaused || len(r.Clients) < 2 {
		r.mu.Unlock()
		return
	}
	r.buildRelayScheduleLocked()
	r.Strokes = nil
	r.startRelayTurnLocked()
	state := r.relayStateLocked()
	r.mu.Unlock()

	r.Broadcast(newMessage(MsgRelayState, state))
	r.Broadcast(newMessage(MsgCanvasState, CanvasStateData{Strokes: []StrokeData{}}))
}
