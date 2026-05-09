import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Panel, Tape } from '../components/chrome'
import { completeMagicLinkSignIn } from '../lib/firebase'
import { broadcastAuth, AUTH_EVENT } from '../lib/authBroadcast'

// AuthCallback — magic-link landing page. Most email clients open
// links in a new tab, so the typical flow is:
//   1) verify the sign-in link with Firebase
//   2) broadcast {type:'authed', email} so the tab the user came from
//      can swap to the MagicLinkConsole and forward to /operator
//   3) try to close this tab (works only when this tab was opened by
//      script — i.e. usually fails for email-client tabs)
//   4) if window.close() didn't take effect, navigate this tab to
//      /operator so the user still lands somewhere useful
//
// We do (3) before (4) because firing both is harmless: window.close()
// either succeeds (and (4) never runs because the tab is gone) or
// silently no-ops. The brief "AUTHENTICATED — CLOSING…" state covers
// the gap so the page never looks frozen.

export default function AuthCallback() {
  const [status, setStatus] = useState('VERIFYING...')
  const [error, setError] = useState(null)
  const navigate = useNavigate()

  useEffect(() => {
    let cancelled = false

    completeMagicLinkSignIn()
      .then((user) => {
        if (cancelled) return
        if (!user) {
          setError('INVALID OR EXPIRED LINK.')
          broadcastAuth({ type: AUTH_EVENT.FAILED, reason: 'invalid_or_expired' })
          return
        }
        setStatus('AUTHENTICATED — HANDING OFF…')
        broadcastAuth({ type: AUTH_EVENT.SUCCEEDED, email: user.email || null })

        // Best-effort tab close. If this tab was opened by a script
        // (window.open), close() succeeds. From an email-client tab
        // it usually doesn't, so we fall through to a redirect.
        setTimeout(() => {
          try { window.close() } catch {}
          if (!cancelled && !window.closed) {
            navigate('/operator', { replace: true })
          }
        }, 350)
      })
      .catch(() => {
        if (cancelled) return
        setError('AUTH VERIFICATION FAILED.')
        broadcastAuth({ type: AUTH_EVENT.FAILED, reason: 'exception' })
      })

    return () => { cancelled = true }
  }, [navigate])

  return (
    <>
      <Tape left="AUTH / / CALLBACK" mid="VERIFYING" right="MAGIC LINK" />
      <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 36 }}>
        <Panel code="AUTH.02" title="LINK VERIFICATION" solid style={{ maxWidth: 420, width: '100%' }}>
          {error ? (
            <div>
              <div className="t-xl" style={{ fontWeight: 700, color: 'var(--danger)' }}>{error}</div>
              <div className="t-xs muted" style={{ marginTop: 8 }}>&gt; TRY REQUESTING A NEW LINK.</div>
            </div>
          ) : (
            <div>
              <div className="t-xl" style={{ fontWeight: 700, color: 'var(--ok)' }}>{status}</div>
              <div className="t-xs muted" style={{ marginTop: 8 }}>&gt; YOU CAN CLOSE THIS TAB<span className="blink">_</span></div>
            </div>
          )}
        </Panel>
      </div>
    </>
  )
}
