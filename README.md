# DoodleGo

[![Live App](https://img.shields.io/badge/Live_Demo-doodlego.vercel.app-brightgreen?style=for-the-badge)](https://doodlego.vercel.app/)
[![GitHub Repo](https://img.shields.io/badge/GitHub-Repository-blue?style=for-the-badge)](https://github.com/NITHISH-2006/DoodleGo)

DoodleGo is a real-time multiplayer drawing game built with Go and React. Create private rooms, invite friends with a room code, guess secret words, complete team missions, or pass the pen in a drawing relay.

- **Live Demo**: [https://doodlego.vercel.app](https://doodlego.vercel.app/)
- **Backend**: [https://doodlego.onrender.com](https://doodlego.onrender.com)

---

## Features

### Classic mode
- **Real-time canvas sync** with low-latency stroke broadcasting over WebSockets.
- **Room management** with instant room creation and join-by-code.
- **Three word difficulties**: 205 Easy, 210 Medium, and 210 Hard words (625 unique words total).
- **No-repeat word decks** per room, with automatic reshuffling only after the deck is exhausted.
- **Selectable round length**: 30s, 60s, 90s, 2 minutes, or 3 minutes.
- **Turn-based rounds** with automatic drawer rotation.
- **Score once per round**: a player earns +10 only on their first correct guess; repeat guesses are ignored.
- **No word leaks**: the secret word is only sent to the drawer, and correct guesses are masked in broadcasts.
- **Mid-game joining**: late joiners receive the active round, masked word, timer, difficulty, and full canvas history.
- **Live scoring and chat**, plus a round-over reveal and automatic next-round start.
- **Disconnect handling** with a rejoin flow and canvas restoration.

### Doodle Together mode
Choose one of three shared activities when creating a room. All of them support live stroke syncing, late joining, and mobile drawing.

- **Free Draw**: everyone draws on the same canvas with no timer, guessing, or scoring. Download the finished artwork as a PNG at any time.
- **Team Missions**: complete a shared checklist from a deck of 42 missions with five drawing tasks each. Any player can check a task, there is no timer, and the group advances only when every task is checked. The canvas clears for each new mission.
- **Drawing Relay**: 2–4 players take turns drawing a surprise prompt. Every player gets two turns, only the active player has the pen, and the canvas clears between turns. The default turn length is 60 seconds, with 30s, 60s, and 90s options. Turns advance automatically, and the relay pauses if fewer than two players remain. Late joiners can watch the active relay and play in the next one.

### Drawing tools
- Pen, marker/highlighter, and eraser.
- 12 color swatches plus a custom color picker.
- Adjustable brush size from 2px to 36px.
- Undo the last gesture and clear the whole canvas.
- Pointer capture, coalesced pointer events, and touch-friendly canvas support on mobile.

### Mobile-first responsive UI
- Single-column room layout on phones and tablets.
- Wrapping/scrollable toolbar with large touch targets.
- Large touch-friendly guess input, safe-area padding, and no page scrolling while drawing.
- Round timer calibrated from the server’s remaining time to avoid client/server clock skew.

---

## Tech Stack

- **Backend**: Go 1.26+, Gorilla WebSockets (`github.com/gorilla/websocket`)
- **Frontend**: React 19, Vite, HTML5 Canvas API, Vanilla CSS
- **Protocol**: JSON-over-WebSocket custom envelope format

---

## Project Structure

```
.
├── backend/
│   ├── client.go      # WebSocket client connection & read/write pumps
│   ├── config.go      # Port, timeouts, max players, allowed origins
│   ├── hub.go         # Room registry, atomic join/leave, unique room codes
│   ├── main.go        # HTTP server entrypoint and health route
│   ├── missions.go    # Team mission deck and checklist state
│   ├── relay.go       # Drawing Relay prompts, schedule and turn timers
│   ├── room.go        # Room state machine, scoring, strokes, activity state
│   ├── room_test.go   # Unit tests for scoring, rounds, canvas and modes
│   ├── server.go      # WebSocket handshake, validation & initial state
│   ├── together_test.go # Team Missions and Drawing Relay tests
│   ├── types.go       # Message structs, event types & JSON envelopes
│   ├── words.go       # Easy, Medium and Hard word banks
│   └── words_test.go  # Word bank and no-repeat deck tests
│
├── frontend/
│   ├── src/
│   │   ├── App.jsx                    # Connection state, protocol handlers
│   │   ├── components/
│   │   │   ├── ClassicRoom.jsx
│   │   │   ├── DrawingCanvas.jsx
│   │   │   ├── GuessPanel.jsx
│   │   │   ├── Lobby.jsx
│   │   │   ├── MissionPanel.jsx
│   │   │   ├── PlayersPanel.jsx
│   │   │   ├── RelayPanel.jsx
│   │   │   ├── TogetherRoom.jsx
│   │   │   └── Toolbar.jsx
│   │   ├── hooks/useDrawingTools.js
│   │   ├── lib/drawing.js
│   │   ├── lib/game.js
│   │   ├── App.css
│   │   └── main.jsx
│   ├── index.html
│   └── vite.config.js
│
└── README.md
```

---

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.26 or newer
- [Node.js](https://nodejs.org/) 22 or newer (npm included)

### Run the backend

```bash
cd backend
go mod download
go run .
```

The backend starts on `http://127.0.0.1:8081` with WebSocket endpoint `ws://127.0.0.1:8081/ws`.

### Run the frontend

```bash
cd frontend
npm install
npm run dev
```

Open `http://localhost:5173`. Open a second tab to test multiplayer play. To test from a phone on the same network, run `npm run dev:local` and open `http://<your-lan-ip>:5174`.

### Environment variables

Backend:

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `8081` | HTTP/WebSocket port |
| `MAX_PLAYERS` | `20` | Maximum players per room |
| `ALLOWED_ORIGINS` | `*` | Comma-separated allowed WebSocket origins |

Frontend:

| Variable | Default | Description |
| --- | --- | --- |
| `VITE_WS_URL` | `ws(s)://<current-host>:8081/ws` | WebSocket server URL (defaults to the page host, port 8081) |
| `VITE_SOCKET_URL` | — | Fallback WebSocket URL variable |

Production builds use `frontend/.env.production`.

---

## Testing & Verification

```bash
# Backend tests and static checks
cd backend
go test -v ./...
go vet ./...

# Frontend lint and production build
cd frontend
npm run lint
npm run build
```

---

## WebSocket Communication Protocol

All socket messages use a JSON envelope:

```json
{
  "type": "MESSAGE_TYPE",
  "playerId": "player-id",
  "data": {},
  "timestamp": 1724089800000
}
```

### Client → Server

- `CREATE_ROOM`: `{ playerName, mode: "classic" | "together", difficulty, togetherActivity, roundDuration, relayTurnDuration }`
- `JOIN_ROOM`: `{ playerName, roomCode }`
- `START_GAME`: start a classic round (any player, needs 2+ players)
- `DRAW_STROKE`: `{ strokeId, points: [{x, y}], color, size, tool, opacity }`
- `UNDO_STROKE`: remove the sender's last gesture
- `CLEAR_CANVAS`: clear the shared canvas (disabled for relay viewers)
- `GUESS`: `{ text }`
- `TOGGLE_MISSION_TASK`: `{ index }` — any player can toggle a Team Missions task
- `NEXT_MISSION`: advance after every Team Missions task is checked
- `START_RELAY`: start or restart a Drawing Relay (needs 2–4 players)
- `LEAVE_ROOM`: leave the room

`difficulty` is `"easy"`, `"medium"`, or `"hard"` and defaults to `"easy"`. `togetherActivity` is `"free_draw"`, `"missions"`, or `"relay"` and defaults to `"free_draw"`. `relayTurnDuration` is the turn length in milliseconds and is normalized to 30, 60, or 90 seconds.

### Server → Client

- `ROOM_JOINED`: room code, player id, players, mode, status, difficulty, together activity, round duration, relay turn duration, active game, mission state, relay state, stroke history
- `PLAYER_JOINED` / `PLAYER_LEFT`: player list updates
- `GAME_STARTED`: drawer, word (drawer only), masked word, difficulty, round number, timer deadline
- `STROKE`: relayed stroke chunk
- `CANVAS_STATE`: full canvas state after join, undo, clear, activity advance, or relay turn change
- `GUESS_RECEIVED`: guess result (correct guesses are masked)
- `MISSION_STATE`: current mission, five task labels, checked flags, mission number, and completed count
- `RELAY_STATE`: relay status (`lobby`, `playing`, `paused`, `finished`), prompt, drawer, turn number, total turns, time left, turn deadline, turn duration, and turns per player
- `SCORE_UPDATE`: updated player score
- `ROUND_END`: revealed word and final player list
- `ERROR`: handshake or validation error

---

## License

MIT License. Free for educational and open-source use.
