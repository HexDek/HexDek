import { cardArtUrl } from '../services/api'
import { useArtContrast } from '../hooks/useArtContrast'

// DeckShelf — grid of commander-art deck tiles. Used by DeckList (with
// the "ADD YOUR DECK" upload card) and by PublicProfile (without it,
// since visitors can't upload to someone else's collection). Centralises
// the visual contract so both screens stay in sync if the tile design
// changes.
//
// Props:
//   decks         — array of deck-summary objects from /api/decks
//   eloByDeckId   — { "owner/id": ELOEntry } for record/rating overlay
//   navigate      — react-router navigate fn (parent decides routing)
//   onAddCard     — optional click handler; if set, renders the
//                   "+ ADD YOUR DECK" tile before the deck cards
//   emptyHint     — optional string shown when decks is empty (no upload)

export function deckBracketLabel(d) {
  const wbs = d.wbs || d.bracket || '?'
  const pls = d.pls || null
  return pls ? `B${pls}` : `B${wbs}`
}

function UploadShelfCard({ onClick }) {
  return (
    <div
      onClick={onClick}
      style={{
        cursor: 'pointer',
        border: '2px dashed var(--rule-2)',
        background: 'transparent',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        textAlign: 'center',
        padding: 18,
        minHeight: '100%',
        transition: 'border-color 80ms ease, background 80ms ease, transform 80ms ease',
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.borderColor = 'var(--accent)'
        e.currentTarget.style.background = 'rgba(255,255,255,0.02)'
        e.currentTarget.style.transform = 'translateY(-2px)'
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.borderColor = 'var(--rule-2)'
        e.currentTarget.style.background = 'transparent'
        e.currentTarget.style.transform = 'translateY(0)'
      }}
    >
      <div style={{ fontSize: 48, lineHeight: 1, fontWeight: 900, color: 'var(--accent)' }}>+</div>
      <div style={{
        marginTop: 12,
        fontSize: 13,
        fontWeight: 800,
        letterSpacing: '0.1em',
        textTransform: 'uppercase',
        color: 'var(--ink)',
      }}>ADD YOUR DECK</div>
      <div className="t-xs muted" style={{ marginTop: 6, lineHeight: 1.4 }}>
        PASTE A LIST. FREYA<br />ANALYZES IN SECONDS.
      </div>
    </div>
  )
}

function DeckShelfCard({ deck: d, deckElo, navigate }) {
  const cmdrName = d.commander_card || d.commander
  const bracketLabel = deckBracketLabel(d)
  const deckKey = `${d.owner}/${d.id}`
  const artUrl = cmdrName ? cardArtUrl(cmdrName) : null
  const artContrast = useArtContrast(artUrl)
  // Light commander art (white/cyan/silver themes) makes the white
  // overlay name disappear; flip to dark text + light shadow when the
  // sampled top of the art is bright. Falls back to the original
  // white-on-dark when contrast is unknown (CORS / still loading).
  const titleStyle = artContrast === 'light'
    ? { color: '#0c0d0a', textShadow: '0 1px 3px rgba(255,255,255,0.85)' }
    : { color: '#f4f0e6', textShadow: '0 1px 3px rgba(0,0,0,0.9)' }
  return (
    <div
      key={deckKey}
      onClick={() => navigate(`/decks/${d.owner}/${d.id}`)}
      data-art-contrast={artContrast || undefined}
      style={{
        cursor: 'pointer',
        background: 'var(--panel)',
        border: '1px solid var(--rule-2)',
        display: 'flex',
        flexDirection: 'column',
        transition: 'transform 80ms ease, border-color 80ms ease',
        ...(artContrast ? { '--art-contrast': artContrast } : null),
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.borderColor = 'var(--accent)'
        e.currentTarget.style.transform = 'translateY(-2px)'
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.borderColor = 'var(--rule-2)'
        e.currentTarget.style.transform = 'translateY(0)'
      }}
    >
      <div
        className={cmdrName ? '' : 'hatch'}
        style={{ aspectRatio: '5/4', position: 'relative', overflow: 'hidden', background: 'var(--bg-2)' }}
      >
        {cmdrName ? (
          <img
            src={cardArtUrl(cmdrName)}
            alt={cmdrName}
            loading="lazy"
            style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }}
            onError={(e) => {
              e.target.style.display = 'none'
              e.target.parentElement.classList.add('hatch')
            }}
          />
        ) : (
          <span style={{ position: 'absolute', top: 6, left: 8, fontSize: 9, letterSpacing: '0.1em', color: 'var(--ink-3)' }}>NO ART</span>
        )}
        <div
          style={{
            position: 'absolute',
            inset: 0,
            background: artContrast === 'light'
              ? 'linear-gradient(to bottom, rgba(255,255,255,0) 45%, rgba(255,255,255,0.7) 100%)'
              : 'linear-gradient(to bottom, rgba(0,0,0,0) 45%, rgba(0,0,0,0.78) 100%)',
            pointerEvents: 'none',
          }}
        />
        <span
          style={{
            position: 'absolute',
            top: 6,
            left: 6,
            background: 'var(--inv-bg)',
            color: 'var(--inv-ink)',
            padding: '2px 6px',
            fontSize: 9,
            fontWeight: 700,
            letterSpacing: '0.08em',
          }}
        >
          {bracketLabel}
          {d.legal != null && (
            <span style={{ marginLeft: 4, color: d.legal ? 'var(--ok)' : 'var(--danger)' }}>{d.legal ? '✓' : '✗'}</span>
          )}
        </span>
        <span
          style={{
            position: 'absolute',
            top: 6,
            right: 6,
            background: 'rgba(0,0,0,0.6)',
            color: 'var(--ink)',
            padding: '2px 6px',
            fontSize: 9,
            letterSpacing: '0.08em',
          }}
        >
          {d.owner?.toUpperCase()}
        </span>
        <div
          style={{
            position: 'absolute',
            bottom: 8,
            left: 10,
            right: 10,
            ...titleStyle,
          }}
        >
          <div style={{ fontSize: 13, fontWeight: 700, lineHeight: 1.15, letterSpacing: '0.02em' }}>
            {d.name || cmdrName}
          </div>
          {cmdrName && cmdrName.toUpperCase() !== (d.name || '').toUpperCase() && (
            <div style={{ fontSize: 10, marginTop: 2, opacity: 0.85 }}>{cmdrName}</div>
          )}
        </div>
      </div>
      <div
        style={{
          padding: '6px 10px',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          borderTop: '1px solid var(--rule-2)',
          fontSize: 10,
          letterSpacing: '0.06em',
        }}
      >
        <span className="muted">{d.card_count || d.cardCount || 0} CARDS</span>
        {deckElo && deckElo.games > 0 ? (
          <span>
            <span style={{ fontWeight: 700 }}>{Math.round(deckElo.rating)}</span>
            <span className="muted"> · </span>
            <span style={{ color: 'var(--ok)' }}>{deckElo.wins}W</span>
            <span className="muted">/</span>
            <span style={{ color: 'var(--danger)' }}>{deckElo.losses}L</span>
          </span>
        ) : (
          <span className="muted">UNRATED</span>
        )}
      </div>
    </div>
  )
}

export default function DeckShelf({ decks, eloByDeckId = {}, navigate, onAddCard, emptyHint }) {
  if ((!decks || decks.length === 0) && !onAddCard) {
    return emptyHint ? (
      <div className="t-md muted" style={{ textAlign: 'center', padding: 36 }}>
        &gt; {emptyHint}
      </div>
    ) : null
  }
  return (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))',
        gap: 14,
      }}
    >
      {onAddCard && <UploadShelfCard onClick={onAddCard} />}
      {(decks || []).map((d) => {
        const deckKey = `${d.owner}/${d.id}`
        const deckElo = eloByDeckId[deckKey] || eloByDeckId[d.id]
        return (
          <DeckShelfCard key={deckKey} deck={d} deckElo={deckElo} navigate={navigate} />
        )
      })}
    </div>
  )
}
