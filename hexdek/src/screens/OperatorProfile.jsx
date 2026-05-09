import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Panel, KV, Tag, Tape } from '../components/chrome'
import DeckShelf from '../components/DeckShelf'
import { api } from '../services/api'
import { useAuth } from '../context/AuthContext'
import { useLiveSocket } from '../hooks/useLiveSocket'
import { countryFlagEmoji } from '../lib/flag'

// OperatorProfile — first-person profile page at /operator (alias /me),
// only reachable when authenticated. Mirrors PublicProfile's structure
// for the public-facing fields (ID card, decks shelf, achievements) but
// adds two operator-only sections: Match History and Friends. Pulls
// owner from auth/localStorage, so unlike PublicProfile no URL param.

const RARITY_COLOR = {
  common:   'var(--ink-2)',
  uncommon: 'var(--ok)',
  rare:     'var(--warn)',
  mythic:   'var(--danger)',
  secret:   'var(--accent)',
}

function fmtMemberSince(decks) {
  let earliest = null
  for (const d of decks || []) {
    const t = d.imported_at ? new Date(d.imported_at) : null
    if (t && !isNaN(t) && (!earliest || t < earliest)) earliest = t
  }
  if (!earliest) return '—'
  return earliest.toISOString().slice(0, 10)
}

// Country code from the user's locale ("en-US" → "US"), preferring the
// stored profile.country when present (the engine writes it from
// Accept-Language on first sighting; treat as authoritative).
function deriveCountry(profileCountry) {
  if (profileCountry) return profileCountry
  if (typeof navigator === 'undefined') return ''
  const lang = navigator.language || (navigator.languages && navigator.languages[0]) || ''
  const parts = lang.split('-')
  if (parts.length < 2) return ''
  const region = parts[parts.length - 1]
  return region.length === 2 ? region.toUpperCase() : ''
}

function fmtTime(ts) {
  if (!ts) return '—'
  const d = new Date(ts)
  if (isNaN(d)) return '—'
  return d.toISOString().slice(0, 10)
}

function BadgeTile({ badge, earned, awardedAt }) {
  const rarityColor = RARITY_COLOR[badge.rarity] || 'var(--ink-2)'
  return (
    <div
      title={`${badge.name} — ${badge.description}${earned ? `\nEarned ${fmtTime(awardedAt)}` : ' (locked)'}`}
      style={{
        border: `1px solid ${earned ? rarityColor : 'var(--rule-2)'}`,
        background: 'var(--panel)',
        padding: '10px 8px',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 4,
        opacity: earned ? 1 : 0.32,
      }}
    >
      <span style={{ fontSize: 24, lineHeight: 1, filter: earned ? 'none' : 'grayscale(1)' }}>
        {badge.icon || '◆'}
      </span>
      <span style={{
        fontSize: 9, fontWeight: 700, letterSpacing: '0.08em',
        color: earned ? 'var(--ink)' : 'var(--ink-3)',
        textAlign: 'center', lineHeight: 1.15,
      }}>
        {badge.name}
      </span>
      <span style={{
        fontSize: 8, letterSpacing: '0.1em', textTransform: 'uppercase',
        color: earned ? rarityColor : 'var(--ink-3)',
      }}>
        {badge.rarity}
      </span>
    </div>
  )
}

export default function OperatorProfile() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const { elo } = useLiveSocket()

  const baseSlug = useMemo(() => {
    if (typeof localStorage !== 'undefined') {
      const stored = localStorage.getItem('hexdek_owner')
      if (stored) return stored.toLowerCase()
    }
    if (user?.displayName) return user.displayName.toLowerCase()
    if (user?.email) return user.email.split('@')[0].split('.')[0].toLowerCase()
    return ''
  }, [user])

  const owner = useMemo(() => {
    if (!baseSlug) return ''
    const match = (elo || []).find(e => {
      const o = e.owner?.toLowerCase()
      if (!o) return false
      return o === baseSlug || baseSlug.startsWith(o) || o.startsWith(baseSlug)
    })
    return match ? match.owner.toLowerCase() : baseSlug
  }, [baseSlug, elo])

  const [decks, setDecks] = useState([])
  const [decksLoading, setDecksLoading] = useState(true)
  const [achievements, setAchievements] = useState(null)
  const [achLoading, setAchLoading] = useState(true)
  const [profile, setProfile] = useState(null)
  const [ownerStats, setOwnerStats] = useState(null)
  const [games, setGames] = useState([])
  const [gamesLoading, setGamesLoading] = useState(true)
  const [friends, setFriends] = useState([])
  const [contribCredits, setContribCredits] = useState(null)

  useEffect(() => {
    if (!owner) return
    setDecksLoading(true)
    api.getDecks({ owner })
      .then(d => setDecks(Array.isArray(d) ? d : []))
      .catch(() => setDecks([]))
      .finally(() => setDecksLoading(false))
  }, [owner])

  useEffect(() => {
    if (!owner) return
    setAchLoading(true)
    api.getAchievements(owner)
      .then(setAchievements)
      .catch(() => setAchievements(null))
      .finally(() => setAchLoading(false))
  }, [owner])

  useEffect(() => {
    if (!owner) return
    api.getOwnerProfile(owner).then(setProfile).catch(() => setProfile(null))
    api.getOwnerStats(owner).then(setOwnerStats).catch(() => setOwnerStats(null))
  }, [owner])

  useEffect(() => {
    if (!owner) return
    setGamesLoading(true)
    api.getOwnerGames(owner, 20)
      .then(g => setGames(Array.isArray(g) ? g : []))
      .catch(() => setGames([]))
      .finally(() => setGamesLoading(false))
  }, [owner])

  useEffect(() => {
    if (!owner) { setFriends([]); return }
    api.listFriends(owner)
      .then(r => setFriends(Array.isArray(r) ? r : (r?.friends || [])))
      .catch(() => setFriends([]))
  }, [owner])

  // BOINC contributor credits — silent fail when no row yet (the
  // server returns zeros for any owner that hasn't run the
  // hexdek-contrib client). Panel renders inline below decks.
  useEffect(() => {
    if (!owner) { setContribCredits(null); return }
    api.getContribCredits(owner)
      .then(setContribCredits)
      .catch(() => setContribCredits(null))
  }, [owner])

  const ownerElo = useMemo(
    () => (elo || []).filter(e => e.owner?.toLowerCase() === owner),
    [elo, owner],
  )
  const eloByDeckId = useMemo(() => {
    const m = {}
    for (const e of ownerElo) {
      if (e.deck_id) {
        m[`${e.owner}/${e.deck_id}`] = e
        m[e.deck_id] = e
      }
    }
    return m
  }, [ownerElo])

  const aggregate = useMemo(() => {
    if (ownerStats && ownerStats.games > 0) {
      return {
        games: ownerStats.games,
        wins: ownerStats.wins,
        losses: ownerStats.losses,
        winRate: ownerStats.win_rate,
      }
    }
    let g = 0, w = 0, l = 0
    for (const e of ownerElo) { g += e.games || 0; w += e.wins || 0; l += e.losses || 0 }
    return { games: g, wins: w, losses: l, winRate: g > 0 ? (w / g) * 100 : 0 }
  }, [ownerStats, ownerElo])

  const bestRating = useMemo(() => {
    let best = null
    for (const e of ownerElo) {
      if ((e.games || 0) < 5) continue
      const r = e.hex_rating || e.rating || 0
      if (!best || r > best.rating) best = { rating: r, deck: e }
    }
    return best
  }, [ownerElo])

  const country = deriveCountry(profile?.country)
  const flag = countryFlagEmoji(country)
  const memberSince = fmtMemberSince(decks)
  const upperOwner = (owner || 'OPERATOR').toUpperCase()

  useEffect(() => {
    if (!owner) { document.title = 'HEXDEK Operator'; return }
    const rating = bestRating?.rating ? ` · ${Math.round(bestRating.rating)} ELO` : ''
    document.title = `${upperOwner}${rating} — HEXDEK`
  }, [owner, upperOwner, bestRating])

  const earnedById = useMemo(() => {
    const m = {}
    for (const b of achievements?.badges || []) m[b.id] = b
    return m
  }, [achievements])
  const catalog = achievements?.catalog || []

  const myGames = games

  if (!owner) {
    return (
      <>
        <Tape left="OPERATOR / / NO OWNER" mid="SET A DISPLAY NAME" right="DOC HX-501" />
        <div style={{ padding: 36, textAlign: 'center' }}>
          <div className="t-md muted">&gt; SET DISPLAY NAME OR HEXDEK_OWNER FIRST.</div>
        </div>
      </>
    )
  }

  return (
    <>
      <Tape
        left={`OPERATOR / / ${upperOwner}${flag ? ' ' + flag : ''}`}
        mid={achLoading || decksLoading ? 'LOADING' : 'PERSONAL RECORD'}
        right="DOC HX-501"
      />

      <div style={{ padding: 18, display: 'flex', flexDirection: 'column', gap: 14 }}>

        {/* ID card */}
        <Panel
          code="OP.0"
          title={
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
              {flag && (
                <span title={country} aria-label={`Country: ${country}`} style={{ fontSize: 18, lineHeight: 1 }}>
                  {flag}
                </span>
              )}
              OPERATOR ID
            </span>
          }
          right={<Tag solid>{upperOwner}</Tag>}
        >
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 18 }}>
            <KV rows={[
              ['DISPLAY NAME',
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
                  {flag && <span style={{ fontSize: 14, lineHeight: 1 }}>{flag}</span>}
                  {upperOwner}
                </span>,
              ],
              ['EMAIL', user?.email || '—'],
              ['COUNTRY', country || '—'],
              ['MEMBER SINCE', memberSince],
              ['DECKS', String(decks.length)],
            ]} />
            <KV rows={[
              ['TOTAL GAMES', aggregate.games > 0 ? String(aggregate.games) : '—'],
              ['RECORD',
                aggregate.games > 0
                  ? <span>
                      <span style={{ color: 'var(--ok)' }}>{aggregate.wins}W</span>
                      <span className="muted"> / </span>
                      <span style={{ color: 'var(--danger)' }}>{aggregate.losses}L</span>
                    </span>
                  : '—',
              ],
              ['WIN RATE', aggregate.games > 0 ? `${aggregate.winRate.toFixed(1)}%` : '—'],
              ['BEST ELO',
                bestRating
                  ? <span>
                      <span style={{ fontWeight: 700 }}>{Math.round(bestRating.rating)}</span>
                      <span className="muted" style={{ marginLeft: 6, fontSize: 9 }}>
                        {bestRating.deck.commander?.toUpperCase()}
                      </span>
                    </span>
                  : '—',
              ],
              ['FRIENDS', String(friends.length)],
            ]} />
          </div>
        </Panel>

        {/* My Decks */}
        <Panel code="OP.1" title="MY DECKS" right={<span className="t-xs muted">{decks.length} BUILDS</span>}>
          {decksLoading ? (
            <div className="t-md muted" style={{ textAlign: 'center', padding: 18 }}>
              &gt; LOADING DECK COLLECTION<span className="blink">_</span>
            </div>
          ) : decks.length === 0 ? (
            <div className="t-md muted" style={{ textAlign: 'center', padding: 18 }}>
              &gt; NO DECKS YET. IMPORT ONE FROM THE DECKS PAGE.
            </div>
          ) : (
            <DeckShelf decks={decks.slice(0, 60)} eloByDeckId={eloByDeckId} navigate={navigate} />
          )}
        </Panel>

        {/* Match History + Friends side by side */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: 14 }}>
          <Panel code="OP.2" title="MATCH HISTORY" right={<span className="t-xs muted">{myGames.length} RECENT</span>}>
            {gamesLoading ? (
              <div className="t-md muted" style={{ textAlign: 'center', padding: 18 }}>
                &gt; LOADING<span className="blink">_</span>
              </div>
            ) : myGames.length === 0 ? (
              <div className="t-md muted" style={{ textAlign: 'center', padding: 18 }}>
                &gt; NO GAMES RECORDED FOR THIS OPERATOR.
              </div>
            ) : (
              <div>
                <div className="op-match-row" style={{
                  display: 'grid',
                  gap: 8,
                  padding: '4px 0',
                  borderBottom: '1px solid var(--rule-2)',
                  fontSize: 9,
                  letterSpacing: '0.1em',
                  color: 'var(--ink-3)',
                  fontWeight: 700,
                }}>
                  <span>RESULT</span>
                  <span>OPPONENT</span>
                  <span style={{ textAlign: 'right' }}>TURNS</span>
                  <span className="op-match-end" style={{ textAlign: 'right' }}>END</span>
                  <span style={{ textAlign: 'right' }}>DATE</span>
                </div>
                {myGames.map((g) => {
                  const won = g.winner === g.my_seat
                  const drew = g.winner < 0
                  const opponents = (g.opponents || [])
                    .map(c => (c || '').split('//')[0].trim())
                    .filter(Boolean)
                    .join(' / ') || '—'
                  return (
                    <div
                      key={g.game_id}
                      onClick={() => navigate(`/report/${g.game_id}`)}
                      className="op-match-row"
                      style={{
                        display: 'grid',
                        gap: 8,
                        padding: '6px 0',
                        borderBottom: '1px dashed var(--rule)',
                        alignItems: 'center',
                        cursor: 'pointer',
                        fontSize: 11,
                      }}
                    >
                      <Tag solid kind={won ? 'ok' : drew ? 'warn' : 'bad'}>
                        {drew ? 'DRAW' : won ? 'WIN' : 'LOSS'}
                      </Tag>
                      <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {opponents}
                      </span>
                      <span className="t-xs" style={{ textAlign: 'right' }}>T{g.turns ?? '—'}</span>
                      <span className="t-xs muted op-match-end" style={{ textAlign: 'right', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {(g.end_reason || '').replace(/_/g, ' ').slice(0, 6).toUpperCase()}
                      </span>
                      <span className="t-xs muted" style={{ textAlign: 'right' }}>
                        {fmtTime(g.finished_at)}
                      </span>
                    </div>
                  )
                })}
              </div>
            )}
          </Panel>

          <Panel code="OP.3" title="FRIENDS" right={<span className="t-xs muted">{friends.length}</span>}>
            {friends.length === 0 ? (
              <div className="t-md muted" style={{ textAlign: 'center', padding: 18 }}>
                &gt; NO FRIENDS ADDED YET.
              </div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column' }}>
                {friends.map((f, i) => {
                  const slug = (typeof f === 'string' ? f : f.owner || f.slug || '').toLowerCase()
                  if (!slug) return null
                  return (
                    <div
                      key={slug + i}
                      onClick={() => navigate(`/profile/${slug}`)}
                      style={{
                        padding: '6px 0',
                        borderBottom: i < friends.length - 1 ? '1px dashed var(--rule)' : 'none',
                        cursor: 'pointer',
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                      }}
                    >
                      <span className="t-md" style={{ fontWeight: 700, letterSpacing: '0.04em' }}>
                        {slug.toUpperCase()}
                      </span>
                      <span className="t-xs muted">↗</span>
                    </div>
                  )
                })}
              </div>
            )}
          </Panel>

          {/* BOINC contributor credits — only renders when the owner
              has actually contributed (credits or chunks > 0). Silent
              for everyone else so we don't clutter the profile with
              a "you've contributed nothing" message. */}
          {contribCredits && (contribCredits.credits_total > 0 || contribCredits.chunks_completed > 0) && (
            <Panel
              code="OP.5"
              title="DISTRIBUTED COMPUTE"
              right={contribCredits.frozen
                ? <Tag solid kind="warn">FROZEN</Tag>
                : <Tag solid kind="ok">{contribCredits.credits_total.toLocaleString()} CREDITS</Tag>}
            >
              <KV rows={[
                ['CREDITS', contribCredits.credits_total.toLocaleString()],
                ['CHUNKS COMPLETED', contribCredits.chunks_completed.toLocaleString()],
                ['CHUNKS REJECTED', contribCredits.chunks_rejected.toLocaleString()],
                ['GAMES SIMULATED', contribCredits.games_simulated.toLocaleString()],
                ...(contribCredits.frozen ? [['FROZEN', contribCredits.frozen_reason || 'manual']] : []),
              ]} />
              <div className="t-xs muted" style={{ marginTop: 8, lineHeight: 1.5 }}>
                &gt; CONTRIBUTING VIA <span style={{ color: 'var(--ink)' }}>HEXDEK-CONTRIB</span> — game simulations run on your machine, validated server-side.
              </div>
            </Panel>
          )}
        </div>

        {/* Achievement showcase */}
        <Panel code="OP.4" title="ACHIEVEMENTS" right={
          <span className="t-xs muted">
            {achLoading
              ? 'LOADING…'
              : `${achievements?.badges?.length || 0} / ${catalog.length} EARNED`}
          </span>
        }>
          {achLoading ? (
            <div className="t-md muted" style={{ textAlign: 'center', padding: 18 }}>
              &gt; LOADING<span className="blink">_</span>
            </div>
          ) : catalog.length === 0 ? (
            <div className="t-md muted" style={{ textAlign: 'center', padding: 18 }}>
              &gt; NO ACHIEVEMENT CATALOG AVAILABLE.
            </div>
          ) : (
            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(110px, 1fr))',
              gap: 8,
            }}>
              {catalog.map(badge => {
                const earned = earnedById[badge.id]
                return (
                  <BadgeTile
                    key={badge.id}
                    badge={badge}
                    earned={!!earned}
                    awardedAt={earned?.awarded_at}
                  />
                )
              })}
            </div>
          )}
        </Panel>
      </div>
    </>
  )
}
