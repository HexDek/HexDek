import { useCallback } from 'react'
import { useCardPopup } from './CardPopup'

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
  const { triggerProps, popup } = useCardPopup(name)

  const handleClick = useCallback((e) => {
    e.preventDefault()
    if (stopPropagation) e.stopPropagation()
    if (triggerProps.onClick) triggerProps.onClick(e)
    if (onClick) onClick(e)
  }, [stopPropagation, triggerProps, onClick])

  const baseStyle = underline
    ? { color: 'inherit', textDecoration: 'none', borderBottom: '1px dotted var(--ink-3)', cursor: 'pointer' }
    : { color: 'inherit', textDecoration: 'none', cursor: 'pointer' }

  return (
    <>
      <span
        ref={triggerProps.ref}
        onClick={handleClick}
        style={{ ...baseStyle, ...style }}
        className={className}
        {...rest}
      >
        {children ?? name}
      </span>
      {popup}
    </>
  )
}

// linkifyAction parses a Spectator/GameBoard log entry's free-text
// `action` field and pulls out the card name when the action follows
// one of the engine's templated patterns.
const ACTION_PATTERNS = [
  / CASTS (.+)$/,
  / PLAYS LAND: (.+)$/,
  / COUNTERS (.+)$/,
  / CREATES TOKEN: (.+)$/,
  / DESTROYS (.+)$/,
  / SACRIFICES (.+)$/,
  / → ETB: (.+)$/,
  / UNTAPS (.+)$/,
  / TAPS (.+)$/,
  / DISCARDS (.+)$/,
  / BOUNCES (.+)$/,
  / FLICKERS (.+)$/,
  / EQUIPS (.+)$/,
  / EXILES (.+)$/,
  / REANIMATES (.+)$/,
  / RETURNS (.+) TO HAND$/,
  / ACTIVATES (.+)$/,
  / TRIGGERS (.+)$/,
  / CASCADE → (.+)$/,
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
