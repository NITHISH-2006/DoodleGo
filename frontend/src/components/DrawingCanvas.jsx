import { forwardRef, useImperativeHandle, useLayoutEffect, useRef } from 'react'
import { CANVAS_HEIGHT, CANVAS_WIDTH, PAPER } from '../lib/game'
import { clamp, clearSurface, renderStroke } from '../lib/drawing'

const DrawingCanvas = forwardRef(function DrawingCanvas(
  { enabled = false, tool = 'pen', color = '#162521', size = 6, onStroke, className = '', label = 'Drawing canvas' },
  ref,
) {
  const canvasRef = useRef(null)
  const ctxRef = useRef(null)
  const historyRef = useRef([])
  const gestureRef = useRef(null)

  const redraw = () => {
    const canvas = canvasRef.current
    const ctx = ctxRef.current
    if (!canvas || !ctx) return
    clearSurface(ctx, canvas.width, canvas.height)
    for (const stroke of historyRef.current) renderStroke(ctx, stroke)
  }

  useLayoutEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return undefined
    const ctx = canvas.getContext('2d')
    ctxRef.current = ctx
    clearSurface(ctx, canvas.width, canvas.height)
    for (const stroke of historyRef.current) renderStroke(ctx, stroke)
    return () => {
      ctxRef.current = null
      gestureRef.current = null
    }
  }, [])

  const pointFromEvent = (event) => {
    const canvas = canvasRef.current
    if (!canvas) return { x: 0, y: 0 }
    const bounds = canvas.getBoundingClientRect()
    const x = ((event.clientX - bounds.left) / Math.max(1, bounds.width)) * canvas.width
    const y = ((event.clientY - bounds.top) / Math.max(1, bounds.height)) * canvas.height
    return {
      x: Math.round(clamp(x, 0, canvas.width) * 100) / 100,
      y: Math.round(clamp(y, 0, canvas.height) * 100) / 100,
    }
  }

  const flushGesture = (final) => {
    const gesture = gestureRef.current
    const ctx = ctxRef.current
    if (!gesture || !ctx) return

    const now = performance.now()
    if (!final && now - gesture.lastFlush < 45) return

    const points = gesture.points
    if (points.length === 0) return
    if (points.length === 1 && !final) return
    if (points.length === 1 && final && gesture.sentAny) {
      gesture.points = []
      return
    }

    const chunk = {
      strokeId: gesture.strokeId,
      points: points.slice(),
      color: gesture.color,
      size: gesture.size,
      tool: gesture.tool,
      opacity: gesture.opacity,
    }
    historyRef.current.push(chunk)
    onStroke?.(chunk)

    gesture.lastFlush = now
    gesture.sentAny = true
    if (final) {
      gesture.points = []
    } else {
      gesture.points = [points[points.length - 1]]
    }
  }

  const beginGesture = (event) => {
    if (!enabled || !canvasRef.current || !ctxRef.current) return
    if (event.pointerType === 'mouse' && event.button !== 0) return
    event.preventDefault()

    const canvas = canvasRef.current
    try {
      canvas.setPointerCapture(event.pointerId)
    } catch {
      // Pointer capture is not available in every browser; drawing still works.
    }

    const point = pointFromEvent(event)
    const activatedTool = tool || 'pen'
    const activatedColor = activatedTool === 'eraser' ? PAPER : color
    const activatedOpacity = activatedTool === 'marker' ? 0.35 : 1

    const gesture = {
      pointerId: event.pointerId,
      strokeId: `g-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`,
      tool: activatedTool,
      color: activatedColor,
      size,
      opacity: activatedOpacity,
      points: [point],
      lastPoint: point,
      lastFlush: performance.now(),
      sentAny: false,
    }

    gestureRef.current = gesture
    renderStroke(ctxRef.current, { ...gesture, points: [point] })
  }

  const moveGesture = (event) => {
    const gesture = gestureRef.current
    const ctx = ctxRef.current
    if (!gesture || !ctx || event.pointerId !== gesture.pointerId) return
    event.preventDefault()

    const nativeEvent = event.nativeEvent
    const coalesced =
      nativeEvent && typeof nativeEvent.getCoalescedEvents === 'function'
        ? nativeEvent.getCoalescedEvents()
        : [nativeEvent]
    const samples = coalesced && coalesced.length > 0 ? coalesced : [nativeEvent]

    let changed = false
    for (const sample of samples) {
      const point = pointFromEvent(sample)
      const last = gesture.lastPoint
      if (!last) continue
      if (Math.abs(point.x - last.x) < 0.15 && Math.abs(point.y - last.y) < 0.15) continue

      renderStroke(ctx, { ...gesture, points: [last, point] })
      gesture.points.push(point)
      gesture.lastPoint = point
      changed = true
    }

    if (changed) flushGesture(false)
  }

  const endGesture = (event) => {
    const gesture = gestureRef.current
    if (!gesture || event.pointerId !== gesture.pointerId) return

    flushGesture(true)
    gestureRef.current = null

    try {
      canvasRef.current?.releasePointerCapture(event.pointerId)
    } catch {
      // Ignore release errors from already-cancelled pointers.
    }
  }

  const applyStroke = (stroke) => {
    if (!stroke || !Array.isArray(stroke.points) || stroke.points.length === 0) return
    historyRef.current.push(stroke)
    renderStroke(ctxRef.current, stroke)
  }

  const setStrokes = (strokes) => {
    historyRef.current = Array.isArray(strokes) ? strokes.slice() : []
    redraw()
  }

  const clearCanvas = () => {
    historyRef.current = []
    const canvas = canvasRef.current
    const ctx = ctxRef.current
    if (canvas && ctx) clearSurface(ctx, canvas.width, canvas.height)
  }

  const download = (filename) => {
    const canvas = canvasRef.current
    if (!canvas) return
    const output = document.createElement('canvas')
    output.width = canvas.width
    output.height = canvas.height
    const ctx = output.getContext('2d')
    ctx.fillStyle = PAPER
    ctx.fillRect(0, 0, output.width, output.height)
    ctx.drawImage(canvas, 0, 0)

    const link = document.createElement('a')
    link.download = filename || `doodlego-${Date.now()}.png`
    link.href = output.toDataURL('image/png')
    document.body.appendChild(link)
    link.click()
    link.remove()
  }

  useImperativeHandle(ref, () => ({
    applyStroke,
    setStrokes,
    clearCanvas,
    download,
  }))

  return (
    <div className={`canvas-wrap ${enabled ? 'canvas-enabled' : 'canvas-locked'} ${className}`.trim()}>
      <canvas
        ref={canvasRef}
        width={CANVAS_WIDTH}
        height={CANVAS_HEIGHT}
        aria-label={label}
        onPointerDown={beginGesture}
        onPointerMove={moveGesture}
        onPointerUp={endGesture}
        onPointerCancel={endGesture}
        onContextMenu={(event) => event.preventDefault()}
      />
      {!enabled && <span className="canvas-lock-label">Canvas is read-only</span>}
    </div>
  )
})

export default DrawingCanvas
