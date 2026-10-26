import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react'
import ClassicRoom from './components/ClassicRoom'
import Lobby from './components/Lobby'
import TogetherRoom from './components/TogetherRoom'
import {
  DIFFICULTY_EASY,
  MODE_CLASSIC,
  TOGETHER_FREE_DRAW,
  sortPlayers,
} from './lib/game'
import './App.css'

const DEFAULT_WS_URL =
  typeof window === 'undefined'
    ? 'ws://127.0.0.1:8081/ws'
    : `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.hostname || '127.0.0.1'}:8081/ws`
const SOCKET_URL = import.meta.env.VITE_WS_URL || import.meta.env.VITE_SOCKET_URL || DEFAULT_WS_URL
const NAME_KEY = 'doodlego:name'

function App() {
  const socket = useRef(null)
  const canvasRef = useRef(null)
  const playerIdRef = useRef('')
  const intentRef = useRef(null)
  const intentionalClose = useRef(false)

  const [screen, setScreen] = useState('lobby')
  const [entryMode, setEntryMode] = useState('create')
  const [lobbyMode, setLobbyMode] = useState(MODE_CLASSIC)
  const [name, setName] = useState(() => {
    try {
      return localStorage.getItem(NAME_KEY) || ''
    } catch {
      return ''
    }
  })
  const [roomCode, setRoomCode] = useState('')
  const [difficulty, setDifficulty] = useState(DIFFICULTY_EASY)
  const [togetherActivity, setTogetherActivity] = useState(TOGETHER_FREE_DRAW)
  const [roundDuration, setRoundDuration] = useState(90)
  const [relayTurnDuration, setRelayTurnDuration] = useState(60)

  const [connection, setConnection] = useState('idle')
  const [room, setRoom] = useState(null)
  const [game, setGame] = useState(null)
  const [mission, setMission] = useState(null)
  const [relay, setRelay] = useState(null)
  const [relayTimeLeft, setRelayTimeLeft] = useState(0)
  const [messages, setMessages] = useState([])
  const [guess, setGuess] = useState('')
  const [timeLeft, setTimeLeft] = useState(0)
  const [correctGuessers, setCorrectGuessers] = useState([])
  const [roundResult, setRoundResult] = useState(null)
  const [initialStrokes, setInitialStrokes] = useState(null)
  const [error, setError] = useState('')
  const [copied, setCopied] = useState(false)

  const connected = connection === 'connected'

  const send = useCallback((type, data) => {
    const ws = socket.current
    if (!ws || ws.readyState !== WebSocket.OPEN) return false
    ws.send(JSON.stringify({ type, data }))
    return true
  }, [])

  function handleMessage(message) {
    const data = message.data || {}

    switch (message.type) {
      case 'ROOM_JOINED': {
        playerIdRef.current = data.playerId
        intentRef.current = { type: 'join', roomCode: data.roomCode }
        setRoom({
          roomCode: data.roomCode,
          playerId: data.playerId,
          players: sortPlayers(data.players || []),
          status: data.status || 'lobby',
          mode: data.mode || MODE_CLASSIC,
          difficulty: data.difficulty || DIFFICULTY_EASY,
          togetherActivity: data.togetherActivity || TOGETHER_FREE_DRAW,
          roundDuration: data.roundDuration || 90,
          relayTurnDuration: data.relayTurnDuration || 60,
          maxPlayers: data.maxPlayers || 20,
        })
        setGame(data.game || null)
        setMission(data.mission || null)
        setRelay(data.relay || null)
        setRelayTimeLeft(data.relay?.timeLeft || 0)
        setTimeLeft(data.game?.timeLeft || 0)
        setMessages([])
        setCorrectGuessers([])
        setRoundResult(null)
        setInitialStrokes(data.strokes || [])
        setScreen('room')
        setError('')
        break
      }

      case 'PLAYER_JOINED':
        setRoom((current) =>
          current
            ? {
                ...current,
                players: sortPlayers([...current.players.filter((player) => player.id !== data.id), data]),
              }
            : current,
        )
        break

      case 'PLAYER_LEFT':
        setRoom((current) =>
          current ? { ...current, players: current.players.filter((player) => player.id !== data.playerId) } : current,
        )
        break

      case 'GAME_STARTED':
        setGame(data)
        setTimeLeft(data.timeLeft || 0)
        setMessages([])
        setGuess('')
        setCorrectGuessers([])
        setRoundResult(null)
        setInitialStrokes([])
        canvasRef.current?.setStrokes([])
        break

      case 'MISSION_STATE':
        setMission(data)
        break

      case 'RELAY_STATE':
        setRelay(data)
        setRelayTimeLeft(data.status === 'playing' ? data.timeLeft || 0 : 0)
        break

      case 'STROKE':
        if (data.ownerId !== playerIdRef.current) canvasRef.current?.applyStroke(data)
        break

      case 'CANVAS_STATE':
        canvasRef.current?.setStrokes(data.strokes || [])
        setInitialStrokes(data.strokes || [])
        break

      case 'GUESS_RECEIVED':
        setMessages((current) => [
          ...current.slice(-99),
          {
            id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
            ...data,
            correct: data.isCorrect,
          },
        ])
        if (data.isCorrect && data.guesserId) {
          setCorrectGuessers((current) =>
            current.includes(data.guesserId) ? current : [...current, data.guesserId],
          )
        }
        break

      case 'SCORE_UPDATE':
        setRoom((current) =>
          current
            ? {
                ...current,
                players: sortPlayers(
                  current.players.map((player) =>
                    player.id === data.playerId ? { ...player, score: data.score } : player,
                  ),
                ),
              }
            : current,
        )
        break

      case 'ROUND_END':
        setRoom((current) =>
          current ? { ...current, status: 'lobby', players: sortPlayers(data.players || current.players) } : current,
        )
        setGame(null)
        setTimeLeft(0)
        setGuess('')
        setCorrectGuessers([])
        setRoundResult({ word: data.word, players: data.players || [] })
        setMessages((current) => [
          ...current.slice(-99),
          {
            id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
            system: true,
            text: `Round over. The word was ${data.word}.`,
          },
        ])
        break

      case 'ERROR':
        setError(data.error || 'The server rejected the request.')
        setConnection('error')
        break

      default:
        break
    }
  }

  const connect = (override) => {
    const cleanName = name.trim().slice(0, 24)
    const intent =
      override ||
      (entryMode === 'create'
        ? {
            type: 'create',
            mode: lobbyMode,
            difficulty,
            togetherActivity,
            roundDuration,
            relayTurnDuration,
          }
        : { type: 'join', roomCode: roomCode.trim().toUpperCase() })

    if (!cleanName) {
      setError('Please enter your name.')
      return
    }
    if (intent.type === 'join' && !String(intent.roomCode || '').trim()) {
      setError('Please enter a room code.')
      return
    }

    try {
      localStorage.setItem(NAME_KEY, cleanName)
    } catch {
      // Ignore unavailable storage.
    }

    const previous = socket.current
    if (previous) {
      previous.onopen = null
      previous.onmessage = null
      previous.onerror = null
      previous.onclose = null
      try {
        previous.close()
      } catch {
        // Ignore close errors from an already closed socket.
      }
    }

    intentionalClose.current = false
    setError('')
    setConnection('connecting')
    setCopied(false)

    const ws = new WebSocket(SOCKET_URL)
    socket.current = ws
    intentRef.current =
      intent.type === 'create'
        ? {
            type: 'create',
            mode: intent.mode || MODE_CLASSIC,
            difficulty: intent.difficulty || DIFFICULTY_EASY,
            togetherActivity: intent.togetherActivity || TOGETHER_FREE_DRAW,
            roundDuration: intent.roundDuration || 90,
            relayTurnDuration: intent.relayTurnDuration || 60,
          }
        : { type: 'join', roomCode: String(intent.roomCode).trim().toUpperCase() }

    const initial =
      intent.type === 'create'
        ? {
            type: 'CREATE_ROOM',
            data: {
              playerName: cleanName,
              mode: intent.mode || MODE_CLASSIC,
              difficulty: intent.difficulty || DIFFICULTY_EASY,
              togetherActivity: intent.togetherActivity || TOGETHER_FREE_DRAW,
              roundDuration: intent.roundDuration || 90,
              relayTurnDuration: intent.relayTurnDuration || 60,
            },
          }
        : {
            type: 'JOIN_ROOM',
            data: {
              playerName: cleanName,
              roomCode: String(intent.roomCode).trim().toUpperCase(),
            },
          }

    ws.onopen = () => {
      setConnection('connected')
      ws.send(JSON.stringify(initial))
    }
    ws.onmessage = (event) => {
      let parsed
      try {
        parsed = JSON.parse(event.data)
      } catch {
        return
      }
      handleMessage(parsed)
    }
    ws.onerror = () => {
      setError('Could not connect to the game server.')
      setConnection('error')
    }
    ws.onclose = () => {
      if (socket.current !== ws) return
      if (intentionalClose.current) {
        intentionalClose.current = false
        setConnection('idle')
        return
      }
      setConnection((current) => (current === 'connecting' ? 'error' : 'closed'))
    }
  }

  const leaveRoom = () => {
    if (socket.current?.readyState === WebSocket.OPEN) {
      try {
        socket.current.send(JSON.stringify({ type: 'LEAVE_ROOM' }))
      } catch {
        // Ignore send errors while leaving.
      }
    }
    intentionalClose.current = true
    const ws = socket.current
    socket.current = null
    if (ws) {
      ws.onopen = null
      ws.onmessage = null
      ws.onerror = null
      ws.onclose = null
      try {
        ws.close()
      } catch {
        // Ignore close errors.
      }
    }
    playerIdRef.current = ''
    setConnection('idle')
    setScreen('lobby')
    setRoom(null)
    setGame(null)
    setMission(null)
    setRelay(null)
    setRelayTimeLeft(0)
    setMessages([])
    setGuess('')
    setCorrectGuessers([])
    setRoundResult(null)
    setInitialStrokes(null)
    setError('')
  }

  const rejoin = () => {
    const intent = intentRef.current
    if (!intent?.roomCode) {
      leaveRoom()
      return
    }
    connect({ type: 'join', roomCode: intent.roomCode })
  }

  const startGame = () => {
    if (!game && (room?.players?.length || 0) >= 2) send('START_GAME')
  }

  const submitGuess = (event) => {
    event.preventDefault()
    const player = room?.playerId
    const text = guess.trim()
    if (!text || !game || game.drawerId === player || correctGuessers.includes(player)) return
    if (send('GUESS', { text })) setGuess('')
  }

  const sendStroke = (stroke) => {
    send('DRAW_STROKE', stroke)
  }

  const undoStroke = () => {
    send('UNDO_STROKE')
  }

  const clearCanvas = () => {
    send('CLEAR_CANVAS')
  }

  const toggleMissionTask = (index) => {
    send('TOGGLE_MISSION_TASK', { index })
  }

  const nextMission = () => {
    send('NEXT_MISSION')
  }

  const startRelay = () => {
    send('START_RELAY')
  }

  const downloadDrawing = () => {
    const stamp = new Date().toISOString().slice(0, 19).replace(/[:T]/g, '-')
    canvasRef.current?.download(`doodlego-${room?.roomCode || 'drawing'}-${stamp}.png`)
  }

  const copyRoomCode = async () => {
    if (!room?.roomCode) return
    try {
      await navigator.clipboard.writeText(room.roomCode)
    } catch {
      const input = document.createElement('textarea')
      input.value = room.roomCode
      document.body.appendChild(input)
      input.select()
      try {
        document.execCommand('copy')
      } catch {
        // Copying is a convenience; ignore failures.
      }
      input.remove()
    }
    setCopied(true)
    window.setTimeout(() => setCopied(false), 1500)
  }

  useLayoutEffect(() => {
    if (screen === 'room' && canvasRef.current) {
      canvasRef.current.setStrokes(initialStrokes || [])
    }
  }, [screen, initialStrokes])

  useEffect(() => {
    if (!game) return undefined
    // Calibrate from the server's remaining time so client/server clock skew
    // cannot make the countdown wrong.
    const deadline = Date.now() + Math.max(0, Number(game.timeLeft) || 0) * 1000
    const update = () => setTimeLeft(Math.max(0, Math.ceil((deadline - Date.now()) / 1000)))
    update()
    const timer = window.setInterval(update, 500)
    return () => window.clearInterval(timer)
  }, [game])

  useEffect(() => {
    if (!relay || relay.status !== 'playing') return undefined
    const deadline = Date.now() + Math.max(0, Number(relay.timeLeft) || 0) * 1000
    const update = () => setRelayTimeLeft(Math.max(0, Math.ceil((deadline - Date.now()) / 1000)))
    update()
    const timer = window.setInterval(update, 500)
    return () => window.clearInterval(timer)
  }, [relay])

  useEffect(
    () => () => {
      const ws = socket.current
      if (ws) {
        ws.onclose = null
        try {
          ws.close()
        } catch {
          // Ignore close errors on unmount.
        }
      }
    },
    [],
  )

  return (
    <main className="app-shell">
      <header className="topbar">
        <div className="brand">
          <div className="brand-mark">D</div>
          <div className="brand-text">
            <strong>DoodleGo</strong>
            <span>draw fast. guess boldly.</span>
          </div>
        </div>

        {screen === 'room' && room && (
          <button type="button" className="room-pill" onClick={copyRoomCode} title="Copy room code">
            <span>ROOM</span> <b>{room.roomCode}</b>
            {copied && <em>copied</em>}
          </button>
        )}

        <div className="topbar-actions">
          {screen === 'room' && (
            <span className={`conn-pill ${connected ? 'live' : 'off'}`.trim()}>
              <i />
              {connected ? 'Live' : connection === 'connecting' ? 'Connecting' : 'Offline'}
            </span>
          )}
          {screen === 'room' && (
            <button type="button" className="ghost-button" onClick={leaveRoom}>
              Leave
            </button>
          )}
        </div>
      </header>

      {error && screen === 'room' && (
        <div className="error-banner" role="alert">
          {error}
        </div>
      )}

      {screen === 'lobby' ? (
        <Lobby
          entryMode={entryMode}
          setEntryMode={setEntryMode}
          name={name}
          setName={setName}
          roomCode={roomCode}
          setRoomCode={setRoomCode}
          mode={lobbyMode}
          setMode={setLobbyMode}
          difficulty={difficulty}
          setDifficulty={setDifficulty}
          activity={togetherActivity}
          setActivity={setTogetherActivity}
          roundDuration={roundDuration}
          setRoundDuration={setRoundDuration}
          relayTurnDuration={relayTurnDuration}
          setRelayTurnDuration={setRelayTurnDuration}
          onSubmit={(event) => {
            event.preventDefault()
            connect()
          }}
          error={error}
          connection={connection}
        />
      ) : room?.mode === 'together' ? (
        <TogetherRoom
          room={room}
          playerId={room.playerId}
          connected={connected}
          onStroke={sendStroke}
          canvasRef={canvasRef}
          onUndo={undoStroke}
          onClear={clearCanvas}
          onDownload={downloadDrawing}
          mission={mission}
          relay={relay}
          relayTimeLeft={relayTimeLeft}
          onToggleMissionTask={toggleMissionTask}
          onNextMission={nextMission}
          onStartRelay={startRelay}
        />
      ) : (
        <ClassicRoom
          room={room}
          game={game}
          playerId={room.playerId}
          timeLeft={timeLeft}
          messages={messages}
          guess={guess}
          setGuess={setGuess}
          onSubmitGuess={submitGuess}
          onStart={startGame}
          connected={connected}
          roundResult={roundResult}
          onStroke={sendStroke}
          canvasRef={canvasRef}
          correctGuessers={correctGuessers}
          onUndo={undoStroke}
          onClear={clearCanvas}
        />
      )}

      {screen === 'room' && !connected && (
        <div className="connection-overlay" role="dialog" aria-modal="true" aria-label="Connection lost">
          <div className="overlay-card">
            <span className="kicker">CONNECTION LOST</span>
            <h2>You dropped out of the room.</h2>
            <p>Rejoin to keep drawing. The shared canvas is restored when the room is still open.</p>
            {error && <p className="error">{error}</p>}
            <div className="overlay-actions">
              <button type="button" className="primary" onClick={rejoin} disabled={connection === 'connecting'}>
                {connection === 'connecting' ? 'Rejoining…' : 'Rejoin room'}
              </button>
              <button type="button" className="ghost-button" onClick={leaveRoom}>
                Back to lobby
              </button>
            </div>
          </div>
        </div>
      )}
    </main>
  )
}

export default App
