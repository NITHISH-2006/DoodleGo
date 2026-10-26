export default function MissionPanel({ mission, connected, onToggleTask, onNextMission }) {
  if (!mission?.mission) {
    return (
      <div className="session-card">
        <span className="kicker">TEAM MISSION</span>
        <h2>Preparing the first mission…</h2>
      </div>
    )
  }

  const tasks = mission.mission.tasks || []
  const checked = mission.checked || []
  const completedCount = checked.filter(Boolean).length
  const allChecked = tasks.length > 0 && completedCount === tasks.length

  return (
    <div className="mission-panel">
      <div className="session-card mission-intro">
        <div className="mission-kicker-row">
          <span className="kicker">TEAM MISSION</span>
          <span className="mission-number">MISSION {mission.number}</span>
        </div>
        <h2>{mission.mission.title}</h2>
        <p>{mission.mission.brief}</p>
        <div className="mission-progress">
          <span>
            {completedCount} / {tasks.length} drawn
          </span>
          <span>{mission.completed} completed</span>
        </div>
      </div>

      <div className="mission-list" aria-label="Mission drawing checklist">
        {tasks.map((task, index) => {
          const isChecked = Boolean(checked[index])
          return (
            <button
              key={`${mission.mission.id}-${task}`}
              type="button"
              className={`mission-task ${isChecked ? 'done' : ''}`.trim()}
              onClick={() => onToggleTask(index)}
              disabled={!connected}
              aria-pressed={isChecked}
            >
              <span className="mission-check" aria-hidden="true">
                {isChecked ? '✓' : ''}
              </span>
              <span>{task}</span>
            </button>
          )
        })}
      </div>

      <button
        type="button"
        className="primary mission-next"
        onClick={onNextMission}
        disabled={!connected || !allChecked}
      >
        {allChecked ? 'Next mission' : 'Check every task to continue'}
        <span aria-hidden="true">↗</span>
      </button>
      <p className="mission-note">Anyone in the room can check off a task. No timer, no pressure.</p>
    </div>
  )
}
