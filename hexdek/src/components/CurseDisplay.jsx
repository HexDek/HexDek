import { Panel, KV, Tag } from './chrome'

// 7 personality params, in radar-axis order. Order is the visual angle
// around the chart, starting at 12 o'clock. Keys match the JSON shape
// returned by /api/decks/:owner/:id/curse.
const TRAITS = [
  { key: 'aggression',         label: 'AGGRESSION' },
  { key: 'combo_patience',     label: 'COMBO PAT.' },
  { key: 'threat_paranoia',    label: 'THREAT' },
  { key: 'resource_greed',     label: 'GREED' },
  { key: 'political_memory',   label: 'POLITICS' },
  { key: 'drain_affinity',     label: 'DRAIN' },
  { key: 'artifact_affinity',  label: 'ARTIFACT' },
]

function RadarChart({ values, size = 220 }) {
  const cx = size / 2
  const cy = size / 2
  const r = size * 0.36
  const n = TRAITS.length
  const angle = (i) => -Math.PI / 2 + (2 * Math.PI * i) / n
  const point = (i, v) => {
    const a = angle(i)
    const rr = r * Math.max(0, Math.min(1, v))
    return [cx + Math.cos(a) * rr, cy + Math.sin(a) * rr]
  }
  const axisPoint = (i, frac = 1) => {
    const a = angle(i)
    return [cx + Math.cos(a) * r * frac, cy + Math.sin(a) * r * frac]
  }
  const polygon = values.map((v, i) => point(i, v).join(',')).join(' ')
  const rings = [0.25, 0.5, 0.75, 1.0]

  return (
    <div style={{ maxWidth: size, margin: '0 auto' }}>
    <svg viewBox={`0 0 ${size} ${size}`} width="100%" style={{ display: 'block', overflow: 'visible' }}>
      {rings.map((f, i) => (
        <polygon
          key={i}
          points={Array.from({ length: n }, (_, j) => axisPoint(j, f).join(',')).join(' ')}
          fill="none"
          stroke="var(--rule)"
          strokeOpacity={f === 1 ? 0.6 : 0.25}
          strokeWidth="1"
        />
      ))}
      {Array.from({ length: n }, (_, i) => {
        const [x, y] = axisPoint(i, 1)
        return <line key={i} x1={cx} y1={cy} x2={x} y2={y} stroke="var(--rule)" strokeOpacity="0.25" strokeWidth="1" />
      })}
      <polygon
        points={polygon}
        fill="var(--ok)"
        fillOpacity="0.18"
        stroke="var(--ok)"
        strokeWidth="1.5"
      />
      {values.map((v, i) => {
        const [px, py] = point(i, v)
        return <circle key={i} cx={px} cy={py} r="2.5" fill="var(--ok)" />
      })}
      {TRAITS.map((t, i) => {
        const [lx, ly] = axisPoint(i, 1.18)
        const anchor = Math.abs(lx - cx) < 4 ? 'middle' : lx < cx ? 'end' : 'start'
        return (
          <text
            key={t.key}
            x={lx}
            y={ly}
            textAnchor={anchor}
            dominantBaseline="middle"
            fontSize="8"
            fill="var(--ink-2)"
            style={{ letterSpacing: '0.04em', textTransform: 'uppercase' }}
          >
            {t.label}
          </text>
        )
      })}
    </svg>
    </div>
  )
}

// FitnessSparkline plots best-fitness per generation across the last
// `window` generations. Values around 1.0 are par; >1 is above-par.
function FitnessSparkline({ values, width = 320, height = 56 }) {
  if (!values || values.length === 0) return null
  const n = values.length
  const max = Math.max(1.0, ...values)
  const min = Math.min(0, ...values)
  const range = max - min || 1
  const xStep = n > 1 ? (width - 8) / (n - 1) : 0
  const y = (v) => height - 4 - ((v - min) / range) * (height - 8)
  const x = (i) => 4 + i * xStep
  const path = values
    .map((v, i) => `${i === 0 ? 'M' : 'L'} ${x(i).toFixed(1)} ${y(v).toFixed(1)}`)
    .join(' ')
  const baselineY = y(1.0)
  return (
    <svg
      viewBox={`0 0 ${width} ${height}`}
      preserveAspectRatio="none"
      style={{ display: 'block', width: '100%', height }}
    >
      <line x1="0" y1={baselineY} x2={width} y2={baselineY} stroke="var(--rule-2)" strokeDasharray="3 3" strokeWidth="1" />
      <path d={path} fill="none" stroke="var(--ok)" strokeWidth="1.5" vectorEffect="non-scaling-stroke" />
      {values.map((v, i) => (
        <circle key={i} cx={x(i)} cy={y(v)} r="2" fill={v >= 1.0 ? 'var(--ok)' : 'var(--ink-2)'} vectorEffect="non-scaling-stroke" />
      ))}
    </svg>
  )
}

// DimHeatmap renders the 20 weight corrections as a 5×4 grid of cells
// tinted by deviation from 1.0. >1 (boosted) → green; <1 (suppressed)
// → red; ≈1 → neutral. Cold (no observations yet) renders muted with
// "—" labels so the panel reads honestly until data accumulates.
function DimHeatmap({ corrections, labels, n }) {
  if (!corrections || corrections.length === 0) return null
  const cold = !n || n < 20
  const cells = corrections.map((v, i) => ({
    label: labels?.[i] || `D${i}`,
    value: v,
    delta: v - 1.0,
  }))
  const cols = 5
  return (
    <div>
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: `repeat(${cols}, 1fr)`,
          gap: 3,
        }}
      >
        {cells.map((c, i) => {
          const intensity = Math.min(1, Math.abs(c.delta) / 0.2) // ±20% saturates
          const colorVar = c.delta > 0 ? 'var(--ok)' : 'var(--danger)'
          const bg = cold
            ? 'transparent'
            : `color-mix(in srgb, ${colorVar} ${Math.round(intensity * 70)}%, transparent)`
          const sign = c.delta >= 0 ? '+' : ''
          const pct = `${sign}${Math.round(c.delta * 100)}`
          return (
            <div
              key={i}
              title={`${c.label}: ${c.value.toFixed(3)} (${sign}${(c.delta * 100).toFixed(1)}%)`}
              style={{
                border: '1px solid var(--rule-2)',
                background: bg,
                padding: '4px 5px',
                minHeight: 38,
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'space-between',
                opacity: cold ? 0.45 : 1,
              }}
            >
              <div
                style={{
                  fontSize: 7,
                  fontWeight: 700,
                  letterSpacing: '0.04em',
                  color: 'var(--ink-2)',
                  textTransform: 'uppercase',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                }}
              >
                {c.label}
              </div>
              <div
                style={{
                  fontSize: 10,
                  fontWeight: 800,
                  letterSpacing: '0.02em',
                  color: cold ? 'var(--ink-3)' : 'var(--ink)',
                  textAlign: 'right',
                  fontVariantNumeric: 'tabular-nums',
                }}
              >
                {cold ? '—' : pct}
              </div>
            </div>
          )
        })}
      </div>
      <div className="t-xs muted-2" style={{ marginTop: 4, lineHeight: 1.4 }}>
        {cold
          ? `COLD — ${n || 0}/20 GAMES OBSERVED. CORRECTIONS ACTIVATE AT 20.`
          : 'PERCENT DEVIATION FROM BASELINE WEIGHT (±20% MAX). GREEN = WIN-PREDICTIVE, RED = LOSS-PREDICTIVE.'}
      </div>
    </div>
  )
}

export default function CurseDisplay({ curse }) {
  if (!curse || !curse.population || curse.population.length === 0) return null

  const pop = curse.population
  const sorted = [...pop].sort((a, b) => b.fitness - a.fitness)
  const top = sorted[0]
  const maxGen = pop.reduce((m, d) => Math.max(m, d.generation || 0), 0)
  const bestFitness = top.fitness
  const avgFitness = pop.reduce((s, d) => s + d.fitness, 0) / pop.length

  // Best fitness per generation, last 20 generations. Each member
  // carries the gen it was created in; group by gen and take the max
  // so the line plots evolution's leading edge over time.
  const fitnessByGen = (() => {
    const bestByGen = new Map()
    for (const d of pop) {
      const g = d.generation ?? 0
      const f = d.fitness ?? 0
      const cur = bestByGen.get(g)
      if (cur == null || f > cur) bestByGen.set(g, f)
    }
    const gens = [...bestByGen.keys()].sort((a, b) => a - b).slice(-20)
    return gens.map(g => bestByGen.get(g))
  })()

  const topValues = TRAITS.map(t => top[t.key] ?? 0)

  return (
    <Panel
      code="04.AM"
      title="CURSE / / GENETIC POPULATION"
      right={<Tag solid>{pop.length} DNA · GEN {maxGen}</Tag>}
    >
      <KV rows={[
        ['GENERATIONS', `${maxGen}`],
        ['POPULATION', `${pop.length}`],
        ['GAMES LOGGED', `${(curse.total_games ?? curse.game_count ?? 0).toLocaleString()}`],
        ['BEST FITNESS', <span style={{ color: bestFitness >= 1.0 ? 'var(--ok)' : 'var(--warn)', fontWeight: 700 }}>{bestFitness.toFixed(2)}</span>],
        ['AVG FITNESS', `${avgFitness.toFixed(2)}`],
        ['TOP GAMES', `${top.games_played ?? 0}`],
      ]} />

      <div className="hr" style={{ margin: '12px 0' }} />
      <div className="t-xs muted" style={{ marginBottom: 4 }}>TOP MEMBER PERSONALITY · 7 PARAMS</div>
      <RadarChart values={topValues} />

      <div className="hr" style={{ margin: '12px 0' }} />
      <div className="t-xs muted" style={{ marginBottom: 4 }}>FITNESS / GEN · LAST {fitnessByGen.length}</div>
      <FitnessSparkline values={fitnessByGen} />
      <div className="t-xs muted-2" style={{ marginTop: 2 }}>
        DASHED = PAR (1.00) · GREEN DOTS ABOVE PAR
      </div>

      <div className="hr" style={{ margin: '12px 0' }} />
      <div className="t-xs muted" style={{ marginBottom: 6 }}>EVAL DIMENSION CORRECTIONS · 20</div>
      <DimHeatmap
        corrections={curse.dim_corrections}
        labels={curse.dim_labels}
        n={curse.dim_stats_n}
      />
    </Panel>
  )
}
