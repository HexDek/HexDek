// ManaCost — render a Scryfall-style mana_cost string as colored pips.
//
// Input shape: "{2}{G}{U}", "{W/U}{W/U}", "{X}{R}{R}", "{2/W}{B/P}",
// "{C}", "{S}", etc. Anything inside {…} is one pip; characters
// outside braces are ignored.
//
// Output: an inline-flex row of small SVG circles, one per pip.
// Pure inline SVG, no external font. Hybrid pips render as a
// split circle (left half = first color, right half = second).
//
// Usage:
//   <ManaCost cost={card.mana_cost} size={16} />

import { useId } from 'react'

// Brutalist palette — muted enough to sit next to the cream/black
// design without screaming. Border darkens each fill slightly so
// pips read on both light and dark backgrounds.
const COLOR = {
  W: '#eddcb1', // bone / off-cream
  U: '#3a76b6', // steel blue
  B: '#2a2526', // near-black with warm tint
  R: '#cc4a3a', // brick red
  G: '#3f8f57', // forest green
  C: '#a59f8e', // light gray (colorless)
  S: '#cdd5dc', // pale snow gray
  X: '#a59f8e', // X variable cost — same gray as colorless
  N: '#a59f8e', // generic numeric — same gray
  P: '#9c6ab0', // phyrexian — purple-ish for visibility
}

// Glyph drawn inside the pip. Numbers use their digit; W/U/B/R/G use
// nothing (color speaks); X/Y/Z use the letter; phyrexian uses Φ.
function pipGlyph(token) {
  const t = token.toUpperCase()
  if (/^\d+$/.test(t)) return t
  if (t === 'X' || t === 'Y' || t === 'Z') return t
  if (t === 'C') return ''
  if (t === 'S') return '❄'
  if (t === 'P') return 'Φ'
  if (t.length === 1 && 'WUBRG'.includes(t)) return ''
  return ''
}

// Pip text color. White-fill (W) and snow-fill (S) need black ink;
// everything else gets ink-on-color.
function glyphColor(fill) {
  return fill === COLOR.W || fill === COLOR.S ? '#1a1814' : '#f0ebd8'
}

// classify(token) → { kind, fill, fill2?, label }
//   kind: 'simple' | 'hybrid' | 'numeric'
function classify(token) {
  const raw = token.toUpperCase().trim()
  // Numeric: {0}, {1}, {2}, ...
  if (/^\d+$/.test(raw)) {
    return { kind: 'numeric', fill: COLOR.N, label: raw }
  }
  // Variable: {X}, {Y}, {Z}
  if (raw === 'X' || raw === 'Y' || raw === 'Z') {
    return { kind: 'numeric', fill: COLOR.X, label: raw }
  }
  // Single color / colorless / snow.
  if (raw.length === 1 && COLOR[raw]) {
    return { kind: 'simple', fill: COLOR[raw], label: pipGlyph(raw) }
  }
  // Hybrid: "W/U", "2/W", "W/P", "B/P", "C/W" etc.
  if (raw.includes('/')) {
    const parts = raw.split('/')
    const a = parts[0], b = parts[1]
    const fillA = /^\d+$/.test(a) ? COLOR.N : (COLOR[a] || COLOR.N)
    const fillB = /^\d+$/.test(b) ? COLOR.N : (COLOR[b] || COLOR.N)
    // Phyrexian (X/P) — render the colored half on the left, purple on
    // the right; the Φ glyph centers across the seam.
    return {
      kind: 'hybrid',
      fill: fillA,
      fill2: fillB,
      label: b === 'P' ? 'Φ' : (/^\d+$/.test(a) ? a : ''),
    }
  }
  // Unknown token → render as a question-mark gray pip so the user
  // sees the gap rather than silently dropping it.
  return { kind: 'simple', fill: COLOR.N, label: '?' }
}

function parseCost(cost) {
  if (!cost || typeof cost !== 'string') return []
  const matches = cost.match(/\{([^}]+)\}/g) || []
  return matches.map(m => m.slice(1, -1))
}

function Pip({ token, size, idPrefix }) {
  const meta = classify(token)
  const stroke = '#1a1814'
  const r = (size - 2) / 2
  const cx = size / 2
  const cy = size / 2
  const fontSize = Math.round(size * 0.6)
  const ink = glyphColor(meta.fill)

  if (meta.kind === 'hybrid') {
    // Right semicircle drawn as an SVG path so it sits cleanly over
    // the base circle. Border drawn last to keep edges crisp.
    return (
      <svg
        viewBox={`0 0 ${size} ${size}`}
        width={size}
        height={size}
        aria-label={token}
      >
        <circle cx={cx} cy={cy} r={r} fill={meta.fill} />
        <path d={`M ${cx} ${cy - r} A ${r} ${r} 0 0 1 ${cx} ${cy + r} Z`} fill={meta.fill2} />
        <circle cx={cx} cy={cy} r={r} fill="none" stroke={stroke} strokeWidth="1" />
        {meta.label && (
          <text
            x={cx} y={cy + 1}
            textAnchor="middle"
            dominantBaseline="central"
            fontSize={fontSize * 0.85}
            fontWeight="700"
            fill={glyphColor(meta.fill2)}
          >
            {meta.label}
          </text>
        )}
      </svg>
    )
  }

  return (
    <svg
      viewBox={`0 0 ${size} ${size}`}
      width={size}
      height={size}
      aria-label={token}
    >
      <circle cx={cx} cy={cy} r={r} fill={meta.fill} stroke={stroke} strokeWidth="1" />
      {meta.label && (
        <text
          x={cx} y={cy + 1}
          textAnchor="middle"
          dominantBaseline="central"
          fontSize={meta.label.length > 1 ? fontSize * 0.7 : fontSize}
          fontWeight="700"
          fill={ink}
          fontFamily="ui-monospace, SFMono-Regular, Menlo, monospace"
        >
          {meta.label}
        </text>
      )}
    </svg>
  )
}

export default function ManaCost({ cost, size = 16, gap = 2, style, className, title }) {
  const tokens = parseCost(cost)
  if (tokens.length === 0) return null
  const id = useId().replace(/:/g, '')
  return (
    <span
      className={className}
      title={title || cost}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap,
        verticalAlign: 'middle',
        ...style,
      }}
    >
      {tokens.map((t, i) => (
        <Pip key={`${id}-${i}`} token={t} size={size} idPrefix={`${id}-${i}`} />
      ))}
    </span>
  )
}
