import { useState } from 'react'
import { COLORS, INK } from '../lib/game'

export function useDrawingTools() {
  const [tool, setTool] = useState('pen')
  const [color, setColor] = useState(INK)
  const [size, setSize] = useState(6)

  const chooseTool = (nextTool) => {
    setTool(nextTool)
    if (nextTool === 'marker') {
      setSize((current) => (current < 12 ? 16 : current))
    }
    if (nextTool === 'pen' && size > 24) {
      setSize(6)
    }
  }

  const chooseColor = (nextColor) => {
    setColor(nextColor)
    if (tool === 'eraser') setTool('pen')
  }

  return {
    tool,
    setTool: chooseTool,
    color,
    setColor: chooseColor,
    size,
    setSize,
    colors: COLORS,
    isEraser: tool === 'eraser',
  }
}
