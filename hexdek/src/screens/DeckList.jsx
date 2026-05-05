import { useState, useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Tag, Tape } from '../components/chrome'
import { api, cardArtUrl } from '../services/api'
import { useAuth } from '../context/AuthContext'
import { useLiveSocket } from '../hooks/useLiveSocket'
import { useUploadDeck } from '../hooks/useUploadDeck'
import { MOCK_DECKS } from '../services/mock'

const VIEW_KEY = 'hexdek_deck_view'

export default function DeckList() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [decks, setDecks] = useState([])
  const [filter, setFilter] = useState(searchParams.get('q') || '')
  const ownerParam = searchParams.get('owner') || ''
  const containsParam = searchParams.get('contains') || ''
  const [tab, setTab] = useState(
    ownerParam || containsParam ? 'all' :
    (searchParams.get('tab') === 'all' ? 'all' : 'mine')
  )
  const [legalFilter, setLegalFilter] = useState('all')
  const [loading, setLoading] = useState(true)
  const [viewMode, setViewMode] = useState(() => {
    if (typeof localStorage === 'undefined') return 'shelf'
    return localStorage.getItem(VIEW_KEY) === 'list' ? 'list' : 'shelf'
  })
  const navigate = useNavigate()
  const { user } = useAuth()
  const { elo } = useLiveSocket()
  const upload = useUploadDeck(() => loadDecks())

  useEffect(() => {
    if (typeof localStorage !== 'undefined') localStorage.setItem(VIEW_KEY, viewMode)
  }, [viewMode])

  useEffect(() => {
    const t = searchParams.get('tab')
    if (t === 'all' || t === 'mine') setTab(t)
  }, [searchParams])

  const loadDecks = () => {
    setLoading(true)
    api.getDecks({ owner: ownerParam, contains: containsParam })
      .then(setDecks)
      .catch(() => setDecks(MOCK_DECKS.map(d => ({ ...d, owner: 'josh' }))))
      .finally(() => setLoading(false))
  }

  useEffect(() => { loadDecks() }, [ownerParam, containsParam])

  const eloByDeckId = {}
  for (const e of elo) {
    if (e.deck_id) eloByDeckId[e.deck_id] = e
  }

  const storedOwner = typeof localStorage !== 'undefined' ? localStorage.getItem('hexdek_owner') : null
  const emailPrefix = user?.email?.split('@')[0]?.split('.')[0]?.toLowerCase() || ''
  const myName = storedOwner || user?.displayName?.toLowerCase() || emailPrefix || ''
  const myDecks = myName ? decks.filter(d => d.owner?.toLowerCase() === myName) : []
  const hasMyDecks = myDecks.length > 0

  const baseDecks = (tab === 'mine' && hasMyDecks) ? myDecks : decks
  const hasLegalityData = decks.some(d => d.legal != null)
  const filtered = baseDecks.filter(d => {
    if (legalFilter === 'legal' && d.legal === false) return false
    if (legalFilter === 'illegal' && d.legal !== false) return false
    if (!filter) return true
    const q = filter.toLowerCase()
    const haystack = `${d.name} ${d.commander_card || ''} ${d.commander || ''} ${d.owner || ''}`.toLowerCase()
    return haystack.includes(q)
  })

  const tapeLabel = tab === 'mine' && hasMyDecks
    ? `DECK ARCHIVE / / MY BUILDS`
    : `DECK ARCHIVE / / ALL BUILDS`

  return (
    <>
      <Tape left={tapeLabel} mid={`${filtered.length} / ${decks.length} TOTAL`} right="DOC HX-400" />

      <div style={{ padding: 18, flex: 1, display: 'flex', flexDirection: 'column', gap: 14, overflow: 'auto' }}>
        {(ownerParam || containsParam) && (
          <div style={{ display: 'flex', gap: 10, alignItems: 'center', fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ink-2)' }}>
            <span>FILTER:</span>
            {ownerParam && (
              <Tag solid style={{ cursor: 'pointer' }} onClick={() => {
                const next = new URLSearchParams(searchParams)
                next.delete('owner')
                setSearchParams(next, { replace: true })
              }}>OWNER · {ownerParam.toUpperCase()} ✕</Tag>
            )}
            {containsParam && (
              <Tag solid style={{ cursor: 'pointer' }} onClick={() => {
                const next = new URLSearchParams(searchParams)
                next.delete('contains')
                setSearchParams(next, { replace: true })
              }}>CONTAINS · {containsParam.toUpperCase()} ✕</Tag>
            )}
          </div>
        )}
        {/* Tabs + Search */}
        <div style={{ display: 'flex', gap: 10, alignItems: 'center', flexWrap: 'wrap' }}>
          {hasMyDecks && (
            <>
              <Tag solid={tab === 'mine'} onClick={() => setTab('mine')} style={{ cursor: 'pointer' }}>MY DECKS</Tag>
              <Tag solid={tab === 'all'} onClick={() => setTab('all')} style={{ cursor: 'pointer' }}>ALL DECKS</Tag>
              <div style={{ width: 1, height: 16, background: 'var(--rule-2)' }} />
            </>
          )}
          {hasLegalityData && (
            <>
              <Tag solid={legalFilter === 'all'} onClick={() => setLegalFilter('all')} style={{ cursor: 'pointer' }}>ALL</Tag>
              <Tag solid={legalFilter === 'legal'} onClick={() => setLegalFilter('legal')} style={{ cursor: 'pointer', color: legalFilter === 'legal' ? undefined : 'var(--ok)' }}>✓ LEGAL</Tag>
              <Tag solid={legalFilter === 'illegal'} onClick={() => setLegalFilter('illegal')} style={{ cursor: 'pointer', color: legalFilter === 'illegal' ? undefined : 'var(--danger)' }}>✗ ILLEGAL</Tag>
              <div style={{ width: 1, height: 16, background: 'var(--rule-2)' }} />
            </>
          )}
          <div className="panel" style={{ padding: 0, flex: 1, minWidth: 200, borderStyle: filter ? 'solid' : 'dashed' }}>
            <input
              type="text"
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              placeholder="SEARCH DECKS..."
              style={{
                width: '100%',
                padding: '8px 12px',
                background: 'transparent',
                border: 'none',
                color: 'var(--ink)',
                fontFamily: 'inherit',
                fontSize: 11,
                letterSpacing: '0.06em',
                textTransform: 'uppercase',
                outline: 'none',
              }}
            />
          </div>
          <span className="t-xs muted">{filtered.length} MATCHES</span>
          <div style={{ width: 1, height: 16, background: 'var(--rule-2)' }} />
          <Tag solid={viewMode === 'shelf'} onClick={() => setViewMode('shelf')} style={{ cursor: 'pointer' }}>SHELF</Tag>
          <Tag solid={viewMode === 'list'} onClick={() => setViewMode('list')} style={{ cursor: 'pointer' }}>LIST</Tag>
        </div>

        {/* Deck grid */}
        {loading ? (
          <div className="t-md muted" style={{ textAlign: 'center', padding: 36 }}>&gt; LOADING DECK ARCHIVE<span className="blink">_</span></div>
        ) : viewMode === 'shelf' ? (
          <ShelfView decks={filtered.slice(0, 60)} eloByDeckId={eloByDeckId} navigate={navigate} onUpload={upload.open} />
        ) : (
          <ListView decks={filtered.slice(0, 60)} eloByDeckId={eloByDeckId} navigate={navigate} onUpload={upload.open} />
        )}

        {filtered.length > 60 && (
          <div className="t-xs muted" style={{ textAlign: 'center' }}>
            &gt; SHOWING 60 / {filtered.length} — REFINE SEARCH TO SEE MORE
          </div>
        )}
      </div>

      {upload.modal}
    </>
  )
}

function deckBracketLabel(d) {
  const wbs = d.wbs || d.bracket || '?'
  const pls = d.pls || null
  return pls ? `B${pls}` : `B${wbs}`
}

function UploadShelfCard({ onUpload }) {
  return (
    <div
      onClick={onUpload}
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

function ShelfView({ decks, eloByDeckId, navigate, onUpload }) {
  return (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))',
        gap: 14,
      }}
    >
      <UploadShelfCard onUpload={onUpload} />
      {decks.map((d) => {
        const deckKey = `${d.owner}/${d.id}`
        const deckElo = eloByDeckId[deckKey] || eloByDeckId[d.id]
        const cmdrName = d.commander_card || d.commander
        const bracketLabel = deckBracketLabel(d)
        return (
          <div
            key={deckKey}
            onClick={() => navigate(`/decks/${d.owner}/${d.id}`)}
            style={{
              cursor: 'pointer',
              background: 'var(--panel)',
              border: '1px solid var(--rule-2)',
              display: 'flex',
              flexDirection: 'column',
              transition: 'transform 80ms ease, border-color 80ms ease',
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
                  background: 'linear-gradient(to bottom, rgba(0,0,0,0) 45%, rgba(0,0,0,0.78) 100%)',
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
                  color: '#f4f0e6',
                  textShadow: '0 1px 3px rgba(0,0,0,0.9)',
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
      })}
    </div>
  )
}

function ListView({ decks, eloByDeckId, navigate, onUpload }) {
  return (
    <div className="panel" style={{ padding: 0 }}>
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: '40px 1fr 1fr 80px 60px 70px 100px',
          gap: 8,
          padding: '6px 10px',
          borderBottom: '1px solid var(--rule-2)',
          fontSize: 9,
          letterSpacing: '0.1em',
          color: 'var(--ink-3)',
          fontWeight: 700,
        }}
      >
        <span></span>
        <span>NAME</span>
        <span>COMMANDER</span>
        <span>OWNER</span>
        <span>BRACKET</span>
        <span>ELO</span>
        <span>RECORD</span>
      </div>
      <div
        onClick={onUpload}
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          gap: 10,
          padding: '10px',
          borderBottom: '2px dashed var(--rule-2)',
          background: 'transparent',
          cursor: 'pointer',
          color: 'var(--ink)',
          fontWeight: 800,
          letterSpacing: '0.1em',
          fontSize: 12,
          textTransform: 'uppercase',
          transition: 'background 80ms ease, color 80ms ease',
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.background = 'var(--accent)'
          e.currentTarget.style.color = 'var(--bg)'
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.background = 'transparent'
          e.currentTarget.style.color = 'var(--ink)'
        }}
      >
        <span style={{ fontSize: 18, lineHeight: 1, fontWeight: 900 }}>+</span>
        <span>ADD YOUR DECK</span>
      </div>
      {decks.map((d) => {
        const deckKey = `${d.owner}/${d.id}`
        const deckElo = eloByDeckId[deckKey] || eloByDeckId[d.id]
        const cmdrName = d.commander_card || d.commander
        const bracketLabel = deckBracketLabel(d)
        return (
          <div
            key={deckKey}
            onClick={() => navigate(`/decks/${d.owner}/${d.id}`)}
            style={{
              display: 'grid',
              gridTemplateColumns: '40px 1fr 1fr 80px 60px 60px 100px',
              gap: 8,
              padding: '6px 10px',
              borderBottom: '1px solid var(--rule)',
              alignItems: 'center',
              cursor: 'pointer',
              fontSize: 11,
            }}
            onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--panel-2)' }}
            onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent' }}
          >
            <div
              className={cmdrName ? '' : 'hatch'}
              style={{
                width: 40,
                height: 28,
                overflow: 'hidden',
                border: '1px solid var(--rule-2)',
                background: 'var(--bg-2)',
              }}
            >
              {cmdrName && (
                <img
                  src={cardArtUrl(cmdrName)}
                  alt=""
                  loading="lazy"
                  style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }}
                  onError={(e) => {
                    e.target.style.display = 'none'
                    e.target.parentElement.classList.add('hatch')
                  }}
                />
              )}
            </div>
            <span style={{ fontWeight: 600, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{d.name || cmdrName}</span>
            <span className="muted" style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{cmdrName || '—'}</span>
            <span className="t-xs">{d.owner?.toUpperCase()}</span>
            <span style={{ fontWeight: 700, letterSpacing: '0.06em' }}>
              {bracketLabel}
              {d.legal != null && (
                <span style={{ marginLeft: 4, color: d.legal ? 'var(--ok)' : 'var(--danger)', fontSize: 9 }}>{d.legal ? '✓' : '✗'}</span>
              )}
            </span>
            <span style={{ fontWeight: 700 }}>{deckElo ? Math.round(deckElo.rating) : '—'}</span>
            <span className="t-xs">
              {deckElo && deckElo.games > 0 ? (
                <>
                  <span style={{ color: 'var(--ok)' }}>{deckElo.wins}</span>
                  <span className="muted"> · </span>
                  <span style={{ color: 'var(--danger)' }}>{deckElo.losses}</span>
                  <span className="muted"> ({deckElo.win_rate}%)</span>
                </>
              ) : (
                <span className="muted">—</span>
              )}
            </span>
          </div>
        )
      })}
    </div>
  )
}
