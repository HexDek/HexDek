import { useEffect, useState } from 'react'

// ReconnectBanner — full-width banner shown above the appbar whenever
// the global LiveSocket is not in 'live' state. Surfaces:
//   - attempt counter ("ATTEMPT 3 / 10")
//   - countdown to next auto-retry (1s tick, with progress bar)
//   - manual "RECONNECT NOW" button
//   - terminal "GAVE UP" state when MAX_ATTEMPTS is exceeded
//
// Hidden entirely when status is 'live' or 'initializing' (mid-handshake)
// — the brief flash during a healthy connect would be noise. The
// 'contacting' state shows for the very first connect attempt and
// during a manual reconnect; that's the one place we want a banner so
// the user knows their click did something.

export default function ReconnectBanner({
  status,
  attempt,
  maxAttempts,
  nextRetryAt,
  onReconnect,
  label = 'LIVE FORGE',
}) {
  const [, tick] = useState(0)
  useEffect(() => {
    if (status === 'live' || status === 'initializing') return
    const id = setInterval(() => tick(t => t + 1), 250)
    return () => clearInterval(id)
  }, [status])

  if (status === 'live' || status === 'initializing') return null

  const now = Date.now()
  const totalDelay = nextRetryAt && attempt > 0
    ? Math.max(1000, nextRetryAt - (now - 250))
    : 0
  const remainingMs = nextRetryAt ? Math.max(0, nextRetryAt - now) : 0
  const remainingSecs = Math.ceil(remainingMs / 1000)
  const pct = totalDelay > 0
    ? Math.min(100, Math.max(0, (1 - remainingMs / totalDelay) * 100))
    : 0

  const failed = status === 'failed'
  const contacting = status === 'contacting'

  let message
  if (contacting) {
    message = 'CONTACTING SERVER...'
  } else if (failed) {
    message = `GAVE UP AFTER ${maxAttempts} ATTEMPTS`
  } else if (attempt > 0) {
    message = `ATTEMPT ${attempt} / ${maxAttempts}`
  } else {
    message = 'DISCONNECTED'
  }

  return (
    <div className={`reconnect-banner${failed ? ' reconnect-banner--failed' : ''}`} role="status" aria-live="polite">
      <div className="reconnect-banner-row">
        <span className="reconnect-banner-led">
          <span className={`led led--${failed ? 'bad' : 'warn'}${!failed ? ' blink' : ''}`} />
        </span>
        <span className="reconnect-banner-label">{label} / / {message}</span>
        <span className="reconnect-banner-spacer" />
        {!failed && !contacting && remainingSecs > 0 && (
          <span className="reconnect-banner-countdown">
            NEXT RETRY IN <strong>{remainingSecs}s</strong>
          </span>
        )}
        <button
          type="button"
          className="reconnect-banner-btn"
          onClick={onReconnect}
          disabled={contacting}
        >
          {contacting ? 'CONNECTING…' : 'RECONNECT NOW ↻'}
        </button>
      </div>
      {!failed && !contacting && totalDelay > 0 && (
        <div className="reconnect-banner-progress" aria-hidden="true">
          <div className="reconnect-banner-progress-fill" style={{ width: `${pct}%` }} />
        </div>
      )}
    </div>
  )
}
