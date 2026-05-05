import { useEffect, useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { useLiveSocket } from '../hooks/useLiveSocket'
import { cardArtUrl } from '../services/api'

// StreamOverlay — transparent OBS-friendly HUD for /stream/:gameId.
//
// Design constraints:
//   * No AppShell chrome (route is mounted outside AppShell).
//   * html/body forced transparent on mount, restored on unmount,
//     so OBS browser source sees alpha through the entire surface.
//   * All elements sit in fixed-position regions with their own
//     semi-opaque backgrounds, never full-bleed — bystander pixels
//     stay alpha=0 so the underlying stream layer shows through.
//   * Reads live engine data via the same useLiveSocket the Spectator
//     uses. The :gameId URL param is a sanity check: when it doesn't
//     match the current live game, the HUD shows a small "waiting"
//     pill instead of the wrong game's data.

const SHORT_NAME = (commander) => {
  if (!commander) return 'UNKNOWN'
  const trimmed = commander.split(',')[0].split('//')[0].trim()
  return trimmed.toUpperCase()
}

function useTransparentRoot() {
  useEffect(() => {
    const root = document.documentElement
    const body = document.body
    const prevHtmlBg = root.style.background
    const prevBodyBg = body.style.background
    const prevHtmlColor = root.style.colorScheme
    root.style.background = 'transparent'
    body.style.background = 'transparent'
    root.style.colorScheme = 'normal'
    return () => {
      root.style.background = prevHtmlBg
      body.style.background = prevBodyBg
      root.style.colorScheme = prevHtmlColor
    }
  }, [])
}

function SeatTile({ seat, idx, isActive, isWinner, isLost }) {
  const art = cardArtUrl(seat.commander)
  const accent = isWinner ? 'var(--ok)' : isActive ? 'var(--warn)' : isLost ? 'var(--danger)' : 'rgba(255,255,255,0.45)'
  return (
    <div style={{
      display: 'grid',
      gridTemplateColumns: '40px 1fr auto',
      gap: 8,
      alignItems: 'center',
      padding: '6px 8px',
      background: 'rgba(0,0,0,0.78)',
      border: `1px solid ${accent}`,
      boxShadow: isActive ? '0 0 0 1px rgba(212, 168, 67, 0.35)' : 'none',
      opacity: isLost && !isWinner ? 0.55 : 1,
    }}>
      <div
        style={{
          width: 40, height: 40,
          backgroundImage: art ? `url(${art})` : undefined,
          backgroundColor: '#0c0d0a',
          backgroundSize: 'cover',
          backgroundPosition: 'center 30%',
          border: '1px solid rgba(255,255,255,0.15)',
          filter: isLost && !isWinner ? 'grayscale(0.8) brightness(0.6)' : 'none',
        }}
      />
      <div style={{ minWidth: 0 }}>
        <div style={{
          fontSize: 11, fontWeight: 800,
          letterSpacing: '0.06em', textTransform: 'uppercase',
          color: '#f4f0e6',
          overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
          textShadow: '0 1px 2px rgba(0,0,0,0.85)',
        }}>{SHORT_NAME(seat.commander)}</div>
        <div style={{
          fontSize: 9, letterSpacing: '0.08em',
          color: 'rgba(255,255,255,0.55)',
          textTransform: 'uppercase',
        }}>
          SEAT.{String(idx + 1).padStart(2, '0')}
          {isActive && <span style={{ color: 'var(--warn)', marginLeft: 6 }}>● ACTIVE</span>}
          {isWinner && <span style={{ color: 'var(--ok)', marginLeft: 6 }}>★ WIN</span>}
          {isLost && !isWinner && <span style={{ color: 'var(--danger)', marginLeft: 6 }}>✕ ELIM</span>}
        </div>
      </div>
      <div style={{
        fontSize: 18, fontWeight: 900,
        fontVariantNumeric: 'tabular-nums',
        letterSpacing: '-0.02em',
        color: seat.life <= 0 ? 'var(--danger)' : seat.life <= 10 ? 'var(--warn)' : '#f4f0e6',
        textShadow: '0 1px 3px rgba(0,0,0,0.85)',
        minWidth: 36, textAlign: 'right',
      }}>
        ♥{seat.life}
      </div>
    </div>
  )
}

function Ticker({ entries }) {
  if (!entries || entries.length === 0) return null
  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      gap: 4,
      maxWidth: 540,
      padding: '8px 12px',
      background: 'rgba(0,0,0,0.78)',
      border: '1px solid rgba(255,255,255,0.18)',
    }}>
      {entries.map((e, i) => (
        <div
          key={`${e.turn}-${i}`}
          style={{
            display: 'grid',
            gridTemplateColumns: '38px 1fr',
            gap: 8,
            fontSize: 11,
            letterSpacing: '0.04em',
            color: i === 0 ? '#f4f0e6' : 'rgba(255,255,255,0.55)',
            opacity: i === 0 ? 1 : 0.7 - i * 0.15,
            fontWeight: i === 0 ? 700 : 500,
          }}
        >
          <span style={{
            color: 'rgba(255,255,255,0.45)',
            fontVariantNumeric: 'tabular-nums',
            fontSize: 10,
          }}>T{e.turn}</span>
          <span style={{
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
            textTransform: 'uppercase',
            textShadow: '0 1px 2px rgba(0,0,0,0.85)',
          }}>
            &gt; {e.action}
          </span>
        </div>
      ))}
    </div>
  )
}

export default function StreamOverlay() {
  useTransparentRoot()
  const { gameId } = useParams()
  const { game, status } = useLiveSocket()

  const seats = game?.seats || []
  const log = game?.log || []
  const activeSeat = game?.active_seat ?? -1
  const turn = game?.turn ?? 0
  const finished = !!game?.finished
  const winner = finished ? game?.winner ?? -1 : -1
  const liveGameId = game?.game_id != null ? String(game.game_id) : null

  // The most-recent 3 log lines, newest first. Reverse since the
  // engine appends in chronological order.
  const tickerEntries = useMemo(() => {
    if (!Array.isArray(log) || log.length === 0) return []
    return [...log].slice(-3).reverse().map(e => ({
      turn: e.turn,
      action: e.action,
      kind: e.kind,
    }))
  }, [log])

  // URL gameId is optional; when present, mismatch shows a waiting
  // pill instead of the wrong game's data so a streamer can pre-set
  // their OBS source to a specific game id.
  const mismatch = gameId && liveGameId && gameId !== liveGameId

  if (status === 'disconnected' || !game) {
    return (
      <div style={overlayShellStyle}>
        <StatusPill label="OFFLINE" tone="danger" />
      </div>
    )
  }

  if (mismatch) {
    return (
      <div style={overlayShellStyle}>
        <StatusPill label={`WAITING FOR GAME #${gameId}`} tone="muted" />
      </div>
    )
  }

  const activeName = activeSeat >= 0 && seats[activeSeat]
    ? SHORT_NAME(seats[activeSeat].commander)
    : '—'
  const activeArt = activeSeat >= 0 && seats[activeSeat]?.commander
    ? cardArtUrl(seats[activeSeat].commander)
    : null

  return (
    <div style={overlayShellStyle}>
      {/* Top-left: turn + active player */}
      <div style={{
        position: 'absolute',
        top: 18, left: 18,
        display: 'flex', alignItems: 'center', gap: 10,
      }}>
        <div style={{
          padding: '8px 14px',
          background: 'rgba(0,0,0,0.82)',
          border: '1px solid var(--warn)',
          display: 'flex', alignItems: 'center', gap: 12,
        }}>
          <div>
            <div style={{
              fontSize: 9, letterSpacing: '0.14em',
              color: 'rgba(255,255,255,0.55)',
              textTransform: 'uppercase',
            }}>TURN</div>
            <div style={{
              fontSize: 24, fontWeight: 900, lineHeight: 1,
              color: '#f4f0e6',
              fontVariantNumeric: 'tabular-nums',
              letterSpacing: '-0.03em',
              textShadow: '0 1px 3px rgba(0,0,0,0.85)',
            }}>{turn}</div>
          </div>
          <div style={{
            width: 1, height: 36,
            background: 'rgba(255,255,255,0.2)',
          }} />
          {activeArt && (
            <div style={{
              width: 36, height: 36,
              backgroundImage: `url(${activeArt})`,
              backgroundSize: 'cover',
              backgroundPosition: 'center 30%',
              border: '1px solid rgba(255,255,255,0.2)',
            }} />
          )}
          <div>
            <div style={{
              fontSize: 9, letterSpacing: '0.14em',
              color: 'rgba(255,255,255,0.55)',
              textTransform: 'uppercase',
            }}>ACTIVE</div>
            <div style={{
              fontSize: 13, fontWeight: 800,
              color: 'var(--warn)',
              letterSpacing: '0.04em',
              textTransform: 'uppercase',
              textShadow: '0 1px 2px rgba(0,0,0,0.85)',
              maxWidth: 220,
              overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
            }}>{activeName}</div>
          </div>
        </div>
        {finished && <StatusPill label={winner >= 0 ? `WINNER · ${SHORT_NAME(seats[winner]?.commander)}` : 'DRAW'} tone="ok" />}
      </div>

      {/* Top-right: seat life totals + commanders */}
      <div style={{
        position: 'absolute',
        top: 18, right: 18,
        display: 'flex', flexDirection: 'column',
        gap: 6, width: 280,
      }}>
        {seats.map((s, i) => (
          <SeatTile
            key={i}
            seat={s}
            idx={i}
            isActive={i === activeSeat && !finished}
            isWinner={finished && i === winner}
            isLost={!!s.lost}
          />
        ))}
      </div>

      {/* Bottom-left: action ticker */}
      <div style={{ position: 'absolute', bottom: 18, left: 18 }}>
        <Ticker entries={tickerEntries} />
      </div>

      {/* Bottom-right: tiny brand mark */}
      <div style={{
        position: 'absolute',
        bottom: 18, right: 18,
        fontSize: 9, letterSpacing: '0.18em',
        color: 'rgba(255,255,255,0.35)',
        textTransform: 'uppercase',
        textShadow: '0 1px 2px rgba(0,0,0,0.85)',
      }}>HEXDEK//STREAM</div>
    </div>
  )
}

const overlayShellStyle = {
  position: 'fixed',
  inset: 0,
  pointerEvents: 'none',
  fontFamily: "'JetBrains Mono', ui-monospace, monospace",
  background: 'transparent',
  overflow: 'hidden',
}

function StatusPill({ label, tone = 'muted' }) {
  const color = tone === 'danger' ? 'var(--danger)' : tone === 'ok' ? 'var(--ok)' : 'rgba(255,255,255,0.55)'
  return (
    <div style={{
      padding: '6px 12px',
      background: 'rgba(0,0,0,0.82)',
      border: `1px solid ${color}`,
      color,
      fontSize: 11,
      fontWeight: 800,
      letterSpacing: '0.14em',
      textTransform: 'uppercase',
      textShadow: '0 1px 2px rgba(0,0,0,0.85)',
    }}>
      {label}
    </div>
  )
}
