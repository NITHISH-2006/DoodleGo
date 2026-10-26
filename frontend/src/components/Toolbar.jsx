function PenIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M4 20l3.5-1 11-11-2.5-2.5-11 11L4 20z" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinejoin="round" />
      <path d="M14 6.5L17.5 10" fill="none" stroke="currentColor" strokeWidth="1.7" />
    </svg>
  )
}

function MarkerIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M7 15l-1 5 5-1 8-8-4-4-8 8z" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinejoin="round" />
      <path d="M14 5l3-3 5 5-3 3" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinejoin="round" />
    </svg>
  )
}

function EraserIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M4 15l7-7 7 7-4 4H8l-4-4z" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinejoin="round" />
      <path d="M9 20h11" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" />
    </svg>
  )
}

function UndoIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M9 7H4v5" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M4 12c2-5 7-7 11-5 3 1.5 5 4.5 5 8" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" />
    </svg>
  )
}

function ClearIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M5 7h14M9 7V4h6v3M8 7l1 13h6l1-13" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function DownloadIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M12 4v11m0 0l-4-4m4 4l4-4M5 19h14" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

const TOOL_BUTTONS = [
  { id: 'pen', label: 'Pen', icon: <PenIcon /> },
  { id: 'marker', label: 'Marker', icon: <MarkerIcon /> },
  { id: 'eraser', label: 'Eraser', icon: <EraserIcon /> },
]

export default function Toolbar({
  tool,
  setTool,
  color,
  setColor,
  size,
  setSize,
  colors = [],
  onUndo,
  onClear,
  onDownload,
  disabled = false,
}) {
  return (
    <div className={`toolbar ${disabled ? 'toolbar-disabled' : ''}`.trim()} role="toolbar" aria-label="Drawing tools">
      <div className="toolbar-group">
        <span className="toolbar-label">Tool</span>
        <div className="tool-buttons">
          {TOOL_BUTTONS.map((item) => (
            <button
              key={item.id}
              type="button"
              className={`tool-button ${tool === item.id ? 'active' : ''}`.trim()}
              onClick={() => setTool(item.id)}
              disabled={disabled}
              aria-pressed={tool === item.id}
              title={item.label}
            >
              {item.icon}
              <span>{item.label}</span>
            </button>
          ))}
        </div>
      </div>

      <div className="toolbar-group">
        <span className="toolbar-label">Color</span>
        <div className="color-grid">
          {colors.map((swatchColor) => (
            <button
              key={swatchColor}
              type="button"
              className={`swatch ${color.toLowerCase() === swatchColor.toLowerCase() ? 'active' : ''}`.trim()}
              style={{ backgroundColor: swatchColor }}
              onClick={() => setColor(swatchColor)}
              disabled={disabled}
              aria-label={`Use color ${swatchColor}`}
              aria-pressed={color.toLowerCase() === swatchColor.toLowerCase()}
            />
          ))}
          <label className="custom-color" title="Custom color">
            <input
              type="color"
              value={color}
              onChange={(event) => setColor(event.target.value)}
              disabled={disabled}
              aria-label="Pick a custom color"
            />
          </label>
        </div>
      </div>

      <div className="toolbar-group size-group">
        <span className="toolbar-label">
          Brush <b>{size}px</b>
        </span>
        <input
          type="range"
          min="2"
          max="36"
          step="1"
          value={size}
          onChange={(event) => setSize(Number(event.target.value))}
          disabled={disabled}
          aria-label="Brush size"
        />
      </div>

      <div className="toolbar-group toolbar-actions">
        <span className="toolbar-label">Actions</span>
        <div className="tool-buttons">
          <button type="button" className="tool-button" onClick={onUndo} disabled={disabled || !onUndo} title="Undo last stroke">
            <UndoIcon />
            <span>Undo</span>
          </button>
          <button type="button" className="tool-button danger" onClick={onClear} disabled={disabled || !onClear} title="Clear the whole canvas">
            <ClearIcon />
            <span>Clear</span>
          </button>
          {onDownload && (
            <button type="button" className="tool-button" onClick={onDownload} disabled={disabled} title="Download this drawing as PNG">
              <DownloadIcon />
              <span>Download</span>
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
