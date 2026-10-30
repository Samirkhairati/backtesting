// Decorative, purely illustrative equity curve (no real data)
const POINTS: [number, number][] = [
  [0, 150], [20, 144], [40, 148], [60, 132], [80, 138], [100, 120], [120, 126],
  [140, 108], [160, 114], [180, 96], [200, 102], [220, 84], [240, 92], [260, 70],
  [280, 76], [300, 58], [320, 64], [340, 44], [360, 50], [380, 30], [400, 22],
]

const LINE = POINTS.map(([x, y], i) => `${i === 0 ? 'M' : 'L'}${x} ${y}`).join(' ')
const AREA = `${LINE} L400 170 L0 170 Z`

export function EquityCurve({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 400 170"
      role="img"
      aria-label="Illustrative equity curve trending upward"
      className={className}
      preserveAspectRatio="none"
    >
      <defs>
        <linearGradient id="equity-fill" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="hsl(var(--primary))" stopOpacity="0.35" />
          <stop offset="100%" stopColor="hsl(var(--primary))" stopOpacity="0" />
        </linearGradient>
      </defs>
      {[34, 68, 102, 136].map((y) => (
        <line key={y} x1="0" x2="400" y1={y} y2={y} stroke="hsl(var(--border))" strokeOpacity="0.5" strokeDasharray="2 6" />
      ))}
      <path d={AREA} fill="url(#equity-fill)" />
      <path
        d={LINE}
        fill="none"
        stroke="hsl(var(--primary))"
        strokeWidth="2"
        strokeLinejoin="round"
        strokeLinecap="round"
        vectorEffect="non-scaling-stroke"
      />
      <circle cx="400" cy="22" r="3.5" fill="hsl(var(--primary))" />
    </svg>
  )
}
