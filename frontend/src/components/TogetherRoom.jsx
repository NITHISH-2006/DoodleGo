import DrawingCanvas from './DrawingCanvas'
import PlayersPanel from './PlayersPanel'
import Toolbar from './Toolbar'
import { useDrawingTools } from '../hooks/useDrawingTools'

export default function TogetherRoom({
  room,
  playerId,
  connected,
  onStroke,
  canvasRef,
  onUndo,
  onClear,
  onDownload,
}) {
  const tools = useDrawingTools()
  const players = room.players || []

  return (
    <section className="room-layout together-layout">
      <PlayersPanel
        players={players}
        playerId={playerId}
        maxPlayers={room.maxPlayers}
        showScores={false}
        roomCode={room.roomCode}
        mode={room.mode}
      />

      <div className="board-column">
        <div className="game-status">
          <span className="status-dot status-dot-together" />
          <span>Everyone can draw at the same time</span>
          <span className="round-label">TOGETHER MODE</span>
        </div>

        <DrawingCanvas
          ref={canvasRef}
          enabled={connected}
          tool={tools.tool}
          color={tools.color}
          size={tools.size}
          onStroke={onStroke}
          label="Shared drawing canvas"
        />

        <Toolbar
          tool={tools.tool}
          setTool={tools.setTool}
          color={tools.color}
          setColor={tools.setColor}
          size={tools.size}
          setSize={tools.setSize}
          colors={tools.colors}
          onUndo={onUndo}
          onClear={onClear}
          onDownload={onDownload}
        />

        {players.length < 2 && (
          <p className="draw-note">You are the only one here. Share the room code so your partner can join.</p>
        )}
      </div>

      <aside className="session-panel">
        <div className="panel-heading">
          <span>SESSION</span>
          <span className={`conn-label ${connected ? 'live' : 'offline'}`.trim()}>
            <i />
            {connected ? 'LIVE' : 'OFFLINE'}
          </span>
        </div>

        <div className="session-card">
          <span className="kicker">DRAW TOGETHER</span>
          <h2>One canvas, two artists.</h2>
          <p>
            Every stroke is shared instantly. Pick a tool, switch colors, and build the drawing together.
          </p>
          <ul>
            <li>Both players can draw at the same time.</li>
            <li>Undo removes your last stroke.</li>
            <li>Download the finished artwork as a PNG.</li>
          </ul>
        </div>

        <button type="button" className="primary download-button" onClick={onDownload}>
          Download PNG <span aria-hidden="true">↓</span>
        </button>

        <div className="session-meta">
          <span>ROOM CODE</span>
          <b>{room.roomCode}</b>
          <small>Anyone with this code can join and draw.</small>
        </div>
      </aside>
    </section>
  )
}
