import { useNavigate } from 'react-router-dom'
import { useLiveSocket } from '../hooks/useLiveSocket'
import { cardArtUrl } from '../services/api'

export default function FishtankEmbed() {
  const navigate = useNavigate()
  const { game, status } = useLiveSocket()

  const goSpectate = () => navigate('/spectate')

  // Loading / disconnected gates
  if (status === 'disconnected') {
    return (
      <div className="fishtank-embed fishtank-embed--state" onClick={goSpectate}>
        <div className="fishtank-embed-hd">
          <span>FISHTANK / / LIVE FORGE</span>
          <span className="fishtank-embed-badge fishtank-embed-badge--off">OFFLINE</span>
        </div>
        <div className="fishtank-embed-empty">
          &gt; FISHTANK OFFLINE<br />
          &gt; SHOWMATCH ENGINE NOT REACHABLE<span className="blink">_</span>
        </div>
      </div>
    )
  }

  if (!game || game.status === 'starting' || !game.seats) {
    return (
      <div className="fishtank-embed fishtank-embed--state" onClick={goSpectate}>
        <div className="fishtank-embed-hd">
          <span>FISHTANK / / LIVE FORGE</span>
          <span className="fishtank-embed-badge">
            <span className="led led--on" /> CONNECTING
          </span>
        </div>
        <div className="fishtank-embed-empty">
          &gt; CONTACTING FORGE...<br />
          &gt; LOADING FIRST SHOWMATCH<span className="blink">_</span>
        </div>
      </div>
    )
  }

  const seats = game.seats || []
  const numSeats = seats.length || 4
  const round = Math.ceil(game.turn / numSeats)
  const rt = `R${round}T${game.turn}`
  const phase = (game.phase || '').toUpperCase()
  const finished = !!game.finished

  return (
    <div
      className="fishtank-embed"
      onClick={goSpectate}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') goSpectate() }}
    >
      <div className="fishtank-embed-hd">
        <span>FISHTANK / / GAME #{game.game_id}</span>
        <span className="fishtank-embed-badge">
          <span className={`led led--on ${!finished ? 'blink' : ''}`} />
          {finished ? 'GAME OVER' : 'WATCH LIVE'}
        </span>
      </div>

      <div className="fishtank-embed-grid">
        {[0, 1, 3, 2].filter(i => i < seats.length).map(i => {
          const s = seats[i]
          const isActive = i === game.active_seat && !finished
          const isWinner = finished && game.winner === i
          const isLost = s.lost && !isWinner
          const artUrl = cardArtUrl(s.commander)
          const permCount = (s.battlefield || []).length

          return (
            <div
              key={i}
              className="fishtank-embed-seat"
              style={{
                borderColor: isWinner ? 'var(--ok)' : isActive ? 'var(--warn)' : undefined,
              }}
            >
              {artUrl && (
                <div
                  className="fishtank-embed-seat-art"
                  style={{
                    backgroundImage: `url(${artUrl})`,
                    opacity: isLost ? 0.18 : 0.42,
                  }}
                />
              )}
              <div className="fishtank-embed-seat-body">
                <div className="fishtank-embed-seat-name">
                  {(s.commander || 'UNKNOWN').toUpperCase().split('//')[0].trim()}
                  {isActive && <span style={{ color: 'var(--warn)' }}> ●</span>}
                  {isWinner && <span style={{ color: 'var(--ok)' }}> ★</span>}
                  {isLost && <span style={{ color: 'var(--danger)' }}> ✕</span>}
                </div>
                <div className="fishtank-embed-seat-stats">
                  <span>♥{s.life ?? '—'}</span>
                  <span className="muted-2">·</span>
                  <span>B{permCount}</span>
                  <span className="muted-2">·</span>
                  <span>H{s.hand_size ?? '?'}</span>
                </div>
              </div>
            </div>
          )
        })}
      </div>

      <div className="fishtank-embed-ft">
        <span>
          {finished
            ? `WINNER: ${game.winner >= 0 ? (seats[game.winner]?.commander || '—').toUpperCase().split('//')[0].trim() : 'DRAW'}`
            : `${rt} · ${phase}${game.step ? ` / ${game.step.toUpperCase()}` : ''}`}
        </span>
        <span className="fishtank-embed-ft-cta">OPEN SPECTATOR ↗</span>
      </div>
    </div>
  )
}
