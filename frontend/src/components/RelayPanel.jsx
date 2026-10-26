import { formatClock } from '../lib/game'

export default function RelayPanel({
  relay,
  relayTimeLeft,
  players,
  playerId,
  connected,
  onStartRelay,
}) {
  if (!relay) {
    return (
      <div className="session-card">
        <span className="kicker">DRAWING RELAY</span>
        <h2>Preparing the relay…</h2>
      </div>
    )
  }

  const isActive = relay.status === 'playing' && relay.drawerId === playerId
  const canStart = connected && players.length >= 2
  const timerPercent =
    relay.status === 'playing' && relay.turnDuration
      ? Math.max(0, Math.min(100, (relayTimeLeft / relay.turnDuration) * 100))
      : 0

  if (relay.status === 'lobby') {
    return (
      <div className="relay-panel">
        <div className="session-card relay-intro">
          <span className="kicker">DRAWING RELAY</span>
          <h2>Pass the pen.</h2>
          <p>Every player gets two turns. The prompt appears when it is your turn, and the canvas clears between turns.</p>
          <ul>
            <li>2–4 players per relay.</li>
            <li>Turns last {relay.turnDuration || 60} seconds.</li>
            <li>Only the active player can draw.</li>
          </ul>
        </div>
        <button type="button" className="primary relay-start" onClick={onStartRelay} disabled={!canStart}>
          {canStart ? 'Start the relay' : 'Waiting for another player'}
          <span aria-hidden="true">↗</span>
        </button>
        <p className="relay-note">Late joiners can watch this relay and play in the next one.</p>
      </div>
    )
  }

  if (relay.status === 'finished') {
    return (
      <div className="relay-panel">
        <div className="session-card relay-intro relay-finished">
          <span className="kicker">RELAY COMPLETE</span>
          <h2>Every turn is in the books.</h2>
          <p>
            The relay finished after {relay.totalTurns} {relay.totalTurns === 1 ? 'turn' : 'turns'}. Start a fresh
            relay to include everyone in the room.
          </p>
        </div>
        <button type="button" className="primary relay-start" onClick={onStartRelay} disabled={!canStart}>
          {canStart ? 'Start a new relay' : 'Waiting for another player'}
          <span aria-hidden="true">↗</span>
        </button>
      </div>
    )
  }

  const paused = relay.status === 'paused'

  return (
    <div className="relay-panel">
      <div className={`relay-turn-card ${isActive ? 'relay-turn-active' : ''}`.trim()}>
        <div className="relay-turn-top">
          <span className="kicker">{paused ? 'RELAY PAUSED' : isActive ? 'YOUR TURN' : 'DRAWING NOW'}</span>
          {!paused && <span className="relay-clock">{formatClock(relayTimeLeft)}</span>}
        </div>
        {paused ? (
          <>
            <h2>Waiting for another player.</h2>
            <p>The turn timer is stopped until someone joins.</p>
          </>
        ) : (
          <>
            <h2>{isActive ? 'Draw this:' : `${relay.drawerName || 'A player'} is drawing:`}</h2>
            <strong className="relay-prompt">{relay.prompt}</strong>
            <p>
              {isActive
                ? 'You have the pen. Make it recognizable before time runs out.'
                : 'Watch the canvas and get ready for your turn.'}
            </p>
          </>
        )}
        {!paused && (
          <div className="relay-progress" aria-hidden="true">
            <span style={{ width: `${timerPercent}%` }} />
          </div>
        )}
      </div>

      <div className="relay-meta">
        <div>
          <span>TURN</span>
          <b>
            {relay.turn} / {relay.totalTurns}
          </b>
        </div>
        <div>
          <span>PLAYERS</span>
          <b>{players.length} / 4</b>
        </div>
        <div>
          <span>TURNS EACH</span>
          <b>{relay.turnsPerPlayer || 2}</b>
        </div>
      </div>
    </div>
  )
}
