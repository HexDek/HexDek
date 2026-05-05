import { Panel, KV, Tag } from './chrome'

const TRAITS = [
  { key: 'aggression',        label: 'AGGRESSION' },
  { key: 'combo_patience',    label: 'COMBO PATIENCE' },
  { key: 'threat_paranoia',   label: 'THREAT PARANOIA' },
  { key: 'resource_greed',    label: 'RESOURCE GREED' },
  { key: 'political_memory',  label: 'POLITICAL MEMORY' },
]

function RadarChart({ values, size = 200 }) {
  const cx = size / 2
  const cy = size / 2
  const r = size * 0.38
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
    <svg viewBox={`0 0 ${size} ${size}`} width="100%" style={{ display: 'block' }}>
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
  )
}

function FitnessSparkline({ values, width = 320, height = 60 }) {
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
    <svg viewBox={`0 0 ${width} ${height}`} width="100%" preserveAspectRatio="none" style={{ display: 'block' }}>
      <line x1="0" y1={baselineY} x2={width} y2={baselineY} stroke="var(--rule-2)" strokeDasharray="3 3" strokeWidth="1" />
      <path d={path} fill="none" stroke="var(--ok)" strokeWidth="1.5" />
      {values.map((v, i) => (
        <circle key={i} cx={x(i)} cy={y(v)} r="2" fill={v >= 1.0 ? 'var(--ok)' : 'var(--ink-2)'} />
      ))}
    </svg>
  )
}

export default function AmiiboPanel({ amiibo }) {
  if (!amiibo || !amiibo.population || amiibo.population.length === 0) return null

  const pop = amiibo.population
  const sorted = [...pop].sort((a, b) => b.fitness - a.fitness)
  const top = sorted[0]
  const maxGen = pop.reduce((m, d) => Math.max(m, d.generation || 0), 0)
  const bestFitness = top.fitness
  const avgFitness = pop.reduce((s, d) => s + d.fitness, 0) / pop.length
  const fitnessByRank = sorted.map(d => d.fitness)

  const topValues = TRAITS.map(t => top[t.key] ?? 0)

  return (
    <Panel
      code="04.AM"
      title={`AMIIBO / / GENETIC POPULATION`}
      right={<Tag solid>{pop.length} DNA · GEN {maxGen}</Tag>}
    >
      <div style={{ display: 'grid', gridTemplateColumns: '220px 1fr', gap: 18, alignItems: 'start' }} className="amiibo-grid">
        <div>
          <div className="t-xs muted" style={{ marginBottom: 4 }}>TOP MEMBER PERSONALITY</div>
          <RadarChart values={topValues} />
        </div>
        <div>
          <KV rows={[
            ['GENERATIONS', `${maxGen}`],
            ['POPULATION', `${pop.length}`],
            ['GAMES LOGGED', `${(amiibo.game_count ?? 0).toLocaleString()}`],
            ['BEST FITNESS', <span style={{ color: bestFitness >= 1.0 ? 'var(--ok)' : 'var(--warn)', fontWeight: 700 }}>{bestFitness.toFixed(2)}</span>],
            ['AVG FITNESS', `${avgFitness.toFixed(2)}`],
            ['TOP GAMES', `${top.games_played ?? 0}`],
          ]} />
          <div className="hr" style={{ margin: '10px 0' }} />
          <div className="t-xs muted" style={{ marginBottom: 4 }}>FITNESS BY RANK (POP SORTED ↓)</div>
          <FitnessSparkline values={fitnessByRank} />
          <div className="t-xs muted" style={{ marginTop: 2 }}>
            DASHED LINE = BRACKET PAR (1.00) · GREEN DOTS ABOVE PAR
          </div>
        </div>
      </div>

      <div className="hr" style={{ margin: '12px 0' }} />
      <div className="t-xs muted" style={{ marginBottom: 6 }}>TOP MEMBER TRAITS</div>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: 8 }} className="amiibo-traits">
        {TRAITS.map(t => {
          const v = top[t.key] ?? 0
          return (
            <div key={t.key} style={{ border: '1px solid var(--rule-2)', padding: '6px 8px' }}>
              <div className="t-xs muted" style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.label}</div>
              <div className="t-xl" style={{ fontWeight: 700, marginTop: 2 }}>{v.toFixed(2)}</div>
              <div style={{ height: 3, background: 'var(--rule-2)', marginTop: 4 }}>
                <div style={{ width: `${Math.round(v * 100)}%`, height: '100%', background: 'var(--ok)' }} />
              </div>
            </div>
          )
        })}
      </div>
    </Panel>
  )
}
