import { useEffect, useRef } from 'react'

export default function GuessPanel({ messages = [], guess, setGuess, onSubmit, disabled, placeholder, connected }) {
  const logRef = useRef(null)

  useEffect(() => {
    const log = logRef.current
    if (log) log.scrollTop = log.scrollHeight
  }, [messages])

  return (
    <aside className="chat-panel">
      <div className="panel-heading">
        <span>GUESSES</span>
        <span className={`conn-label ${connected ? 'live' : 'offline'}`.trim()}>
          <i />
          {connected ? 'LIVE' : 'OFFLINE'}
        </span>
      </div>

      <div className="chat-log" ref={logRef}>
        {messages.length === 0 && (
          <p className="chat-empty">Guesses will appear here. The secret word is only visible to the person drawing.</p>
        )}
        {messages.map((message) => {
          if (message.system) {
            return (
              <p className="chat-message chat-system" key={message.id}>
                {message.text}
              </p>
            )
          }
          return (
            <p className={`chat-message ${message.correct ? 'chat-correct' : ''}`.trim()} key={message.id}>
              <b>{message.guesserName}</b>{' '}
              {message.correct ? 'got it!' : message.text}
            </p>
          )
        })}
      </div>

      <form className="guess-form" onSubmit={onSubmit}>
        <input
          className="guess-input"
          value={guess}
          onChange={(event) => setGuess(event.target.value)}
          placeholder={placeholder}
          disabled={disabled}
          maxLength="80"
          autoComplete="off"
          autoCapitalize="none"
          aria-label="Your guess"
        />
        <button type="submit" className="guess-send" disabled={disabled} aria-label="Send guess">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="M4 12l16-8-5 16-3-6-8-2z" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinejoin="round" />
          </svg>
        </button>
      </form>
    </aside>
  )
}
