import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { Tape, Tag, ConfidenceDots } from '../components/chrome'
import { useLiveSocket } from '../hooks/useLiveSocket'
import { cardArtUrl } from '../services/api'

const SORT_KEYS = [
  { key: 'hex_rating', label: 'HEXELO' },
  { key: 'mu', label: 'TS μ' },
  { key: 'games', label: 'GAMES' },
  { key: 'win_rate', label: 'WIN %' },
  { key: 'wins', label: 'WINS' },
  { key: 'losses', label: 'LOSSES' },
  { key: 'delta', label: 'DELTA' },
]

const BRACKETS = [
  { value: null, label: 'ALL' },
  { value: 1, label: 'B1' },
  { value: 2, label: 'B2' },
  { value: 3, label: 'B3' },
  { value: 4, label: 'B4' },
  { value: 5, label: 'B5' },
]

const BRACKET_KIND = { 1: undefined, 2: 'ok', 3: 'info', 4: 'warn', 5: 'danger' }

function BandTag({ band, bracket }) {
  if (!band) return null
  return (
    <Tag kind={BRACKET_KIND[bracket]} style={{ fontSize: 8, padding: '1px 4px', marginLeft: 4 }}>
      {band}
    </Tag>
  )
}

function shameTier(rating) {
  const r = rating ?? 0
  if (r <= -500) return { label: 'UNINSTALL', kind: 'bad' }
  if (r <= -300) return { label: 'PACK IT UP', kind: 'bad' }
  if (r <= -200) return { label: 'COOKED', kind: 'bad' }
  if (r <= -100) return { label: 'DOWN BAD', kind: 'bad' }
  if (r <= 0) return { label: 'MID', kind: 'warn' }
  return null
}

function ShameBadge({ rating }) {
  const tier = shameTier(rating)
  if (!tier) return null
  return (
    <Tag kind={tier.kind} solid style={{ fontSize: 8, padding: '1px 4px', marginLeft: 4 }}>
      {tier.label}
    </Tag>
  )
}

function DeltaDisplay({ delta }) {
  if (delta == null || delta === 0) return <span className="muted-2">--</span>
  const positive = delta > 0
  return (
    <span style={{ color: positive ? 'var(--ok)' : 'var(--danger)', fontWeight: 600 }}>
      {positive ? '+' : ''}{delta.toFixed(1)}
    </span>
  )
}

function RecordDisplay({ wins, losses }) {
  return (
    <span>
      <span style={{ color: 'var(--ok)' }}>{wins}W</span>
      <span className="muted-2"> - </span>
      <span style={{ color: 'var(--danger)' }}>{losses}L</span>
    </span>
  )
}

export default function Leaderboard() {
  const [filter, setFilter] = useState('')
  const [sortKey, setSortKey] = useState('hex_rating')
  const [sortAsc, setSortAsc] = useState(false)
  const [bracket, setBracket] = useState(null)
  const navigate = useNavigate()
  const { elo } = useLiveSocket()

  const handleSort = (key) => {
    if (sortKey === key) {
      setSortAsc(!sortAsc)
    } else {
      setSortKey(key)
      setSortAsc(false)
    }
  }

  const sorted = useMemo(() => {
    let list = [...(elo || [])]

    if (bracket != null) {
      list = list.filter(e => e.bracket === bracket)
    }

    if (filter) {
      const q = filter.toLowerCase()
      list = list.filter(e => {
        const haystack = `${e.commander || ''} ${e.owner || ''}`.toLowerCase()
        return haystack.includes(q)
      })
    }

    list.sort((a, b) => {
      const av = a[sortKey] ?? 0
      const bv = b[sortKey] ?? 0
      const cmp = sortAsc ? av - bv : bv - av
      if (cmp !== 0) return cmp
      return (b.hex_rating ?? 0) - (a.hex_rating ?? 0)
    })

    return list
  }, [elo, filter, sortKey, sortAsc, bracket])

  const wallOfShame = useMemo(() => {
    let list = [...(elo || [])]
    if (bracket != null) list = list.filter(e => e.bracket === bracket)
    list = list.filter(e => (e.rating ?? 0) <= 0)
    list.sort((a, b) => (a.rating ?? 0) - (b.rating ?? 0))
    return list.slice(0, 10)
  }, [elo, bracket])

  const sortArrow = (key) => {
    if (sortKey !== key) return ''
    return sortAsc ? ' ▲' : ' ▼'
  }

  const handleRowClick = (entry) => {
    if (entry.owner && entry.deck_id) {
      navigate(`/decks/${entry.owner}/${entry.deck_id}`)
    } else if (entry.deck_id) {
      const parts = entry.deck_id.split('/')
      if (parts.length === 2) {
        navigate(`/decks/${parts[0]}/${parts[1]}`)
      } else {
        navigate(`/decks?q=${encodeURIComponent(entry.commander || entry.deck_id)}`)
      }
    } else if (entry.commander) {
      navigate(`/decks?q=${encodeURIComponent(entry.commander)}`)
    }
  }

  const bracketLabel = bracket != null ? `B${bracket}` : 'ALL'

  return (
    <>
      <Tape left={`LEADERBOARD / ${bracketLabel} / LIVE RANKINGS`} mid={`${sorted.length} DECKS`} right="DOC HX-500" />

      <div style={{ padding: 18, flex: 1, display: 'flex', flexDirection: 'column', gap: 14, overflow: 'auto' }}>
        {/* Search + Sort controls */}
        <div style={{ display: 'flex', gap: 10, alignItems: 'center', flexWrap: 'wrap' }}>
          <div className="panel" style={{ padding: 0, flex: 1, minWidth: 200, borderStyle: filter ? 'solid' : 'dashed' }}>
            <input
              type="text"
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              placeholder="SEARCH COMMANDER OR OWNER..."
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
          <span className="t-xs muted">{sorted.length} RANKED</span>
        </div>

        {/* Bracket filter */}
        <div className="lb-sort-bar">
          {BRACKETS.map(b => (
            <Tag
              key={b.label}
              kind={bracket === b.value ? (BRACKET_KIND[b.value] || 'ok') : undefined}
              solid={bracket === b.value}
              onClick={() => setBracket(b.value)}
              style={{ cursor: 'pointer', fontSize: 9 }}
            >
              {b.label}
            </Tag>
          ))}
        </div>

        {/* Sort controls */}
        <div className="lb-sort-bar">
          {SORT_KEYS.map(sk => (
            <Tag
              key={sk.key}
              kind={sortKey === sk.key ? 'ok' : undefined}
              solid={sortKey === sk.key}
              onClick={() => handleSort(sk.key)}
              style={{ cursor: 'pointer', fontSize: 9 }}
            >
              {sk.label}{sortArrow(sk.key)}
            </Tag>
          ))}
        </div>

        {/* Desktop table */}
        <div className="lb-table-wrap">
          <table className="lb-table">
            <thead>
              <tr>
                <th className="lb-th lb-th--rank">#</th>
                <th className="lb-th lb-th--cmdr">COMMANDER</th>
                <th className="lb-th lb-th--owner">OWNER</th>
                {SORT_KEYS.map(sk => (
                  <th
                    key={sk.key}
                    className={`lb-th lb-th--${sk.key} lb-th--sortable${sortKey === sk.key ? ' lb-th--active' : ''}`}
                    onClick={() => handleSort(sk.key)}
                  >
                    {sk.label}{sortArrow(sk.key)}
                  </th>
                ))}
                <th className="lb-th lb-th--conf">CONF</th>
                <th className="lb-th lb-th--record">RECORD</th>
              </tr>
            </thead>
            <tbody>
              {sorted.map((entry, i) => (
                <tr
                  key={entry.deck_id || i}
                  className="lb-row"
                  onClick={() => handleRowClick(entry)}
                >
                  <td className="lb-td lb-td--rank">
                    <span className={i < 3 ? 'lb-medal' : ''}>{i + 1}</span>
                  </td>
                  <td className="lb-td lb-td--cmdr" style={{ textDecoration: 'underline', textDecorationColor: 'var(--rule-2)' }}>
                    {entry.commander || '--'}
                    <BandTag band={entry.band} bracket={entry.bracket} />
                    <ShameBadge rating={entry.rating} />
                  </td>
                  <td className="lb-td lb-td--owner muted">
                    {entry.owner ? (
                      <a
                        onClick={(e) => { e.stopPropagation(); navigate(`/profile/${entry.owner}`) }}
                        style={{ cursor: 'pointer', color: 'var(--ink-2)', textDecoration: 'none', borderBottom: '1px dotted var(--ink-3)' }}
                      >
                        {entry.owner.toUpperCase()}
                      </a>
                    ) : '--'}
                  </td>
                  <td className="lb-td lb-td--rating">
                    <span style={{ fontWeight: 700 }}>{Math.round(entry.hex_rating || 0)}</span>
                  </td>
                  <td className="lb-td lb-td--mu">
                    <span className="t-xs muted-2">{Math.round(entry.mu || 0)}</span>
                  </td>
                  <td className="lb-td lb-td--games">{entry.games || 0}</td>
                  <td className="lb-td lb-td--winrate">
                    {entry.win_rate != null ? `${entry.win_rate.toFixed(1)}%` : '--'}
                  </td>
                  <td className="lb-td lb-td--delta">
                    <DeltaDisplay delta={entry.delta} />
                  </td>
                  <td className="lb-td lb-td--conf">
                    <ConfidenceDots games={entry.games} />
                  </td>
                  <td className="lb-td lb-td--record">
                    <RecordDisplay wins={entry.wins || 0} losses={entry.losses || 0} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {sorted.length === 0 && (
            <div className="t-md muted" style={{ textAlign: 'center', padding: 36 }}>
              {elo?.length === 0
                ? <>&gt; AWAITING ELO DATA<span className="blink">_</span></>
                : <>&gt; NO MATCHES FOUND</>
              }
            </div>
          )}
        </div>

        {/* Mobile card layout */}
        <div className="lb-cards">
          {sorted.map((entry, i) => {
            const artUrl = cardArtUrl(entry.commander_card || entry.commander)
            return (
              <div
                key={entry.deck_id || i}
                className="panel lb-card"
                style={{ padding: 0, cursor: 'pointer' }}
                onClick={() => handleRowClick(entry)}
              >
                <div className="lb-card-row">
                  <div className="lb-card-art">
                    {artUrl && (
                      <img
                        src={artUrl}
                        alt={entry.commander || ''}
                        onError={e => { e.target.style.display = 'none'; e.target.parentElement.classList.add('hatch') }}
                      />
                    )}
                    <span className={`lb-card-rank${i < 3 ? ' lb-medal' : ''}`}>#{i + 1}</span>
                  </div>
                  <div className="lb-card-body">
                    <div className="panel-hd">
                      <span style={{ fontWeight: 700, color: 'var(--ink)', minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {entry.commander || '--'}
                        <BandTag band={entry.band} bracket={entry.bracket} />
                        <ShameBadge rating={entry.rating} />
                      </span>
                      <span style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', lineHeight: 1.1 }}>
                        <span className="t-xs" style={{ fontWeight: 700 }}>
                          HexELO {Math.round(entry.hex_rating || 0)}
                        </span>
                        <span className="t-xs muted-2">
                          TS μ {Math.round(entry.mu || 0)}
                        </span>
                      </span>
                    </div>
                    <div style={{ padding: '8px 10px', display: 'flex', flexDirection: 'column', gap: 6 }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        {entry.owner ? (
                          <a
                            onClick={(e) => { e.stopPropagation(); navigate(`/profile/${entry.owner}`) }}
                            className="t-xs muted"
                            style={{ cursor: 'pointer', textDecoration: 'none', borderBottom: '1px dotted var(--ink-3)' }}
                          >
                            {entry.owner.toUpperCase()}
                          </a>
                        ) : <span className="t-xs muted">--</span>}
                        <ConfidenceDots games={entry.games} showLabel />
                      </div>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <RecordDisplay wins={entry.wins || 0} losses={entry.losses || 0} />
                        <span className="t-xs">
                          {entry.win_rate != null ? `${entry.win_rate.toFixed(1)}%` : '--'}
                        </span>
                      </div>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <span className="t-xs muted">{entry.games || 0} GAMES</span>
                        <DeltaDisplay delta={entry.delta} />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )
          })}
          {sorted.length === 0 && (
            <div className="t-md muted" style={{ textAlign: 'center', padding: 36 }}>
              {elo?.length === 0
                ? <>&gt; AWAITING ELO DATA<span className="blink">_</span></>
                : <>&gt; NO MATCHES FOUND</>
              }
            </div>
          )}
        </div>

        {wallOfShame.length > 0 && (
          <div style={{ marginTop: 8, display: 'flex', flexDirection: 'column', gap: 8 }}>
            <Tape
              left={`WALL OF SHAME / ${bracketLabel} / BOTTOM ${wallOfShame.length}`}
              mid="NEGATIVE ELO REGISTRY"
              right="DOC HX-666"
            />
            <div className="panel" style={{ padding: 0 }}>
              {wallOfShame.map((entry, i) => (
                <div
                  key={`shame-${entry.deck_id || i}`}
                  className="lb-row"
                  onClick={() => handleRowClick(entry)}
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: '6px 10px',
                    gap: 10,
                    borderTop: i === 0 ? 'none' : '1px solid var(--rule-2)',
                    cursor: 'pointer',
                    flexWrap: 'wrap',
                  }}
                >
                  <span style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap', minWidth: 0 }}>
                    <span className="muted-2" style={{ minWidth: 22, fontWeight: 700 }}>#{i + 1}</span>
                    <span style={{ fontWeight: 700, color: 'var(--ink)' }}>
                      {entry.commander || '--'}
                    </span>
                    <ShameBadge rating={entry.rating} />
                    <span className="t-xs muted">{entry.owner?.toUpperCase() || '--'}</span>
                  </span>
                  <span style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                    <RecordDisplay wins={entry.wins || 0} losses={entry.losses || 0} />
                    <span style={{ color: 'var(--danger)', fontWeight: 700, minWidth: 48, textAlign: 'right' }}>
                      {Math.round(entry.rating || 0)}
                    </span>
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </>
  )
}
