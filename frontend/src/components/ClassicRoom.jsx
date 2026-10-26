import DrawingCanvas from './DrawingCanvas'
import GuessPanel from './GuessPanel'
import PlayersPanel from './PlayersPanel'
import Toolbar from './Toolbar'
import { useDrawingTools } from '../hooks/useDrawingTools'
import { formatClock } from '../lib/game'

export default function ClassicRoom({
  room,
  game,
  playerId,
  timeLeft,
  messages,
  guess,
  setGuess,
  onSubmitGuess,
  onStart,
  connected,
  roundResult,
  onStroke,
  canvasRef,
  correctGuessers,
  onUndo,
  onClear,
}) {
  const tools = useDrawingTools()
  const players = room.players || []
  const isDrawer = Boolean(game && game.drawerId === playerId)
  const solved = correctGuessers.includes(playerId)
  const canStart = !game && players.length >= 2
  const guessDisabled = !game || isDrawer || solved || !connected
  const duration = Math.max(1, room.roundDuration || 90)
  const timerPercent = game ? Math.max(0, Math.min(100, (timeLeft / duration) * 100)) : 0
  const wordMask = game?.maskedWord || (game?.wordLength ? '_'.repeat(game.wordLength) : '')

  let placeholder = 'Type a guess…'
  if (!connected) placeholder = 'Reconnecting…'
  else if (isDrawer) placeholder = 'You are drawing'
  else if (solved) placeholder = 'You guessed it!'
  else if (!game) placeholder = 'Waiting for a round…'

  return (
    <section className="room-layout">
      <PlayersPanel
        players={players}
        playerId={playerId}
        maxPlayers={room.maxPlayers}
        showScores
        roomCode={room.roomCode}
        mode={room.mode}
      />

      <div className="board-column">
        <div className="game-status">
          <span className={`status-dot ${game ? 'status-dot-live' : 'status-dot-wait'}`.trim()} />
          {game ? (
            isDrawer ? (
              <span className="word-display">
                Your word is <strong>{game.word}</strong>
              </span>
            ) : (
              <span className="word-display">
                Guess the word <span className="word-mask">{wordMask.split('').join(' ')}</span>
                <small>{game.wordLength} letters</small>
              </span>
            )
          ) : (
            <span>{players.length >= 2 ? 'Ready for the next round' : 'Waiting for another player to join'}</span>
          )}
          <span className="round-label">
            {game ? `ROUND ${game.roundNum} · ${formatClock(timeLeft)}` : 'LOBBY'}
          </span>
        </div>

        {game && (
          <div className="timer-track" aria-hidden="true">
            <span className="timer-fill" style={{ width: `${timerPercent}%` }} />
          </div>
        )}

        <DrawingCanvas
          ref={canvasRef}
          enabled={isDrawer}
          tool={tools.tool}
          color={tools.color}
          size={tools.size}
          onStroke={onStroke}
          label={isDrawer ? 'Your drawing canvas' : 'Live drawing canvas'}
        />

        {isDrawer && (
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
          />
        )}

        {game && !isDrawer && (
          <p className="draw-note">
            <b>{game.drawerName}</b> is drawing. Watch the canvas and type your guess.
          </p>
        )}

        {roundResult && !game && (
          <div className="round-banner" role="status">
            <span>ROUND OVER</span>
            <strong>{roundResult.word}</strong>
            <small>
              {players.length >= 2 ? 'The next round starts automatically in a few seconds.' : 'Invite another player to keep playing.'}
            </small>
          </div>
        )}

        {!game && (
          <div className="lobby-actions">
            {canStart ? (
              <button type="button" className="primary start-button" onClick={onStart}>
                Start round <span aria-hidden="true">↗</span>
              </button>
            ) : (
              <p className="lobby-hint">At least 2 players are needed to start. Share the room code above.</p>
            )}
          </div>
        )}
      </div>

      <GuessPanel
        messages={messages}
        guess={guess}
        setGuess={setGuess}
        onSubmit={onSubmitGuess}
        disabled={guessDisabled}
        placeholder={placeholder}
        connected={connected}
      />
    </section>
  )
}
