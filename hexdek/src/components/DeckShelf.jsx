import { cardArtUrl } from '../services/api'
import { useArtContrast } from '../hooks/useArtContrast'

// Compact game-count formatter: 30957 → "31K", 1234567 → "1.2M". Used
// on deck tiles where the raw int blows the row width on mobile.
function humanizeGames(n) {
  if (n == null) return '—'
  if (n < 1000) return `${n} GAMES`
  if (n < 100_000) return `${(n / 1000).toFixed(1).replace(/\.0$/, '')}K GAMES`
  if (n < 1_000_000) return `${Math.round(n / 1000)}K GAMES`
  return `${(n / 1_000_000).toFixed(1).replace(/\.0$/, '')}M GAMES`
}

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

function DeckShelfCard({ deck: d, deckElo, navigate, siblingIndex, siblingCount }) {
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
          <div
            title={d.name || cmdrName}
            style={{
              fontSize: 13, fontWeight: 700, lineHeight: 1.15, letterSpacing: '0.02em',
              display: '-webkit-box',
              WebkitLineClamp: 2,
              WebkitBoxOrient: 'vertical',
              overflow: 'hidden',
              overflowWrap: 'anywhere',
            }}
          >
            {d.name || cmdrName}
          </div>
          {cmdrName && cmdrName.toUpperCase() !== (d.name || '').toUpperCase() && (
            <div
              title={cmdrName}
              style={{
                fontSize: 10, marginTop: 2, opacity: 0.85,
                display: '-webkit-box',
                WebkitLineClamp: 1,
                WebkitBoxOrient: 'vertical',
                overflow: 'hidden',
                overflowWrap: 'anywhere',
              }}
            >{cmdrName}</div>
          )}
          {/* Sibling disambiguator: when the same owner has multiple decks
              of this commander, show "v3 of 7" so the tiles are visually
              distinguishable even when name + art are identical. */}
          {siblingCount > 1 && (
            <div style={{ fontSize: 9, marginTop: 3, opacity: 0.7, letterSpacing: '0.08em' }}>
              V{siblingIndex + 1} OF {siblingCount}
            </div>
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
            <span style={{ color: deckElo.win_rate >= 25 ? 'var(--ok)' : 'var(--danger)', fontWeight: 700 }}>
              {deckElo.win_rate != null ? `${Math.round(deckElo.win_rate)}%` : '—'}
            </span>
            <span className="muted"> · </span>
            <span className="muted">{humanizeGames(deckElo.games)}</span>
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
      {(() => {
        // Build sibling index for the V-of-N disambiguator badge. Two
        // tiles are siblings only when (owner, commander, *and* effective
        // name) all match — i.e. the user hasn't disambiguated them with
        // a custom name. Two different decks of the same commander that
        // the user *did* name distinctly ("Neera Storm" vs "Neera Stax")
        // are treated as independent and don't get the badge — their
        // names already do the disambiguation.
        const groupKey = (d) => {
          const cmdr = (d.commander_card || d.commander || '').toUpperCase()
          const name = (d.name || '').toUpperCase()
          // If the deck name equals the commander, the user didn't pick
          // a distinct name — group these as siblings. Otherwise treat
          // each named deck as its own independent entry.
          const effectiveName = name && name !== cmdr ? name : ''
          return `${d.owner || ''}|${cmdr}|${effectiveName}`
        }
        const counts = new Map()
        const seenCounts = new Map()
        for (const d of decks || []) counts.set(groupKey(d), (counts.get(groupKey(d)) || 0) + 1)
        return (decks || []).map((d) => {
          const deckKey = `${d.owner}/${d.id}`
          const deckElo = eloByDeckId[deckKey] || eloByDeckId[d.id]
          const g = groupKey(d)
          const total = counts.get(g) || 1
          const idx = seenCounts.get(g) || 0
          seenCounts.set(g, idx + 1)
          return (
            <DeckShelfCard
              key={deckKey}
              deck={d}
              deckElo={deckElo}
              navigate={navigate}
              siblingIndex={idx}
              siblingCount={total}
            />
          )
        })
      })()}
    </div>
  )
}
