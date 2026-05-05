import { Link } from 'react-router-dom'

// CardLink wraps a card name (or arbitrary children) in a Link to
// /cards/:cardName. Single canonical helper so every card-name display
// across the app navigates to the same place.
//
// Props:
//   name              — required card name (URL-encoded by this component)
//   children          — optional override for displayed text; defaults to name
//   stopPropagation   — true by default; lets CardLink sit inside parent
//                       rows that have their own onClick (DeckList rows,
//                       Leaderboard rows) without double-firing
//   underline         — true by default; renders a 1px dotted underline so
//                       links are visible against the brutalist palette.
//                       Pass false when the link wraps an art tile / icon
//                       and an underline would be visual noise.
//   onClick           — optional extra handler (after stopPropagation)
//   style, className  — passthrough
export default function CardLink({
  name,
  children,
  stopPropagation = true,
  underline = true,
  onClick,
  style,
  className,
  ...rest
}) {
  if (!name) return <span className={className} style={style}>{children}</span>
  const href = `/cards/${encodeURIComponent(name)}`
  const handleClick = (e) => {
    if (stopPropagation) e.stopPropagation()
    if (onClick) onClick(e)
  }
  const baseStyle = underline
    ? { color: 'inherit', textDecoration: 'none', borderBottom: '1px dotted var(--ink-3)' }
    : { color: 'inherit', textDecoration: 'none' }
  return (
    <Link
      to={href}
      onClick={handleClick}
      style={{ ...baseStyle, ...style }}
      className={className}
      {...rest}
    >
      {children ?? name}
    </Link>
  )
}

// linkifyAction parses a Spectator/GameBoard log entry's free-text
// `action` field and pulls out the card name when the action follows
// one of the engine's templated patterns (CASTS X, PLAYS LAND: X,
// COUNTERS X, CREATES TOKEN: X, DESTROYS X, SACRIFICES X, → ETB: X).
//
// Returns { prefix, cardName } where:
//   - prefix is everything BEFORE the card name (including the verb +
//     any separator), already trimmed of trailing whitespace
//   - cardName is the raw matched substring (typically uppercased to
//     match the log's visual style); use it as both the displayed text
//     and the CardLink `name` prop — Scryfall's exact-match endpoint
//     is case-insensitive
//   - cardName is null when the action doesn't match any known pattern;
//     callers should fall back to rendering the action verbatim.
const ACTION_PATTERNS = [
  / CASTS (.+)$/,
  / PLAYS LAND: (.+)$/,
  / COUNTERS (.+)$/,
  / CREATES TOKEN: (.+)$/,
  / DESTROYS (.+)$/,
  / SACRIFICES (.+)$/,
  / → ETB: (.+)$/,
]

export function linkifyAction(action) {
  if (!action || typeof action !== 'string') return { prefix: action || '', cardName: null }
  for (const re of ACTION_PATTERNS) {
    const m = action.match(re)
    if (m) {
      const card = m[1].trim()
      if (!card) continue
      const splitAt = m.index + m[0].length - m[1].length
      return { prefix: action.slice(0, splitAt), cardName: card }
    }
  }
  return { prefix: action, cardName: null }
}
