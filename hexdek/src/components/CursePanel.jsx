import { useState, useEffect } from 'react'
import { Panel, KV, Tag } from './chrome'
import { api } from '../services/api'

const TRAITS = [
  { key: 'aggression',        label: 'AGGRESSION' },
  { key: 'combo_patience',    label: 'COMBO PATIENCE' },
  { key: 'threat_paranoia',   label: 'THREAT PARANOIA' },
  { key: 'resource_greed',    label: 'RESOURCE GREED' },
  { key: 'political_memory',  label: 'POLITICAL MEMORY' },
]

function RadarChart({ values, locked, size = 200 }) {
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
        const isLocked = !!(locked && locked[TRAITS[i].key])
        return (
          <circle
            key={i}
            cx={px}
            cy={py}
            r={isLocked ? 3.5 : 2.5}
            fill={isLocked ? 'var(--warn)' : 'var(--ok)'}
            stroke={isLocked ? 'var(--warn)' : 'none'}
            strokeWidth={isLocked ? 1 : 0}
          />
        )
      })}
      {TRAITS.map((t, i) => {
        const [lx, ly] = axisPoint(i, 1.18)
        const anchor = Math.abs(lx - cx) < 4 ? 'middle' : lx < cx ? 'end' : 'start'
        const isLocked = !!(locked && locked[t.key])
        return (
          <text
            key={t.key}
            x={lx}
            y={ly}
            textAnchor={anchor}
            dominantBaseline="middle"
            fontSize="8"
            fill={isLocked ? 'var(--warn)' : 'var(--ink-2)'}
            style={{ letterSpacing: '0.04em', textTransform: 'uppercase', fontWeight: isLocked ? 700 : 400 }}
          >
            {isLocked ? '\u{1F512} ' : ''}{t.label}
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

// num coerces an arbitrary value into a finite number, falling back to
// `fallback` for null / undefined / NaN. Centralizes the partial-snapshot
// defense so every fitness / generation read survives a sparse DNA payload.
const num = (v, fallback = 0) => {
  const n = typeof v === 'number' ? v : Number(v)
  return Number.isFinite(n) ? n : fallback
}

export default function CursePanel({ curse, isOwner = false, deckId = null, onConstraintsChange }) {
  const remoteConstraints = curse?.constraints || null
  const [constraints, setConstraints] = useState(remoteConstraints)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState(null)
  useEffect(() => { setConstraints(remoteConstraints) }, [remoteConstraints])

  // Filter out null / undefined entries from the population snapshot —
  // a partially-populated pool from the backend would otherwise crash
  // on member.fitness / member.generation reads below.
  const pop = (curse?.population || []).filter(d => d != null)
  if (!curse || pop.length === 0) return null

  const sorted = [...pop].sort((a, b) => num(b.fitness) - num(a.fitness))
  const top = sorted[0]
  const maxGen = pop.reduce((m, d) => Math.max(m, num(d.generation)), 0)
  const bestFitness = num(top.fitness)
  const avgFitness = pop.reduce((s, d) => s + num(d.fitness), 0) / pop.length

  // Best fitness per generation, last 20 generations. Each member carries
  // the generation it was created in; we group by gen and take the max
  // fitness so the sparkline plots the leading edge of evolution over time.
  const fitnessByGen = (() => {
    const bestByGen = new Map()
    for (const d of pop) {
      const g = num(d.generation)
      const f = num(d.fitness)
      const cur = bestByGen.get(g)
      if (cur == null || f > cur) bestByGen.set(g, f)
    }
    const gens = [...bestByGen.keys()].sort((a, b) => a - b).slice(-20)
    return gens.map(g => bestByGen.get(g))
  })()

  const topValues = TRAITS.map(t => num(top[t.key]))

  return (
    <Panel
      code="04.AM"
      title={`CURSE / / GENETIC POPULATION`}
      right={<Tag solid>{pop.length} DNA · GEN {maxGen}</Tag>}
    >
      <div className="t-xs muted" style={{ marginBottom: 4 }}>TOP MEMBER PERSONALITY</div>
      <RadarChart values={topValues} locked={constraints} />

      <KV rows={[
        ['GENERATIONS', `${maxGen}`],
        ['POPULATION', `${pop.length}`],
        ['GAMES LOGGED', `${(curse.total_games ?? curse.game_count ?? 0).toLocaleString()}`],
        ['BEST FITNESS', <span style={{ color: bestFitness >= 1.0 ? 'var(--ok)' : 'var(--warn)', fontWeight: 700 }}>{bestFitness.toFixed(2)}</span>],
        ['AVG FITNESS', `${avgFitness.toFixed(2)}`],
        ['TOP GAMES', `${num(top.games_played).toLocaleString()}`],
      ]} />

      <div className="hr" style={{ margin: '10px 0' }} />
      <div className="t-xs muted" style={{ marginBottom: 4 }}>FITNESS / GEN (LAST {fitnessByGen.length})</div>
      <FitnessSparkline values={fitnessByGen} />
      <div className="t-xs muted" style={{ marginTop: 2 }}>
        DASHED = PAR (1.00) · GREEN DOTS ABOVE PAR
      </div>

      <div className="hr" style={{ margin: '12px 0' }} />
      <div className="t-xs muted" style={{ marginBottom: 6 }}>TOP MEMBER TRAITS</div>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(110px, 1fr))', gap: 6 }} className="curse-traits">
        {TRAITS.map(t => {
          const v = num(top[t.key])
          const pct = Math.max(0, Math.min(100, Math.round(v * 100)))
          const isLocked = !!(constraints && Object.prototype.hasOwnProperty.call(constraints, t.key))
          return (
            <div key={t.key} style={{
              border: '1px solid var(--rule-2)',
              padding: '6px 8px',
              background: isLocked ? 'color-mix(in srgb, var(--warn) 8%, transparent)' : 'transparent',
            }}>
              <div className="t-xs muted" style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                {isLocked ? '\u{1F512} ' : ''}{t.label}
              </div>
              <div className="t-xl" style={{ fontWeight: 700, marginTop: 2, color: isLocked ? 'var(--warn)' : 'inherit' }}>{v.toFixed(2)}</div>
              <div style={{ height: 3, background: 'var(--rule-2)', marginTop: 4 }}>
                <div style={{ width: `${pct}%`, height: '100%', background: isLocked ? 'var(--warn)' : 'var(--ok)' }} />
              </div>
            </div>
          )
        })}
      </div>

      {isOwner && (
        <>
          <div className="hr" style={{ margin: '12px 0' }} />
          <div className="t-xs muted" style={{ marginBottom: 4, display: 'flex', justifyContent: 'space-between' }}>
            <span>TRAIT OVERRIDES · OWNER LOCKS</span>
            {busy && <span className="t-xs muted-2">SAVING…</span>}
          </div>
          <div style={{ display: 'grid', gap: 4 }}>
            {TRAITS.map(t => {
              const cur = constraints && Object.prototype.hasOwnProperty.call(constraints, t.key)
              const target = cur ? constraints[t.key] : num(top[t.key])
              const apply = async (next) => {
                if (!deckId) return
                setBusy(true); setError(null)
                const prev = constraints
                setConstraints(next)
                try {
                  const resp = await api.patchDeckCurse(deckId, next)
                  const accepted = resp?.constraints || {}
                  setConstraints(accepted)
                  if (onConstraintsChange) onConstraintsChange(accepted)
                } catch (e) {
                  setError(String(e.message || e)); setConstraints(prev)
                } finally { setBusy(false) }
              }
              const setVal = (v) => apply({ ...(constraints || {}), [t.key]: Math.max(0, Math.min(1, v)) })
              const unlock = () => { const n = { ...(constraints || {}) }; delete n[t.key]; apply(n) }
              return (
                <div key={t.key} style={{
                  display: 'grid',
                  gridTemplateColumns: '110px 28px 1fr 42px',
                  alignItems: 'center', gap: 6, padding: '4px 6px',
                  border: '1px solid var(--rule-2)',
                  background: cur ? 'color-mix(in srgb, var(--warn) 8%, transparent)' : 'transparent',
                }}>
                  <div className="t-xs" style={{ fontWeight: 700, color: cur ? 'var(--warn)' : 'var(--ink-2)' }}>{t.label}</div>
                  <button
                    type="button" disabled={busy}
                    onClick={() => cur ? unlock() : setVal(num(top[t.key], 0.5))}
                    title={cur ? 'Unlock' : 'Lock'}
                    style={{ background: 'transparent', border: '1px solid var(--rule-2)', color: cur ? 'var(--warn)' : 'var(--ink-2)', padding: '2px 4px', cursor: busy ? 'wait' : 'pointer', fontSize: 12, lineHeight: 1 }}
                  >{cur ? '\u{1F512}' : '\u{1F513}'}</button>
                  {cur ? (
                    <input type="range" min="0" max="1" step="0.01" value={target} disabled={busy}
                      onChange={e => setVal(parseFloat(e.target.value))} style={{ width: '100%' }} />
                  ) : (
                    <div style={{ height: 3, background: 'var(--rule-2)' }}>
                      <div style={{ width: `${Math.round(num(top[t.key]) * 100)}%`, height: '100%', background: 'var(--ok)' }} />
                    </div>
                  )}
                  <div className="t-xs" style={{ textAlign: 'right', fontVariantNumeric: 'tabular-nums', color: cur ? 'var(--warn)' : 'var(--ink-2)' }}>{target.toFixed(2)}</div>
                </div>
              )
            })}
          </div>
          {error && <div className="t-xs" style={{ color: 'var(--danger)', marginTop: 4 }}>ERROR: {error}</div>}
          <div className="t-xs muted-2" style={{ marginTop: 4, lineHeight: 1.4 }}>
            LOCKED TRAITS PIN ±0.10 OF TARGET. EVOLUTION RESPECTS LOCKS.
          </div>
        </>
      )}
    </Panel>
  )
}
