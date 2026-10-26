import { PAPER } from './game'

export function clamp(value, min, max) {
  return Math.min(max, Math.max(min, value))
}

export function clearSurface(ctx, width, height) {
  ctx.save()
  ctx.globalAlpha = 1
  ctx.globalCompositeOperation = 'source-over'
  ctx.fillStyle = PAPER
  ctx.fillRect(0, 0, width, height)
  ctx.restore()
}

export function renderStroke(ctx, stroke) {
  const rawPoints = stroke?.points
  if (!ctx || !Array.isArray(rawPoints) || rawPoints.length === 0) return

  const points = rawPoints.filter((point) => Number.isFinite(point?.x) && Number.isFinite(point?.y))
  if (points.length === 0) return

  const isEraser = stroke.tool === 'eraser'
  const color = isEraser ? PAPER : stroke.color || '#162521'
  const size = clamp(Number(stroke.size) || 6, 1, 64)
  const opacity = clamp(
    typeof stroke.opacity === 'number' && stroke.opacity > 0
      ? stroke.opacity
      : stroke.tool === 'marker'
        ? 0.35
        : 1,
    0.05,
    1,
  )

  ctx.save()
  ctx.globalAlpha = opacity
  ctx.globalCompositeOperation = 'source-over'
  ctx.strokeStyle = color
  ctx.fillStyle = color
  ctx.lineWidth = size
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'

  if (points.length === 1) {
    ctx.beginPath()
    ctx.arc(points[0].x, points[0].y, Math.max(0.5, size / 2), 0, Math.PI * 2)
    ctx.fill()
  } else {
    ctx.beginPath()
    ctx.moveTo(points[0].x, points[0].y)
    for (let index = 1; index < points.length; index += 1) {
      ctx.lineTo(points[index].x, points[index].y)
    }
    ctx.stroke()
  }
  ctx.restore()
}
