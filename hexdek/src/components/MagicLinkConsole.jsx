import { useEffect, useRef, useState } from 'react'

// MagicLinkConsole — fullscreen "logging in" feed shown on the
// original tab once the email-link tab BroadcastChannels a successful
// auth. Streams a scripted sequence of one-line status messages with
// timing tuned to feel like a real auth handshake (~2.4s total) before
// firing onComplete, which the parent uses to navigate to /operator.
//
// Each line streams in newest-at-bottom; the last line wears a
// blinking caret until the next line lands, so the eye always has a
// "currently working" anchor. Colors map to console conventions:
//   info: muted, ok: green, warn: amber, accent: blue/header.
//
// The component is purely presentational — the parent owns the
// BroadcastChannel listener and decides when to mount/unmount it.

const SCRIPT = [
  { t: 0,    tone: 'info',   text: '> handshake received from auth-callback tab' },
  { t: 220,  tone: 'info',   text: '> verifying signed firebase id-token' },
  { t: 480,  tone: 'ok',     text: '+ token signature OK' },
  { t: 700,  tone: 'info',   text: '> resolving owner slug from email' },
  { t: 1020, tone: 'ok',     text: '+ owner resolved' },
  { t: 1240, tone: 'info',   text: '> stitching browser session to operator id' },
  { t: 1520, tone: 'ok',     text: '+ session stitched' },
  { t: 1760, tone: 'info',   text: '> hydrating operator profile from /api/decks' },
  { t: 2080, tone: 'ok',     text: '+ profile hydrated' },
  { t: 2280, tone: 'accent', text: '>>> ENTERING OPERATOR CONSOLE' },
]

export default function MagicLinkConsole({ email, onComplete, redirectTo = '/operator' }) {
  const [lines, setLines] = useState([])
  const completedRef = useRef(false)

  useEffect(() => {
    const timers = []
    SCRIPT.forEach((line, idx) => {
      timers.push(setTimeout(() => {
        setLines(prev => [...prev, { id: idx, ...line }])
      }, line.t))
    })
    timers.push(setTimeout(() => {
      if (completedRef.current) return
      completedRef.current = true
      if (onComplete) onComplete()
    }, SCRIPT[SCRIPT.length - 1].t + 600))
    return () => timers.forEach(clearTimeout)
  }, [onComplete])

  return (
    <div className="magic-console" role="dialog" aria-label="Logging in">
      <div className="magic-console-hd">
        <span>HEXDEK / / AUTH HANDSHAKE</span>
        <span>{email ? `OPERATOR ${email.toUpperCase()}` : 'NEW SESSION'}</span>
        <span>→ {redirectTo.toUpperCase()}</span>
      </div>
      <div className="magic-console-body">
        {lines.map((line, i) => {
          const isLast = i === lines.length - 1
          return (
            <span
              key={line.id}
              className={`magic-console-line magic-console-line--${line.tone}${isLast ? ' magic-console-caret' : ''}`}
            >
              {line.text}
            </span>
          )
        })}
      </div>
    </div>
  )
}
