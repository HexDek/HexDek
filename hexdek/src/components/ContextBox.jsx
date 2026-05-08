// ContextBox — small muted helper text rendered above CTAs.
// Designed for neurodivergent UX clarity: every action button should be
// preceded by a 1-2 sentence TLDR explaining what will happen on click.
//
// Usage:
//   <ContextBox>Runs a 500-game gauntlet against the meta. Takes a few minutes.</ContextBox>
//   <Btn solid>RUN GAUNTLET</Btn>

export default function ContextBox({ children, tone = 'info', style, compact = false }) {
  if (!children) return null

  const accent =
    tone === 'warn' ? 'var(--warn, #c9a227)' :
    tone === 'danger' ? 'var(--danger, #b3433d)' :
    'var(--ink-3)'

  return (
    <div
      className="ctx-box"
      style={{
        margin: compact ? '0 0 4px 0' : '0 0 6px 0',
        padding: compact ? '4px 8px' : '6px 10px',
        borderLeft: `2px solid ${accent}`,
        background: 'var(--bg-2, rgba(0,0,0,0.18))',
        color: 'var(--ink-2)',
        fontSize: 10,
        lineHeight: 1.45,
        letterSpacing: '0.03em',
        fontStyle: 'normal',
        ...style,
      }}
    >
      {children}
    </div>
  )
}
