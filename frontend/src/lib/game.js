export const PAPER = '#fffdf8'
export const INK = '#162521'
export const CANVAS_WIDTH = 900
export const CANVAS_HEIGHT = 560

export const COLORS = [
  '#162521',
  '#d64f3b',
  '#ff795f',
  '#f2a541',
  '#f7d154',
  '#c8ef6c',
  '#57a340',
  '#3aa6a6',
  '#2b6cb0',
  '#6b46c1',
  '#d53f8c',
  '#8b5a2b',
]

export const ROUND_OPTIONS = [
  { value: 30, label: '30 seconds' },
  { value: 60, label: '60 seconds' },
  { value: 90, label: '90 seconds' },
  { value: 120, label: '2 minutes' },
  { value: 180, label: '3 minutes' },
]

export const MODE_CLASSIC = 'classic'
export const MODE_TOGETHER = 'together'

export const DIFFICULTY_EASY = 'easy'
export const DIFFICULTY_MEDIUM = 'medium'
export const DIFFICULTY_HARD = 'hard'

export const DIFFICULTY_OPTIONS = [
  { value: DIFFICULTY_EASY, label: 'Easy', detail: 'Simple everyday things' },
  { value: DIFFICULTY_MEDIUM, label: 'Medium', detail: 'More detail and trickier words' },
  { value: DIFFICULTY_HARD, label: 'Hard', detail: 'Complex and surprising words' },
]

export const TOGETHER_FREE_DRAW = 'free_draw'
export const TOGETHER_MISSIONS = 'missions'
export const TOGETHER_RELAY = 'relay'

export const TOGETHER_ACTIVITY_OPTIONS = [
  { value: TOGETHER_FREE_DRAW, label: 'Free Draw', detail: 'One shared canvas with no rules.' },
  { value: TOGETHER_MISSIONS, label: 'Team Missions', detail: 'Complete a shared drawing checklist.' },
  { value: TOGETHER_RELAY, label: 'Drawing Relay', detail: 'Take turns drawing a surprise prompt.' },
]

export const RELAY_TURN_OPTIONS = [
  { value: 30, label: '30 seconds' },
  { value: 60, label: '60 seconds' },
  { value: 90, label: '90 seconds' },
]

export const RELAY_TURNS_PER_PLAYER = 2

export const TOOLS = {
  pen: { label: 'Pen', opacity: 1 },
  marker: { label: 'Marker', opacity: 0.35 },
  eraser: { label: 'Eraser', opacity: 1 },
}

export function sortPlayers(players) {
  return [...players].sort((a, b) => {
    if (b.score !== a.score) return b.score - a.score
    return String(a.name).localeCompare(String(b.name))
  })
}

export function formatClock(seconds) {
  const safe = Math.max(0, Math.floor(seconds || 0))
  const minutes = Math.floor(safe / 60)
  const rest = safe % 60
  return `${minutes}:${String(rest).padStart(2, '0')}`
}

export function difficultyLabel(value) {
  return DIFFICULTY_OPTIONS.find((option) => option.value === value)?.label || 'Easy'
}

export function togetherActivityLabel(value) {
  return TOGETHER_ACTIVITY_OPTIONS.find((option) => option.value === value)?.label || 'Free Draw'
}
