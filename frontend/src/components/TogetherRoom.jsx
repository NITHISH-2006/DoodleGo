import DrawingCanvas from './DrawingCanvas'
import MissionPanel from './MissionPanel'
import PlayersPanel from './PlayersPanel'
import RelayPanel from './RelayPanel'
import Toolbar from './Toolbar'
import { useDrawingTools } from '../hooks/useDrawingTools'
import { formatClock, TOGETHER_FREE_DRAW, TOGETHER_MISSIONS, TOGETHER_RELAY } from '../lib/game'

export default function TogetherRoom({
  room,
  playerId,
  connected,
  onStroke,
  canvasRef,
  onUndo,
  onClear,
  onDownload,
  mission,
  relay,
  relayTimeLeft,
  onToggleMissionTask,
  onNextMission,
  onStartRelay,
}) {
  const tools = useDrawingTools()
  const players = room.players || []
  const activity = room.togetherActivity || TOGETHER_FREE_DRAW
  const isMissions = activity === TOGETHER_MISSIONS
  const isRelay = activity === TOGETHER_RELAY
  const relayActive = Boolean(isRelay && relay?.status === 'playing' && relay.drawerId === playerId)
  const canvasEnabled = connected && (!isRelay || relayActive)
  const showToolbar = connected && (!isRelay || relayActive)

  let statusText = 'Everyone can draw at the same time'
  let statusClass = 'status-dot status-dot-together'
  let statusLabel = 'FREE DRAW'
  let statusClock = ''

  if (isMissions) {
    statusText = mission?.mission ? `Team mission: ${mission.mission.title}` : 'Loading the team mission'
    statusLabel = mission ? `MISSION ${mission.number}` : 'TEAM MISSION'
  }

  if (isRelay) {
    if (relay?.status === 'playing') {
      statusText = relayActive ? 'You have the pen' : `${relay.drawerName || 'A player'} has the pen`
      statusClass = relayActive ? 'status-dot status-dot-live' : 'status-dot status-dot-wait'
      statusClock = formatClock(relayTimeLeft)
    } else if (relay?.status === 'paused') {
      statusText = 'Relay paused until another player joins'
      statusClass = 'status-dot status-dot-wait'
    } else if (relay?.status === 'finished') {
      statusText = 'Relay complete'
      statusClass = 'status-dot status-dot-together'
    } else {
      statusText = players.length >= 2 ? 'Ready to start the relay' : 'Waiting for another player'
      statusClass = 'status-dot status-dot-wait'
    }
    statusLabel = relay?.status === 'playing' ? `RELAY · ${formatClock(relayTimeLeft)}` : 'DRAWING RELAY'
  }

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
          <span className={statusClass} />
          <span>{statusText}</span>
          <span className="round-label">{statusLabel}</span>
        </div>

        {statusClock && (
          <div className="timer-track" aria-hidden="true">
            <span
              className="timer-fill"
              style={{
                width: `${relay?.turnDuration ? Math.max(0, Math.min(100, (relayTimeLeft / relay.turnDuration) * 100)) : 0}%`,
              }}
            />
          </div>
        )}

        <DrawingCanvas
          ref={canvasRef}
          enabled={canvasEnabled}
          tool={tools.tool}
          color={tools.color}
          size={tools.size}
          onStroke={onStroke}
          label={isRelay ? 'Relay drawing canvas' : 'Shared drawing canvas'}
        />

        {showToolbar && (
          <Toolbar
            tool={tools.tool}
            setTool={tools.setTool}
            color={tools.color}
            setColor={tools.setColor}
            size={tools.size}
            setSize={tools.setSize}
            colors={tools.colors}
            onUndo={onUndo}
            onClear={isRelay ? null : onClear}
            onDownload={isRelay ? undefined : onDownload}
          />
        )}

        {isRelay && relay?.status === 'playing' && !relayActive && (
          <p className="draw-note">
            <b>{relay.drawerName || 'Another player'}</b> is drawing. The canvas clears when the next turn starts.
          </p>
        )}

        {isRelay && relay?.status === 'finished' && (
          <p className="draw-note">The final drawing stays on the canvas while you decide who gets the pen next.</p>
        )}

        {!isRelay && players.length < 2 && (
          <p className="draw-note">You are the only one here. Share the room code so others can join.</p>
        )}

        {isRelay && relay?.status === 'lobby' && players.length < 2 && (
          <p className="draw-note">Drawing Relay needs at least 2 players. Share the room code to get started.</p>
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

        {isMissions && (
          <MissionPanel
            mission={mission}
            connected={connected}
            onToggleTask={onToggleMissionTask}
            onNextMission={onNextMission}
          />
        )}

        {isRelay && (
          <RelayPanel
            relay={relay}
            relayTimeLeft={relayTimeLeft}
            players={players}
            playerId={playerId}
            connected={connected}
            onStartRelay={onStartRelay}
          />
        )}

        {!isMissions && !isRelay && (
          <>
            <div className="session-card">
              <span className="kicker">DRAW TOGETHER</span>
              <h2>One canvas, many artists.</h2>
              <p>
                Every stroke is shared instantly. Pick a tool, switch colors, and build the drawing together.
              </p>
              <ul>
                <li>Everyone can draw at the same time.</li>
                <li>Undo removes your last stroke.</li>
                <li>Download the finished artwork as a PNG.</li>
              </ul>
            </div>

            <button type="button" className="primary download-button" onClick={onDownload}>
              Download PNG <span aria-hidden="true">↓</span>
            </button>
          </>
        )}

        <div className="session-meta">
          <span>ROOM CODE</span>
          <b>{room.roomCode}</b>
          <small>
            {isRelay ? 'Up to 4 players can join this relay room.' : 'Anyone with this code can join and draw.'}
          </small>
        </div>
      </aside>
    </section>
  )
}
