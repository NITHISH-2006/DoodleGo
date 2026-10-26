# DoodleGo

[![Live App](https://img.shields.io/badge/Live_Demo-doodlego.vercel.app-brightgreen?style=for-the-badge)](https://doodlego.vercel.app/)
[![GitHub Repo](https://img.shields.io/badge/GitHub-Repository-blue?style=for-the-badge)](https://github.com/NITHISH-2006/DoodleGo)

DoodleGo is a real-time multiplayer drawing and word-guessing game built with Go and React. Create private rooms, invite friends with a room code, draw on a shared canvas, and guess the secret word.

- **Live Demo**: [https://doodlego.vercel.app](https://doodlego.vercel.app/)
- **Backend**: [https://doodlego.onrender.com](https://doodlego.onrender.com)

---

## Features

### Classic mode
- **Real-time canvas sync** with low-latency stroke broadcasting over WebSockets.
- **Room management** with instant room creation and join-by-code.
- **Selectable round length**: 30s, 60s, 90s, 2 minutes, or 3 minutes.
- **Turn-based rounds** with automatic drawer rotation and 190+ words.
- **Score once per round**: a player earns +10 only on their first correct guess; repeat guesses are ignored.
- **No word leaks**: the secret word is only sent to the drawer, and correct guesses are masked in broadcasts.
- **Mid-game joining**: late joiners receive the active round, masked word, timer, and full canvas history.
- **Live scoring and chat**, plus a round-over reveal and automatic next-round start.
- **Disconnect handling** with a rejoin flow and canvas restoration.

### Doodle Together mode
- **Two or more players draw on the same canvas at the same time** — ideal for friends and couples.
- No timer, no guessing, no scoring; just a shared sketchbook.
- **Download the finished drawing as a PNG** at any time.
- Late joiners receive the complete drawing history.

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
│   ├── room.go        # Room state machine, scoring, strokes, round timers
│   ├── room_test.go   # Unit tests for scoring, rounds, canvas and modes
│   ├── server.go      # WebSocket handshake, validation & initial state
│   ├── types.go       # Message structs, event types & JSON envelopes
│   └── words.go       # Word list for classic mode
│
├── frontend/
│   ├── src/
│   │   ├── App.jsx                    # Connection state, protocol handlers
│   │   ├── components/
│   │   │   ├── ClassicRoom.jsx
│   │   │   ├── DrawingCanvas.jsx
│   │   │   ├── GuessPanel.jsx
│   │   │   ├── Lobby.jsx
│   │   │   ├── PlayersPanel.jsx
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

- `CREATE_ROOM`: `{ playerName, mode: "classic" | "together", roundDuration }`
- `JOIN_ROOM`: `{ playerName, roomCode }`
- `START_GAME`: start a classic round (any player, needs 2+ players)
- `DRAW_STROKE`: `{ strokeId, points: [{x, y}], color, size, tool, opacity }`
- `UNDO_STROKE`: remove the sender's last gesture
- `CLEAR_CANVAS`: clear the shared canvas
- `GUESS`: `{ text }`
- `LEAVE_ROOM`: leave the room

### Server → Client

- `ROOM_JOINED`: room code, player id, players, mode, status, round duration, active game, stroke history
- `PLAYER_JOINED` / `PLAYER_LEFT`: player list updates
- `GAME_STARTED`: drawer, word (drawer only), masked word, round number, timer deadline
- `STROKE`: relayed stroke chunk
- `CANVAS_STATE`: full canvas state after join, undo, or clear
- `GUESS_RECEIVED`: guess result (correct guesses are masked)
- `SCORE_UPDATE`: updated player score
- `ROUND_END`: revealed word and final player list
- `ERROR`: handshake or validation error

---

## License

MIT License. Free for educational and open-source use.
