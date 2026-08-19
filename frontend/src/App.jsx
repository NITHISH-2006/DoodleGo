import { useEffect, useRef, useState } from 'react'
import './App.css'

const SOCKET_URL = import.meta.env.VITE_WS_URL || import.meta.env.VITE_SOCKET_URL || 'ws://127.0.0.1:8081/ws'

function App() {
  const socket = useRef(null)
  const canvas = useRef(null)
  const drawing = useRef(false)
  const lastPoint = useRef(null)
  const [screen, setScreen] = useState('lobby')
  const [mode, setMode] = useState('create')
  const [name, setName] = useState('')
  const [roomCode, setRoomCode] = useState('')
  const [room, setRoom] = useState(null)
  const [playerId, setPlayerId] = useState('')
  const [messages, setMessages] = useState([])
  const [guess, setGuess] = useState('')
  const [error, setError] = useState('')
  const [connected, setConnected] = useState(false)
  const [timeLeft, setTimeLeft] = useState(0)

  const send = (type, data) => {
    if (socket.current?.readyState === WebSocket.OPEN) socket.current.send(JSON.stringify({ type, data }))
  }

  const connect = () => {
    if (!name.trim() || (mode === 'join' && !roomCode.trim())) return
    setError('')
    const ws = new WebSocket(SOCKET_URL)
    socket.current = ws
    ws.onopen = () => {
      setConnected(true)
      send(mode === 'create' ? 'CREATE_ROOM' : 'JOIN_ROOM', mode === 'create' ? { playerName: name.trim() } : { roomCode: roomCode.trim().toUpperCase(), playerName: name.trim() })
    }
    ws.onmessage = (event) => handleMessage(JSON.parse(event.data))
    ws.onerror = () => setError('Could not connect to the game server.')
    ws.onclose = () => setConnected(false)
  }

  const handleMessage = (message) => {
    const data = message.data || {}
    if (message.type === 'ROOM_JOINED') {
      setPlayerId(data.playerId)
      setRoom({ ...data, players: data.players || [] })
      setScreen('game')
    } else if (message.type === 'PLAYER_JOINED') {
      setRoom((current) => current && { ...current, players: [...current.players.filter((p) => p.id !== data.id), data] })
    } else if (message.type === 'PLAYER_LEFT') {
      setRoom((current) => current && { ...current, players: current.players.filter((p) => p.id !== data.playerId) })
    } else if (message.type === 'GAME_STARTED') {
      setRoom((current) => current && { ...current, state: 'playing', game: data })
      setTimeLeft(data.timeLeft || 90)
      clearCanvas()
    } else if (message.type === 'STROKE') {
      drawStroke(data)
    } else if (message.type === 'GUESS_RECEIVED') {
      setMessages((current) => [...current.slice(-40), { ...data, correct: data.isCorrect }])
    } else if (message.type === 'SCORE_UPDATE') {
      setRoom((current) => current && { ...current, players: current.players.map((player) => player.id === data.playerId ? { ...player, score: data.score } : player) })
    } else if (message.type === 'ROUND_END') {
      setRoom((current) => current && { ...current, state: 'lobby', game: null, players: data.players || current.players })
      setTimeLeft(0)
      setMessages((current) => [...current, { system: true, text: `Round ended. The word was ${data.word}.` }])
    } else if (message.type === 'ERROR') setError(data.error || 'The server rejected the request.')
  }

  const clearCanvas = () => { const context = canvas.current?.getContext('2d'); if (context) { context.fillStyle = '#fffdf8'; context.fillRect(0, 0, canvas.current.width, canvas.current.height) } }
  const drawStroke = (stroke) => { const context = canvas.current?.getContext('2d'); if (!context || !stroke.points?.length) return; context.strokeStyle = stroke.color || '#162521'; context.lineWidth = stroke.size || 5; context.beginPath(); context.moveTo(stroke.points[0].x, stroke.points[0].y); stroke.points.slice(1).forEach((point) => context.lineTo(point.x, point.y)); context.stroke() }
  const pointFromEvent = (event) => { const bounds = canvas.current.getBoundingClientRect(); return { x: (event.clientX - bounds.left) * canvas.current.width / bounds.width, y: (event.clientY - bounds.top) * canvas.current.height / bounds.height } }
  const beginDrawing = (event) => { if (room?.game?.drawerId !== playerId) return; drawing.current = true; lastPoint.current = pointFromEvent(event) }
  const continueDrawing = (event) => { if (!drawing.current || !lastPoint.current) return; const point = pointFromEvent(event); const stroke = { points: [lastPoint.current, point], color: '#162521', size: 6 }; drawStroke(stroke); send('DRAW_STROKE', stroke); lastPoint.current = point }
  const submitGuess = (event) => { event.preventDefault(); if (!guess.trim()) return; send('GUESS', { text: guess.trim() }); setGuess('') }
  const isDrawer = room?.game?.drawerId === playerId

  useEffect(() => () => socket.current?.close(), [])
  useEffect(() => { const context = canvas.current?.getContext('2d'); if (context) { context.lineCap = 'round'; context.lineJoin = 'round'; clearCanvas() } }, [screen])
  useEffect(() => { if (!room?.game || timeLeft <= 0) return undefined; const timer = window.setInterval(() => setTimeLeft((current) => Math.max(0, current - 1)), 1000); return () => window.clearInterval(timer) }, [room?.game, timeLeft])

  return <main className="app-shell">
    <header className="topbar"><div className="brand-mark">D</div><div><strong>DoodleGo</strong><span>draw fast. guess boldly.</span></div>{screen === 'game' && <div className="room-pill">ROOM <b>{room?.roomCode}</b></div>}</header>
    {screen === 'lobby' ? <section className="welcome"><div className="intro"><p className="kicker">REAL-TIME DRAWING ROOM</p><h1>Make a mark.<br /><em>Make them guess.</em></h1><p className="lede">A tiny online sketchbook for friends, strangers, and suspiciously good doodlers.</p></div><div className="join-panel"><div className="tab-row"><button className={mode === 'create' ? 'active' : ''} onClick={() => setMode('create')}>Create room</button><button className={mode === 'join' ? 'active' : ''} onClick={() => setMode('join')}>Join room</button></div><label>Your name<input value={name} onChange={(event) => setName(event.target.value)} placeholder="e.g. Ada" maxLength="24" /></label>{mode === 'join' && <label>Room code<input value={roomCode} onChange={(event) => setRoomCode(event.target.value)} placeholder="ROOM1234" maxLength="8" /></label>}<button className="primary" onClick={connect}>Enter the room <span>↗</span></button>{error && <p className="error">{error}</p>}<small>{connected ? 'Connected' : 'Server: 127.0.0.1:8081'}</small></div></section> : <section className="game-layout"><aside className="side-panel"><div className="panel-heading"><span>PLAYERS</span><b>{room?.players?.length || 0}/20</b></div><div className="players">{room?.players?.map((player) => <div className={player.id === playerId ? 'player me' : 'player'} key={player.id}><span className="avatar">{player.name[0]?.toUpperCase()}</span><span>{player.name}{player.id === playerId && ' (you)'}</span><strong>{player.score}</strong></div>)}</div><div className="room-note">Share the room code<br /><b>{room?.roomCode}</b></div></aside><div className="board-column"><div className="game-status"><span className={room?.game ? 'live-dot' : 'wait-dot'}></span>{room?.game ? (isDrawer ? `Your word is ${room.game.word}` : 'Watch closely and guess') : 'Waiting for another player'}<span className="round-label">{room?.game ? `ROUND ${room.game.roundNum} · ${timeLeft}s` : 'LOBBY'}</span></div><div className="canvas-wrap"><canvas ref={canvas} width="900" height="560" onPointerDown={beginDrawing} onPointerMove={continueDrawing} onPointerUp={() => { drawing.current = false; lastPoint.current = null }} onPointerLeave={() => { drawing.current = false; lastPoint.current = null }} /></div>{!room?.game && <button className="primary start-button" onClick={() => send('START_GAME')}>Start game <span>▶</span></button>}</div><aside className="chat-panel"><div className="panel-heading"><span>GUESSES</span><span className="connected-label">● {connected ? 'LIVE' : 'OFFLINE'}</span></div><div className="chat-log">{messages.map((message, index) => message.system ? <p className="system-message" key={index}>{message.text}</p> : <p className={message.correct ? 'correct-message' : ''} key={index}><b>{message.guesserName}</b> {message.correct ? 'got it!' : message.text}</p>)}</div><form className="guess-form" onSubmit={submitGuess}><input disabled={!room?.game || isDrawer} value={guess} onChange={(event) => setGuess(event.target.value)} placeholder={isDrawer ? 'You are drawing' : 'Type a guess...'} maxLength="80" /><button aria-label="Send guess">↗</button></form></aside></section>}
  </main>
}

export default App
