// EloSparkline — tiny inline trend of a deck's ending HexELO across the
// last N gauntlet runs. Lives in the vital-signs strip alongside the big
// HexELO number on the deck archive page. Oldest left → newest right.
// Stroke color tracks net delta from first → last run, so a glance tells
// you trajectory before you read the dot. Returns null with fewer than
// two runs — the cell falls back to its placeholder sublabel.
export default function EloSparkline({ runs, width = 80, height = 22 }) {
  if (!Array.isArray(runs) || runs.length < 2) return null
  const ends = runs.map(r => Number(r?.elo_end) || 0)
  const minY = Math.min(...ends)
  const maxY = Math.max(...ends)
  const span = maxY - minY || 1
  const padY = 2
  const plotH = height - padY * 2
  const stepX = width / (ends.length - 1)
  const yAt = (v) => padY + plotH - ((v - minY) / span) * plotH
  const points = ends.map((v, i) => `${(i * stepX).toFixed(1)},${yAt(v).toFixed(1)}`)
  const path = `M ${points.join(' L ')}`
  const netDelta = ends[ends.length - 1] - ends[0]
  const stroke = netDelta > 0 ? 'var(--ok)' : netDelta < 0 ? 'var(--danger)' : 'var(--ink-2)'
  const lastX = (ends.length - 1) * stepX
  const lastY = yAt(ends[ends.length - 1])
  const tip = `Last ${runs.length} runs: ${Math.round(ends[0])} → ${Math.round(ends[ends.length - 1])} (${netDelta >= 0 ? '+' : ''}${Math.round(netDelta)})`
  return (
    <svg className="elo-sparkline"
         viewBox={`0 0 ${width} ${height}`}
         width={width} height={height}
         preserveAspectRatio="none"
         role="img" aria-label={tip}>
      <title>{tip}</title>
      <path d={path} fill="none" stroke={stroke} strokeWidth="1.25"
            strokeLinejoin="round" strokeLinecap="round" />
      <circle cx={lastX} cy={lastY} r="1.6" fill={stroke} />
    </svg>
  )
}
