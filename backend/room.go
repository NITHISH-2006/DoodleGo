package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	modeClassic  = "classic"
	modeTogether = "together"

	stateLobby    = "lobby"
	statePlaying  = "playing"
	stateTogether = "together"
	stateClosed   = "closed"

	maxStrokesPerRoom  = 20000
	maxPointsPerStroke = 4096
	maxGuessLength     = 80
)

type Room struct {
	Code string

	mu      sync.RWMutex
	Clients map[string]*Client

	State string
	Mode  string

	HostId        string
	CurrentDrawer string
	CurrentWord   string
	RoundNum      int

	Scores          map[string]int
	CorrectGuessers map[string]bool
	Strokes         []StrokeData

	RoundDuration time.Duration
	MaxPlayers    int

	roundTimer   *time.Timer
	restartTimer *time.Timer
	roundEndsAt  time.Time
	lastDrawerId string
	lastWord     string
}

type RoomSnapshot struct {
	Players       []PlayerInfo
	Status        string
	Mode          string
	RoundDuration int
	MaxPlayers    int
	Game          *GameStartedData
	Strokes       []StrokeData
}

func NewRoom(code string, options RoomOptions) *Room {
	state := stateLobby
	if options.Mode == modeTogether {
		state = stateTogether
	}
	maxPlayers := options.MaxPlayers
	if maxPlayers <= 1 {
		maxPlayers = 20
	}
	return &Room{
		Code:            code,
		Clients:         make(map[string]*Client),
		State:           state,
		Mode:            options.Mode,
		Scores:          make(map[string]int),
		CorrectGuessers: make(map[string]bool),
		RoundDuration:   options.RoundDuration,
		MaxPlayers:      maxPlayers,
	}
}

func (r *Room) AddClient(c *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State == stateClosed {
		return errRoomClosed
	}
	if len(r.Clients) >= r.MaxPlayers {
		return errRoomFull
	}
	if _, exists := r.Clients[c.Id]; exists {
		return nil
	}

	r.Clients[c.Id] = c
	if _, exists := r.Scores[c.Id]; !exists {
		r.Scores[c.Id] = 0
	}
	if r.HostId == "" {
		r.HostId = c.Id
	}
	return nil
}

// RemoveClient removes a client and reports whether it was actually present.
func (r *Room) RemoveClient(clientId string) bool {
	r.mu.Lock()
	if _, exists := r.Clients[clientId]; !exists {
		r.mu.Unlock()
		return false
	}

	delete(r.Clients, clientId)
	delete(r.Scores, clientId)
	delete(r.CorrectGuessers, clientId)

	wasDrawer := r.State == statePlaying && r.CurrentDrawer == clientId
	if r.HostId == clientId {
		r.HostId = ""
		for id := range r.Clients {
			r.HostId = id
			break
		}
	}
	empty := len(r.Clients) == 0
	r.mu.Unlock()

	if wasDrawer && !empty {
		r.endRound()
	}
	return true
}

func (r *Room) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Clients)
}

func (r *Room) Close() {
	r.mu.Lock()
	r.State = stateClosed
	if r.roundTimer != nil {
		r.roundTimer.Stop()
		r.roundTimer = nil
	}
	if r.restartTimer != nil {
		r.restartTimer.Stop()
		r.restartTimer = nil
	}
	r.Clients = make(map[string]*Client)
	r.Scores = make(map[string]int)
	r.CorrectGuessers = make(map[string]bool)
	r.Strokes = nil
	r.mu.Unlock()
}

func (r *Room) Broadcast(msg *Message) {
	r.BroadcastExcept("", msg)
}

func (r *Room) BroadcastExcept(excludeClientId string, msg *Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for id, client := range r.Clients {
		if id == excludeClientId {
			continue
		}
		client.trySend(msg)
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
	sort.Slice(players, func(i, j int) bool {
		if players[i].Score == players[j].Score {
			return players[i].Name < players[j].Name
		}
		return players[i].Score > players[j].Score
	})
	return players
}

func (r *Room) Snapshot(clientId string) RoomSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshot := RoomSnapshot{
		Players:       r.playersLocked(),
		Status:        r.State,
		Mode:          r.Mode,
		RoundDuration: int(r.RoundDuration / time.Second),
		MaxPlayers:    r.MaxPlayers,
	}
	if r.State == statePlaying {
		snapshot.Game = r.gameStartedDataLocked(clientId)
	}
	if r.State == statePlaying || r.State == stateTogether {
		snapshot.Strokes = append([]StrokeData(nil), r.Strokes...)
	}
	return snapshot
}

func (r *Room) StartGame() {
	r.mu.Lock()
	if r.Mode != modeClassic || r.State == statePlaying || r.State == stateClosed || len(r.Clients) < 2 {
		r.mu.Unlock()
		return
	}
	if r.restartTimer != nil {
		r.restartTimer.Stop()
		r.restartTimer = nil
	}

	drawerId := r.pickDrawerLocked()
	word := r.pickWordLocked()

	r.State = statePlaying
	r.CurrentDrawer = drawerId
	r.CurrentWord = word
	r.RoundNum++
	r.CorrectGuessers = make(map[string]bool)
	r.Strokes = nil
	r.roundEndsAt = time.Now().Add(r.effectiveRoundDurationLocked())
	r.startRoundTimerLocked()

	type target struct {
		client *Client
		msg    *Message
	}
	targets := make([]target, 0, len(r.Clients))
	for _, client := range r.Clients {
		data := r.gameStartedDataLocked(client.Id)
		targets = append(targets, target{
			client: client,
			msg:    newMessage(MsgGameStarted, data),
		})
	}
	roundNum := r.RoundNum
	r.mu.Unlock()

	for _, t := range targets {
		t.client.trySend(t.msg)
	}
	log.Printf("Room %s round %d started. Drawer: %s, Word: %s", r.Code, roundNum, drawerId, word)
}

func (r *Room) effectiveRoundDurationLocked() time.Duration {
	if r.RoundDuration <= 0 {
		return 90 * time.Second
	}
	return r.RoundDuration
}

func (r *Room) startRoundTimerLocked() {
	if r.roundTimer != nil {
		r.roundTimer.Stop()
	}
	r.roundTimer = time.AfterFunc(r.effectiveRoundDurationLocked(), r.endRound)
}

func (r *Room) pickDrawerLocked() string {
	ids := make([]string, 0, len(r.Clients))
	for id := range r.Clients {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	if len(ids) == 0 {
		return ""
	}

	index := 0
	if r.lastDrawerId != "" {
		for i, id := range ids {
			if id == r.lastDrawerId {
				index = (i + 1) % len(ids)
				break
			}
		}
	}
	drawer := ids[index]
	r.lastDrawerId = drawer
	return drawer
}

func (r *Room) pickWordLocked() string {
	if len(wordList) == 0 {
		return "apple"
	}
	for attempt := 0; attempt < 10; attempt++ {
		word := wordList[rand.Intn(len(wordList))]
		if word != r.lastWord || len(wordList) == 1 {
			r.lastWord = word
			return word
		}
	}
	return wordList[rand.Intn(len(wordList))]
}

func (r *Room) endRound() {
	r.mu.Lock()
	if r.State != statePlaying {
		r.mu.Unlock()
		return
	}

	if r.roundTimer != nil {
		r.roundTimer.Stop()
		r.roundTimer = nil
	}
	r.State = stateLobby
	r.CurrentDrawer = ""
	word := r.CurrentWord
	r.CurrentWord = ""
	r.CorrectGuessers = make(map[string]bool)
	players := r.playersLocked()
	shouldRestart := r.Mode == modeClassic && len(r.Clients) >= 2 && r.State != stateClosed
	r.mu.Unlock()

	r.Broadcast(newMessage(MsgRoundEnd, RoundEndData{Word: word, Players: players}))

	if shouldRestart {
		r.mu.Lock()
		if r.State == stateLobby && r.Mode == modeClassic && len(r.Clients) >= 2 && r.State != stateClosed {
			if r.restartTimer != nil {
				r.restartTimer.Stop()
			}
			r.restartTimer = time.AfterFunc(5*time.Second, r.StartGame)
		}
		r.mu.Unlock()
	}
}

func (r *Room) gameStartedDataLocked(clientId string) *GameStartedData {
	if r.State != statePlaying {
		return nil
	}
	wordLength := len([]rune(r.CurrentWord))
	word := ""
	if clientId == r.CurrentDrawer {
		word = r.CurrentWord
	}
	timeLeft := int(math.Ceil(time.Until(r.roundEndsAt).Seconds()))
	if timeLeft < 0 {
		timeLeft = 0
	}
	drawerName := ""
	if drawer, exists := r.Clients[r.CurrentDrawer]; exists {
		drawerName = drawer.Name
	}
	return &GameStartedData{
		DrawerId:    r.CurrentDrawer,
		DrawerName:  drawerName,
		Word:        word,
		WordLength:  wordLength,
		MaskedWord:  strings.Repeat("_", wordLength),
		RoundNum:    r.RoundNum,
		TimeLeft:    timeLeft,
		RoundEndsAt: r.roundEndsAt.UnixMilli(),
	}
}

func (r *Room) HandleGuess(clientId, text string) {
	r.mu.Lock()

	if r.Mode != modeClassic || r.State != statePlaying {
		r.mu.Unlock()
		return
	}
	if clientId == r.CurrentDrawer {
		r.mu.Unlock()
		return
	}
	client, exists := r.Clients[clientId]
	if !exists {
		r.mu.Unlock()
		return
	}
	// A player can only score once per round.
	if r.CorrectGuessers[clientId] {
		r.mu.Unlock()
		return
	}

	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		r.mu.Unlock()
		return
	}
	if runes := []rune(trimmed); len(runes) > maxGuessLength {
		trimmed = string(runes[:maxGuessLength])
	}

	isCorrect := strings.EqualFold(trimmed, r.CurrentWord)
	displayText := trimmed
	messageText := fmt.Sprintf("%s guessed '%s'", client.Name, trimmed)
	if isCorrect {
		r.Scores[clientId] += 10
		r.CorrectGuessers[clientId] = true
		// Never broadcast the secret word after a correct guess.
		displayText = ""
		messageText = fmt.Sprintf("%s guessed correctly!", client.Name)
	}

	msg := newMessage(MsgGuessReceived, GuessReceivedData{
		GuesserId:   clientId,
		GuesserName: client.Name,
		Text:        displayText,
		IsCorrect:   isCorrect,
		Message:     messageText,
	})

	var scoreMsg *Message
	if isCorrect {
		scoreMsg = newMessage(MsgScoreUpdate, PlayerScoreUpdate{
			PlayerId: clientId,
			Score:    r.Scores[clientId],
		})
	}
	r.mu.Unlock()

	r.Broadcast(msg)
	if scoreMsg != nil {
		r.Broadcast(scoreMsg)
	}
}

// HandleStroke validates and stores a stroke chunk, then relays it.
func (r *Room) HandleStroke(clientId string, stroke StrokeData) {
	r.mu.Lock()
	if !r.canDrawLocked(clientId) {
		r.mu.Unlock()
		return
	}
	if len(stroke.Points) == 0 || len(stroke.Points) > maxPointsPerStroke {
		r.mu.Unlock()
		return
	}

	stroke.OwnerId = clientId
	if stroke.StrokeId == "" {
		stroke.StrokeId = randomID("s-")
	}
	stroke.Tool = normalizeTool(stroke.Tool)
	stroke.Color = normalizeColor(stroke.Color)
	stroke.Size = clampInt(stroke.Size, 1, 64)
	stroke.Opacity = normalizeOpacity(stroke.Opacity, stroke.Tool)

	if len(r.Strokes) >= maxStrokesPerRoom {
		r.Strokes = append([]StrokeData(nil), r.Strokes[len(r.Strokes)/2:]...)
	}
	r.Strokes = append(r.Strokes, stroke)
	r.mu.Unlock()

	msg := newMessage(MsgStroke, stroke)
	msg.PlayerId = clientId
	r.Broadcast(msg)
}

// HandleUndo removes the requesting player's most recent gesture.
func (r *Room) HandleUndo(clientId string) {
	r.mu.Lock()
	if !r.canDrawLocked(clientId) {
		r.mu.Unlock()
		return
	}

	targetIndex := -1
	targetStrokeId := ""
	for i := len(r.Strokes) - 1; i >= 0; i-- {
		if r.Strokes[i].OwnerId == clientId {
			targetIndex = i
			targetStrokeId = r.Strokes[i].StrokeId
			break
		}
	}
	if targetIndex < 0 {
		r.mu.Unlock()
		return
	}

	newStrokes := make([]StrokeData, 0, len(r.Strokes))
	for i, stroke := range r.Strokes {
		if targetStrokeId != "" && stroke.StrokeId == targetStrokeId {
			continue
		}
		if targetStrokeId == "" && i == targetIndex {
			continue
		}
		newStrokes = append(newStrokes, stroke)
	}
	r.Strokes = newStrokes
	snapshot := make([]StrokeData, 0, len(r.Strokes))
	snapshot = append(snapshot, r.Strokes...)
	r.mu.Unlock()

	r.Broadcast(newMessage(MsgCanvasState, CanvasStateData{Strokes: snapshot}))
}

func (r *Room) HandleClear(clientId string) {
	r.mu.Lock()
	if !r.canDrawLocked(clientId) {
		r.mu.Unlock()
		return
	}
	r.Strokes = nil
	r.mu.Unlock()

	r.Broadcast(newMessage(MsgCanvasState, CanvasStateData{Strokes: []StrokeData{}}))
}

func (r *Room) canDrawLocked(clientId string) bool {
	switch r.Mode {
	case modeTogether:
		return r.State == stateTogether
	case modeClassic:
		return r.State == statePlaying && r.CurrentDrawer == clientId
	default:
		return false
	}
}

func normalizeTool(tool string) string {
	switch strings.ToLower(strings.TrimSpace(tool)) {
	case "marker", "eraser":
		return strings.ToLower(strings.TrimSpace(tool))
	default:
		return "pen"
	}
}

func normalizeColor(color string) string {
	color = strings.TrimSpace(color)
	if len(color) != 7 || color[0] != '#' {
		return "#162521"
	}
	for _, ch := range color[1:] {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
			return "#162521"
		}
	}
	return color
}

func normalizeOpacity(opacity float64, tool string) float64 {
	if opacity <= 0 {
		if tool == "marker" {
			return 0.35
		}
		return 1
	}
	if opacity > 1 {
		return 1
	}
	return opacity
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
