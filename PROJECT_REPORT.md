# DoodleGo (GoSkribbl) — Deep Architectural & Technical Report

## 1. Project Overview

**DoodleGo** (also referenced as **GoSkribbl**) is a real-time multiplayer drawing and word-guessing web application inspired by Skribbl.io. The project features a concurrent backend server written in Go using WebSockets for low-latency bidirectional messaging, paired with a React + Vite frontend leveraging HTML5 Canvas for real-time stroke broadcasting.

---

## 2. Technical Architecture & Component Breakdown

```
                       ┌──────────────────────────────────────────┐
                       │          React + Vite Frontend           │
                       │   (HTML5 Canvas, WebSocket Client)       │
                       └────────────────────┬─────────────────────┘
                                            │ WebSocket Connection
                                            ▼
                       ┌──────────────────────────────────────────┐
                       │             Go HTTP Server               │
                       │           (/ws Endpoint Handler)         │
                       └────────────────────┬─────────────────────┘
                                            │ Upgrades & Registers
                                            ▼
                       ┌──────────────────────────────────────────┐
                       │                 Hub                      │
                       │      (Room Registry & Event Router)      │
                       └───────────┬──────────────────┬───────────┘
                                   │                  │
                         Creates   │                  │ Dispatches
                                   ▼                  ▼
                       ┌──────────────────────┬───────────────────┐
                       │       Room           │      Client       │
                       │ (State, Round Timer, │ (ReadPump /       │
                       │  Scores, Drawer)     │  WritePump)       │
                       └──────────────────────┴───────────────────┘
```

### 2.1 Backend (Go)
The backend is structured around three primary abstractions: **Hub**, **Room**, and **Client**.

1. **Hub (`hub.go`)**:
   - Manages room creation and lookup in a synchronized map (`map[string]*Room`).
   - Acts as the central message router registering clients and dispatching socket events.
   - Generates random 8-character room codes (`ROOMXXXX`) upon client creation requests.

2. **Room (`room.go`)**:
   - Encapsulates game state: current active players, current drawer ID, secret target word, current round number, and player scores (`map[string]int`).
   - **State Machine**: Transitions between `"lobby"` and `"playing"`.
   - **Round Timers**: Uses Go's `time.AfterFunc` to automatically count down round durations (default 90 seconds) and end rounds, resetting state to lobby and revealing the secret word to all players.
   - **Thread Safety**: Uses `sync.RWMutex` to guard state mutations and player queries while releasing locks prior to message broadcasts to prevent deadlocks with client read/write loops.

3. **Client (`client.go`)**:
   - Wraps a gorilla/websocket `Conn`.
   - Runs paired goroutines per connection:
     - `readPump`: Reads incoming messages, enforces deadlines, processes socket pings/pongs, and routes inbound JSON messages to room handlers.
     - `writePump`: Reads outgoing messages from a buffered Go channel (`chan *Message`, capacity 256) and pushes them downstream over WebSocket.
   - Implements explicit channel synchronization (`registered chan struct{}`) ensuring room membership is fully registered in the `Hub` before handshakes return.

4. **Types & Envelope Protocol (`types.go`)**:
   - Defines a uniform JSON message envelope:
     ```json
     {
       "type": "MESSAGE_TYPE",
       "playerId": "player-XXXXXX",
       "data": { ... },
       "timestamp": 1724089800000
     }
     ```

### 2.2 Frontend (React + Vite)
The frontend (`frontend/src/App.jsx`, `App.css`) is built with React 18 and Vite:
- **Canvas Rendering Engine**: Captures high-frequency pointer events (`onPointerDown`, `onPointerMove`, `onPointerUp`, `onPointerLeave`) on an HTML5 `<canvas>` element.
- **Stroke Normalization**: Calculates normalized relative coordinates inside the canvas bounding box (`getBoundingClientRect`) to maintain drawing integrity across screen resizes and aspect ratios.
- **Game UI Layout**:
  - Left Panel: Live player list with real-time score indicators and room sharing info.
  - Center Column: Interactive Canvas board, drawer word prompt, round status indicator, and start button.
  - Right Panel: Live WebSocket guess stream, highlighting correct guesses and system announcements.

---

## 3. WebSocket Event Protocol Specification

| Message Type | Direction | Description | Payload Schema |
| :--- | :--- | :--- | :--- |
| `CREATE_ROOM` | Client → Server | Client requests a new game room | `{ "playerName": "Ada" }` |
| `JOIN_ROOM` | Client → Server | Client joins an existing room | `{ "roomCode": "ROOM1234", "playerName": "Bob" }` |
| `ROOM_JOINED` | Server → Client | Confirms successful room join | `{ "roomCode": "ROOM1234", "playerId": "player-123", "players": [...] }` |
| `PLAYER_JOINED` | Server → Client | Broadcasts new player arrival | `{ "id": "player-456", "name": "Charlie", "score": 0 }` |
| `START_GAME` | Client → Server | Initiates round starting sequence | `{}` |
| `GAME_STARTED` | Server → Client | Announces round start & assigns drawer | `{ "drawerId": "player-123", "word": "cat", "roundNum": 1, "timeLeft": 90 }` *(Word masked as `*****` for guessers)* |
| `DRAW_STROKE` | Client → Server | Drawer sends line segments | `{ "points": [{"x": 10, "y": 20}, ...], "color": "#162521", "size": 5 }` |
| `STROKE` | Server → Client | Broadcasts stroke to guessers | `{ "points": [...], "color": "#162521", "size": 5 }` |
| `GUESS` | Client → Server | Guesser submits text guess | `{ "text": "cat" }` |
| `GUESS_RECEIVED` | Server → Client | Broadcasts chat guess attempt | `{ "guesserId": "...", "guesserName": "Bob", "text": "cat", "isCorrect": true }` |
| `SCORE_UPDATE` | Server → Client | Broadcasts updated score | `{ "playerId": "player-456", "score": 10 }` |
| `ROUND_END` | Server → Client | Announces round time expiry or word guessed | `{ "word": "cat", "players": [...] }` |

---

## 4. Phase-by-Phase Development Log

### Phase 1: Server & WebSocket Core Foundation (`d4a4588`)
- Initialized Go module (`github.com/NITHISH-2006/goskribbl`).
- Built HTTP server listening on `:8080`/`:8081` with Gorilla WebSockets upgrader.
- Implemented core envelope struct, `Client` struct, and concurrent `Hub` runner.

### Phase 2: Room Creation & Broadcast Logic (`75564dc`)
- Implemented `Hub.CreateRoom` and `Hub.GetRoom`.
- Built player join routing, assigning unique IDs (`player-XXXXXXXX`).
- Added multi-client broadcasting (`Room.Broadcast`) and player list aggregation.

### Phase 3: Game Lifecycle & Word Selection (`ae595b0`)
- Introduced room game state machine (`lobby` ↔ `playing`).
- Implemented random drawer assignment and word selection from a dictionary (`dog`, `cat`, `house`, `car`, `tree`, `sun`, `apple`, `computer`).
- Built word masking logic (`GameStartedData.Word` masked for non-drawers).

### Phase 4: Real-time Canvas Stroke Synchronization (`a716a24`, `e626d27`)
- Implemented `MsgDrawStroke` / `MsgStroke` serialization.
- Bound client `readPump` to relay strokes instantly to all other clients in the room.

### Phase 5: Guess Processing & Scoring System (`dd024d5`)
- Implemented case-insensitive guess verification in `Room.HandleGuess`.
- Added point allocation (+10 points per correct guess).
- Broadcasted score updates (`MsgScoreUpdate`) to all room participants.

### Phase 6: Concurrency Hardening, Timers & Unit Testing
- Synchronized client WebSocket handshakes with channel signaling (`client.registered`).
- Fixed mutex re-entrancy deadlocks during room broadcasts by releasing locks before sending.
- Integrated automated round expiration timer (`time.AfterFunc`) and round-end announcements (`MsgRoundEnd`).
- Created unit tests (`room_test.go`) covering scoring, word masking, and timer expiry.

### Phase 7: Web Frontend Canvas & React Integration
- Built modern single-page web app using React 18, Vite, and custom CSS styling.
- Built pointer-event canvas drawing interface supporting stroke creation, clearing, and live drawing playback.
- Created live score table, chat guess history panel, dynamic room code display, and connection indicators.

### Phase 8: Project Documentation & Architectural Reporting
- Created detailed, non-generic human-authored `README.md` with complete installation and usage instructions.
- Generated `PROJECT_REPORT.md` detailing the design decisions, protocol, and phase breakdown.

---

## 5. Verification & Test Suite

### Go Backend Tests
```bash
cd backend
go test -v ./...
```
- `TestHandleGuessAwardsScoreAndBroadcasts`: Verifies correct guess score increments (+10) and correct output messages.
- `TestRoundEndsAndRevealsWord`: Verifies timer-triggered round state reset and word disclosure.

### Frontend Production Build Test
```bash
cd frontend
npm run build
```
- Compiles JSX/CSS into production JS bundle inside `frontend/dist/`.
