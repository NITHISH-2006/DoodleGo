import {
  DIFFICULTY_OPTIONS,
  MODE_CLASSIC,
  MODE_TOGETHER,
  RELAY_TURN_OPTIONS,
  ROUND_OPTIONS,
  TOGETHER_ACTIVITY_OPTIONS,
  TOGETHER_RELAY,
} from '../lib/game'

export default function Lobby({
  entryMode,
  setEntryMode,
  name,
  setName,
  roomCode,
  setRoomCode,
  mode,
  setMode,
  difficulty,
  setDifficulty,
  activity,
  setActivity,
  roundDuration,
  setRoundDuration,
  relayTurnDuration,
  setRelayTurnDuration,
  onSubmit,
  error,
  connection,
}) {
  const connecting = connection === 'connecting'

  return (
    <section className="welcome">
      <div className="intro">
        <p className="kicker">REAL-TIME DRAWING ROOM</p>
        <h1>
          Make a mark.
          <br />
          <em>Make them guess.</em>
        </h1>
        <p className="lede">
          A tiny online sketchbook for friends, strangers, and suspiciously good doodlers. Guess a secret word,
          complete a team mission, or pass the pen in a drawing relay.
        </p>
        <div className="intro-chips">
          <span>Live canvas</span>
          <span>Three difficulties</span>
          <span>Team missions</span>
          <span>Drawing relay</span>
        </div>
      </div>

      <div className="join-panel">
        <div className="tab-row">
          <button
            type="button"
            className={entryMode === 'create' ? 'active' : ''}
            onClick={() => setEntryMode('create')}
          >
            Create room
          </button>
          <button
            type="button"
            className={entryMode === 'join' ? 'active' : ''}
            onClick={() => setEntryMode('join')}
          >
            Join room
          </button>
        </div>

        <form onSubmit={onSubmit}>
          <label className="field">
            <span className="field-label">Your name</span>
            <input
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="e.g. Ada"
              maxLength="24"
              autoComplete="nickname"
              autoFocus
              required
            />
          </label>

          {entryMode === 'join' && (
            <label className="field">
              <span className="field-label">Room code</span>
              <input
                value={roomCode}
                onChange={(event) => setRoomCode(event.target.value.toUpperCase())}
                placeholder="ROOM1234"
                maxLength="8"
                autoComplete="off"
                autoCapitalize="characters"
                spellCheck="false"
                required
              />
            </label>
          )}

          {entryMode === 'create' && (
            <>
              <div className="field">
                <span className="field-label">Game mode</span>
                <div className="mode-cards">
                  <button
                    type="button"
                    className={`mode-card ${mode === MODE_CLASSIC ? 'active' : ''}`.trim()}
                    onClick={() => setMode(MODE_CLASSIC)}
                    aria-pressed={mode === MODE_CLASSIC}
                  >
                    <strong>Classic</strong>
                    <span>One draws, everyone guesses the word.</span>
                  </button>
                  <button
                    type="button"
                    className={`mode-card ${mode === MODE_TOGETHER ? 'active' : ''}`.trim()}
                    onClick={() => setMode(MODE_TOGETHER)}
                    aria-pressed={mode === MODE_TOGETHER}
                  >
                    <strong>Doodle Together</strong>
                    <span>Pick a shared activity and draw with everyone.</span>
                  </button>
                </div>
              </div>

              {mode === MODE_CLASSIC && (
                <>
                  <label className="field duration-field">
                    <span className="field-label">Word difficulty</span>
                    <select value={difficulty} onChange={(event) => setDifficulty(event.target.value)}>
                      {DIFFICULTY_OPTIONS.map((option) => (
                        <option value={option.value} key={option.value}>
                          {option.label} — {option.detail}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label className="field duration-field">
                    <span className="field-label">Round length</span>
                    <select value={roundDuration} onChange={(event) => setRoundDuration(Number(event.target.value))}>
                      {ROUND_OPTIONS.map((option) => (
                        <option value={option.value} key={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </select>
                  </label>
                </>
              )}

              {mode === MODE_TOGETHER && (
                <>
                  <div className="field">
                    <span className="field-label">Together activity</span>
                    <div className="mode-cards activity-cards">
                      {TOGETHER_ACTIVITY_OPTIONS.map((option) => (
                        <button
                          type="button"
                          key={option.value}
                          className={`mode-card ${activity === option.value ? 'active' : ''}`.trim()}
                          onClick={() => setActivity(option.value)}
                          aria-pressed={activity === option.value}
                        >
                          <strong>{option.label}</strong>
                          <span>{option.detail}</span>
                        </button>
                      ))}
                    </div>
                  </div>
                  {activity === TOGETHER_RELAY && (
                    <label className="field duration-field">
                      <span className="field-label">Turn length</span>
                      <select
                        value={relayTurnDuration}
                        onChange={(event) => setRelayTurnDuration(Number(event.target.value))}
                      >
                        {RELAY_TURN_OPTIONS.map((option) => (
                          <option value={option.value} key={option.value}>
                            {option.label}
                          </option>
                        ))}
                      </select>
                    </label>
                  )}
                </>
              )}
            </>
          )}

          <button className="primary" type="submit" disabled={connecting}>
            {connecting ? 'Connecting…' : entryMode === 'create' ? 'Create the room' : 'Join the room'}
            <span aria-hidden="true">↗</span>
          </button>
        </form>

        {error && <p className="error" role="alert">{error}</p>}

        <small className="server-note">
          {connection === 'connected'
            ? 'Connected to the game server.'
            : connecting
              ? 'Connecting to the game server…'
              : 'Real-time WebSocket connection.'}
        </small>
      </div>
    </section>
  )
}
