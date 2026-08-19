# DoodleGo

[![Live App](https://img.shields.io/badge/Live_Demo-doodlego.vercel.app-brightgreen?style=for-the-badge)](https://doodlego.vercel.app/)
[![GitHub Repo](https://img.shields.io/badge/GitHub-Repository-blue?style=for-the-badge)](https://github.com/NITHISH-2006/DoodleGo)

DoodleGo is a real-time multiplayer drawing and word-guessing game built with Go and React. Players can create private rooms, join via room codes, draw on a shared canvas, and guess words in real time.

- 🌐 **Live Demo**: [https://doodlego.vercel.app](https://doodlego.vercel.app/)
- ⚙️ **Backend**: [https://doodlego.onrender.com](https://doodlego.onrender.com)

---

## Features

- **Real-Time Canvas Sync**: HTML5 Canvas drawing with instant stroke broadcasting across WebSockets.
- **Room Management**: Instant room creation and joining via unique room codes (e.g. `ROOM1234`).
- **Turn-Based Gameplay**: Automatic drawer assignment, word selection, and round timers.
- **Word Masking**: Secret target word is visible only to the current drawer; other players see hint placeholders (`*****`).
- **Live Scoring & Chat**: Instant guess checking with score allocation (+10 points per correct guess) and real-time chat log.
- **Clean Concurrency Model**: Go backend utilizing Hub-Room-Client goroutine architecture with mutex protection and ping/pong keep-alives.

---

## Tech Stack

- **Backend**: Go 1.20+, Gorilla WebSockets (`github.com/gorilla/websocket`)
- **Frontend**: React 18, Vite, HTML5 Canvas API, Vanilla CSS
- **Protocol**: JSON-over-WebSocket custom envelope format

---

## Project Structure

```
.
├── backend/
│   ├── client.go      # WebSocket client connection & read/write pumps
│   ├── config.go      # Server timeouts, port configuration, round settings
│   ├── main.go        # HTTP server entrypoint and route handlers
│   ├── hub.go         # Central room registry and connection router
│   ├── room.go        # Room state machine, scoring, drawing & round timers
│   ├── room_test.go   # Go unit tests for room logic & guessing
│   ├── server.go      # WebSocket handshake & upgrade handler
│   └── types.go       # Message structs, event types & JSON envelopes
│
├── frontend/
│   ├── src/
│   │   ├── App.jsx    # React application layout, canvas & WS hooks
│   │   ├── App.css    # Clean custom UI styling & responsive layout
│   │   └── main.jsx   # App entry point
│   ├── index.html     # Base HTML shell
│   └── vite.config.js # Vite configuration
│
├── PROJECT_REPORT.md  # Detailed technical architecture report
└── README.md          # Project documentation
```

---

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.20 or newer
- [Node.js](https://nodejs.org/) 18 or newer (npm included)

---

### Running the Backend

1. Navigate to the `backend` directory:
   ```bash
   cd backend
   ```

2. Download Go module dependencies:
   ```bash
   go mod download
   ```

3. Run the Go server:
   ```bash
   go run .
   ```
   The backend server will start on `http://127.0.0.1:8081` with WebSocket endpoint `ws://127.0.0.1:8081/ws`.

---

### Running the Frontend

1. Open a new terminal and navigate to the `frontend` directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start the Vite development server:
   ```bash
   npm run dev
   ```
   Open `http://localhost:5173` in your browser. Open a second tab or window to test multiplayer drawing and guessing between two players.

---

## Running Backend Tests

To run the Go unit tests for room logic, scoring, and round timers:

```bash
cd backend
go test -v ./...
```

---

## WebSocket Communication Protocol

All socket messages use a JSON envelope structure:

```json
{
  "type": "MESSAGE_TYPE",
  "playerId": "player-ID",
  "data": {},
  "timestamp": 1724089800000
}
```

### Core Message Types

- `CREATE_ROOM` / `JOIN_ROOM`: Establish room connection.
- `ROOM_JOINED`: Confirms room code and returns initial player list.
- `START_GAME`: Sent by host to begin round.
- `GAME_STARTED`: Broadcasts drawer ID, secret word (or masked length), and round duration.
- `DRAW_STROKE` / `STROKE`: Sends/receives continuous line coordinates (`{ points: [{x, y}], color, size }`).
- `GUESS` / `GUESS_RECEIVED`: Sends player text guess and broadcasts results.
- `SCORE_UPDATE`: Updates player score leaderboard.
- `ROUND_END`: Broadcasted when the round timer expires or word is guessed.

---

## License

MIT License. Free for educational and open-source use.