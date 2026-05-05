import { lazy, Suspense, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Panel, KV, Btn, Stripes, Tape } from '../components/chrome'
import { useAuth } from '../context/AuthContext'
import { useLiveSocket } from '../hooks/useLiveSocket'
import { AnimatedCounter } from '../hooks/useAnimatedCounter.jsx'
import { useUploadDeck } from '../hooks/useUploadDeck'

// FishtankEmbed is the heaviest piece on the splash — full WS-driven
// re-render per snapshot. Code-split it so the hero paints first, and
// only mount once the section scrolls into view (poor-man's
// `client:visible`). useLiveSocket is a Context singleton, so deferring
// the mount also defers the WS handshake.
const FishtankEmbed = lazy(() => import('../components/FishtankEmbed'))

const RUNTIME_LABELS = {
  disconnected: 'DISCONNECTED',
  contacting: 'CONTACTING FORGE...',
  initializing: 'INITIALIZING LINK...',
}

export default function Splash() {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { stats, elo, status } = useLiveSocket()
  const upload = useUploadDeck(() => navigate('/decks?tab=mine'))

  const gpm = stats?.games_per_min || 0
  const runtimeText = status === 'live'
    ? `LIVE // ${gpm ? Math.round(gpm / 60).toLocaleString() : '?'} GAMES/SEC`
    : RUNTIME_LABELS[status] || 'OFFLINE'

  const ledClass = status === 'live' ? 'led--on blink' :
    (status === 'contacting' || status === 'initializing') ? 'led--on' : ''

  return (
    <>
      <Tape left="LANDING / / DOC HX-001" mid="REV C.25" right="FORGE TERMINAL" />

      <div className="splash-layout">
        {/* LEFT */}
        <div className="splash-left">
          <div>
            <div className="t-xs muted">DOC. HX-001 / / FORGE TERMINAL / / REV. C.25</div>
            <div style={{ marginTop: 18, display: 'flex', alignItems: 'flex-start', gap: 18 }}>
              <div className="splash-hero">HEX</div>
              <Stripes height={148} w={140} />
            </div>
            <div className="splash-hero" style={{ marginTop: -12 }}>DEK/</div>

            <div className="t-md muted" style={{ marginTop: 24, maxWidth: 540, lineHeight: 1.6, textTransform: 'uppercase', letterSpacing: '0.04em', fontSize: 13 }}>
              &gt; THE FORGE WHERE DECKS BECOME WEAPONS.
              <br />
              &gt; OPEN-SOURCE COMMANDER ANALYSIS ENGINE.
              <br />
              &gt; WASM-NATIVE. RUNS ON YOUR MACHINE.
            </div>
          </div>

          {/* Hero CTA — primary conversion: upload a deck */}
          <button
            onClick={upload.open}
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              gap: 18,
              width: '100%',
              maxWidth: 540,
              padding: '18px 22px',
              background: 'var(--accent)',
              color: 'var(--bg)',
              border: '1px solid var(--accent)',
              fontFamily: 'inherit',
              fontSize: 20,
              fontWeight: 800,
              letterSpacing: '0.08em',
              textTransform: 'uppercase',
              cursor: 'pointer',
              boxShadow: '4px 4px 0 var(--rule-2)',
              transition: 'transform 80ms ease, box-shadow 80ms ease',
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translate(-2px, -2px)'
              e.currentTarget.style.boxShadow = '6px 6px 0 var(--rule-2)'
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translate(0, 0)'
              e.currentTarget.style.boxShadow = '4px 4px 0 var(--rule-2)'
            }}
          >
            <span style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
              <span style={{ fontSize: 28, lineHeight: 1, fontWeight: 900 }}>+</span>
              <span>UPLOAD YOUR DECK</span>
            </span>
            <span style={{ fontSize: 22, lineHeight: 1 }}>▶</span>
          </button>

          <div style={{ display: 'flex', gap: 14, alignItems: 'center', flexWrap: 'wrap' }}>
            <Btn solid arrow="▶" onClick={() => navigate(user ? '/dash' : '/login')}>ENTER THE FORGE</Btn>
            <a href="https://github.com/hexdek-labs/HexDek#readme" target="_blank" rel="noopener noreferrer" style={{ textDecoration: 'none' }}><Btn ghost arrow="↗">DOCS / / README</Btn></a>
            <a href="https://github.com/hexdek-labs/HexDek" target="_blank" rel="noopener noreferrer" style={{ textDecoration: 'none' }}><Btn ghost arrow="↗">GITHUB / / SRC</Btn></a>
            <a href="https://discord.gg/Mz2ueRFXds" target="_blank" rel="noopener noreferrer" style={{ textDecoration: 'none' }}><Btn ghost arrow="↗">DISCORD</Btn></a>
          </div>
        </div>

        {/* RIGHT */}
        <div className="splash-right">
          <Panel code="II.A" title="LIVE FORGE STATS" right={<span className={`led ${ledClass}`} />}>
            <KV rows={[
              ['GAMES SIM.', <AnimatedCounter target={stats?.games_played} rate={gpm} className="punch" style={{ fontSize: 24 }} />],
              ['GAMES/MIN', gpm ? Math.round(gpm).toLocaleString() : '—'],
              ['AVG TURNS', stats ? stats.avg_turns : '—'],
              ['DOMINANT', stats?.dominant?.split(',')[0]?.toUpperCase() || '—'],
              ['TOP WIN RATE', stats ? `${stats.dominant_win_rate}%` : '—'],
              ['ELO POOL', `${elo.length} DECKS`],
            ]} />
          </Panel>

          <div className="panel inv" style={{ padding: 0 }}>
            <div className="panel-hd" style={{ borderColor: 'rgba(0,0,0,0.2)' }}>
              <span>CATALOGUED BUILDS C.25</span>
              <span>SLOT.01</span>
            </div>
            <div style={{ padding: '14px 16px', textAlign: 'center' }}>
              <div style={{ fontSize: 11, letterSpacing: '0.06em', lineHeight: 1.6 }}>
                HEXDEK COMBAT CORE<br />
                ENGINE: HEXDEK V0.10D<br />
                FORMAT: COMMANDER / 1V1 / ARCHENEMY<br />
                RUNTIME: {runtimeText}
              </div>
              <div style={{ borderTop: '1px solid rgba(0,0,0,0.15)', marginTop: 12, paddingTop: 8, fontSize: 10, letterSpacing: '0.1em' }}>
                HEXDEK__©2026
              </div>
            </div>
          </div>

          <Panel code="II.B" title="SYS NOTICE">
            <div className="t-md muted" style={{ lineHeight: 1.6 }}>
              &gt; OPEN SOURCE.<br />
              &gt; DONATIONS-POWERED.<br />
              &gt; NO ADS. NO PAYWALLS.<br />
              &gt; LOGIN OPTIONAL — GUEST = FULL ANALYSIS.
            </div>
          </Panel>
        </div>
      </div>

      {/* Full-width live fishtank section, below the hero. Lazy-mounted
          via WhenVisible so the WS handshake doesn't compete with the
          first paint. The status text re-uses the same useLiveSocket
          context the right column already drives, so a single shared
          WS powers both regions. */}
      <section className="splash-fishtank">
        <Tape
          left="LIVE FORGE / / FISHTANK"
          mid={status === 'live' ? `LIVE · ${gpm ? Math.round(gpm).toLocaleString() : '?'} GAMES/MIN` : (RUNTIME_LABELS[status] || 'OFFLINE')}
          right="WATCH ↗"
        />
        <div className="splash-fishtank-body">
          <WhenVisible
            placeholder={
              <div className="fishtank-embed fishtank-embed--state" style={{ maxHeight: 240 }}>
                <div className="fishtank-embed-hd">
                  <span>FISHTANK / / LIVE FORGE</span>
                  <span className="fishtank-embed-badge">
                    <span className="led" /> STANDBY
                  </span>
                </div>
                <div className="fishtank-embed-empty">
                  &gt; SCROLL TO LOAD LIVE FEED<span className="blink">_</span>
                </div>
              </div>
            }
          >
            <Suspense fallback={
              <div className="fishtank-embed fishtank-embed--state" style={{ maxHeight: 240 }}>
                <div className="fishtank-embed-hd">
                  <span>FISHTANK / / LIVE FORGE</span>
                  <span className="fishtank-embed-badge">
                    <span className="led led--on blink" /> LOADING
                  </span>
                </div>
                <div className="fishtank-embed-empty">
                  &gt; LOADING FISHTANK<span className="blink">_</span>
                </div>
              </div>
            }>
              <FishtankEmbed />
            </Suspense>
          </WhenVisible>
        </div>
      </section>

      {upload.modal}
    </>
  )
}

// WhenVisible defers rendering its children until the placeholder is
// scrolled into (or near) the viewport. Useful for deferring expensive
// mounts — here, the FishtankEmbed's WS-driven snapshot loop. Once
// mounted the children persist (no unmount on scroll-out).
function WhenVisible({ children, placeholder = null, rootMargin = '200px' }) {
  const [visible, setVisible] = useState(false)
  const ref = useRef(null)
  useEffect(() => {
    if (visible) return
    if (typeof IntersectionObserver === 'undefined') {
      setVisible(true)
      return
    }
    const node = ref.current
    if (!node) return
    const obs = new IntersectionObserver((entries) => {
      for (const e of entries) {
        if (e.isIntersecting) {
          setVisible(true)
          obs.disconnect()
          return
        }
      }
    }, { rootMargin })
    obs.observe(node)
    return () => obs.disconnect()
  }, [visible, rootMargin])
  return <div ref={ref}>{visible ? children : placeholder}</div>
}
