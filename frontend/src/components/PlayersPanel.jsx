import { sortPlayers } from '../lib/game'

export default function PlayersPanel({ players = [], playerId, maxPlayers = 20, showScores = true, roomCode, mode }) {
  const sorted = sortPlayers(players)

  return (
    <aside className="side-panel">
      <div className="panel-heading">
        <span>PLAYERS</span>
        <b className="panel-count">{players.length}/{maxPlayers}</b>
      </div>

      <div className="players">
        {sorted.map((player) => (
          <div className={`player ${player.id === playerId ? 'player-self' : ''}`.trim()} key={player.id}>
            <span className="avatar">{String(player.name || '?').slice(0, 1).toUpperCase()}</span>
            <span className="player-name">
              {player.name}
              {player.id === playerId && <small> (you)</small>}
            </span>
            {showScores && <strong className="player-score">{player.score}</strong>}
          </div>
        ))}
        {players.length === 0 && <p className="empty-note">Nobody here yet.</p>}
      </div>

      <div className="room-note">
        <span>{mode === 'together' ? 'DRAWING ROOM' : 'ROOM CODE'}</span>
        <b className="room-note-code">{roomCode}</b>
        <small>{mode === 'together' ? 'Share this code to doodle together.' : 'Share this code with your friends.'}</small>
      </div>
    </aside>
  )
}
