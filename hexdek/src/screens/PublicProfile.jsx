import { useEffect, useState, useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Panel, KV, Tape, Tag } from '../components/chrome'
import DeckShelf from '../components/DeckShelf'
import { api } from '../services/api'
import { useLiveSocket } from '../hooks/useLiveSocket'
import { countryFlagEmoji } from '../lib/flag'

// PublicProfile — read-only player page at /profile/:owner.
// Aggregates data from three sources:
//   - /api/decks?owner=:owner       (deck collection + earliest imported_at as
//                                     "member since" proxy)
//   - /api/achievements/:owner      (TotalGames/TotalWins/streaks + badges)
//   - useLiveSocket().elo           (per-deck Rating/HexRating + W/L)
//
// The settings page at /profile (no owner) is a separate screen
// (Profile.jsx) — this one renders nothing of the auth user's local
// preferences, only public state.

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

function BadgeTile({ badge, earned, awardedAt }) {
  const rarityColor = RARITY_COLOR[badge.rarity] || 'var(--ink-2)'
  const opacity = earned ? 1 : 0.32
  const filter = earned ? 'none' : 'grayscale(1)'
  return (
    <div
      title={`${badge.name} — ${badge.description}${earned ? `\nEarned ${new Date(awardedAt).toISOString().slice(0, 10)}` : ' (locked)'}`}
      style={{
        border: `1px solid ${earned ? rarityColor : 'var(--rule-2)'}`,
        background: 'var(--panel)',
        padding: '10px 8px',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 4,
        opacity,
      }}
    >
      <span style={{ fontSize: 24, lineHeight: 1, filter }}>{badge.icon || '◆'}</span>
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

export default function PublicProfile() {
  const { owner } = useParams()
  const navigate = useNavigate()
  const { elo } = useLiveSocket()

  const [decks, setDecks] = useState([])
  const [decksLoading, setDecksLoading] = useState(true)
  const [achievements, setAchievements] = useState(null)
  const [achLoading, setAchLoading] = useState(true)
  const [profile, setProfile] = useState(null)

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
    api.getOwnerProfile(owner)
      .then(setProfile)
      .catch(() => setProfile(null))
  }, [owner])

  // Filter ELO entries to this owner so per-deck rating/record overlays
  // populate on the shelf, plus the aggregate W/L summary.
  const ownerElo = useMemo(
    () => (elo || []).filter(e => e.owner?.toLowerCase() === owner?.toLowerCase()),
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

  // Aggregate W/L: prefer the achievements snapshot (TotalGames /
  // TotalWins are the canonical owner-level counts the engine writes
  // per finished game). Fall back to summing per-deck ELO entries when
  // achievements is empty (new player, no game data yet).
  const aggregate = useMemo(() => {
    const a = achievements
    if (a && a.total_games > 0) {
      return {
        games: a.total_games,
        wins: a.total_wins,
        losses: Math.max(0, a.total_games - a.total_wins),
        winRate: a.total_games > 0 ? (a.total_wins / a.total_games) * 100 : 0,
        source: 'achievements',
      }
    }
    let games = 0, wins = 0, losses = 0
    for (const e of ownerElo) {
      games += e.games || 0
      wins += e.wins || 0
      losses += e.losses || 0
    }
    return {
      games, wins, losses,
      winRate: games > 0 ? (wins / games) * 100 : 0,
      source: ownerElo.length > 0 ? 'elo' : 'none',
    }
  }, [achievements, ownerElo])

  // Best ELO across this owner's decks (HexRating preferred — it's the
  // engine's display rating; falls back to TrueSkill conservative).
  const bestRating = useMemo(() => {
    let best = null
    for (const e of ownerElo) {
      if ((e.games || 0) < 5) continue // require minimum sample
      const r = e.hex_rating || e.rating || 0
      if (!best || r > best.rating) best = { rating: r, deck: e }
    }
    return best
  }, [ownerElo])

  if (!owner) {
    return (
      <>
        <Tape left="PROFILE" mid="MISSING OWNER" right="DOC HX-500" />
        <div style={{ padding: 36, textAlign: 'center' }}>
          <div className="t-md muted">&gt; NO OWNER SPECIFIED IN URL.</div>
        </div>
      </>
    )
  }

  const memberSince = fmtMemberSince(decks)
  const upperOwner = owner.toUpperCase()
  const flag = countryFlagEmoji(profile?.country)
  const earnedById = useMemo(() => {
    const m = {}
    for (const b of achievements?.badges || []) m[b.id] = b
    return m
  }, [achievements])
  const catalog = achievements?.catalog || []

  return (
    <>
      <Tape
        left={`PROFILE / / ${upperOwner}${flag ? ' ' + flag : ''}`}
        mid={achLoading || decksLoading ? 'LOADING' : 'PUBLIC RECORD'}
        right="DOC HX-500"
      />

      <div style={{ padding: 18, flex: 1, display: 'flex', flexDirection: 'column', gap: 14, overflow: 'auto' }}>

        {/* Header — name + summary stats */}
        <Panel
          code="USR.0"
          title={
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
              {flag && (
                <span
                  className="player-flag"
                  title={profile?.country}
                  aria-label={`Country: ${profile?.country || 'unknown'}`}
                  style={{ fontSize: 18, lineHeight: 1 }}
                >
                  {flag}
                </span>
              )}
              PLAYER RECORD
            </span>
          }
          right={<Tag solid>{upperOwner}</Tag>}
        >
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 18 }}>
            <KV rows={[
              ['DISPLAY NAME',
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
                  {flag && <span style={{ fontSize: 14, lineHeight: 1 }}>{flag}</span>}
                  {upperOwner}
                </span>,
              ],
              ['COUNTRY', profile?.country || '—'],
              ['MEMBER SINCE', memberSince],
              ['DECKS', String(decks.length)],
              ['OPPONENTS FACED', String(achievements?.opponents_faced ?? '—')],
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
            ]} />
          </div>
          {(achievements?.current_win_streak > 0 || achievements?.max_win_streak > 0) && (
            <div style={{ marginTop: 10, display: 'flex', gap: 14, fontSize: 10, letterSpacing: '0.08em', color: 'var(--ink-2)', textTransform: 'uppercase' }}>
              <span>CURRENT STREAK: <span style={{ color: 'var(--ok)', fontWeight: 700 }}>{achievements?.current_win_streak || 0}W</span></span>
              <span className="muted-2">·</span>
              <span>BEST STREAK: <span style={{ color: 'var(--accent)', fontWeight: 700 }}>{achievements?.max_win_streak || 0}W</span></span>
            </div>
          )}
        </Panel>

        {/* Achievement badges */}
        <Panel code="USR.1" title="ACHIEVEMENTS" right={
          <span className="t-xs muted">
            {achLoading
              ? 'LOADING…'
              : `${achievements?.badges?.length || 0} / ${catalog.length} EARNED`}
          </span>
        }>
          {achLoading ? (
            <div className="t-md muted" style={{ textAlign: 'center', padding: 18 }}>
              &gt; LOADING ACHIEVEMENTS<span className="blink">_</span>
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

        {/* Deck collection */}
        <Panel code="USR.2" title={`${upperOwner}'S DECKS`} right={
          <span className="t-xs muted">{decks.length} BUILDS</span>
        }>
          {decksLoading ? (
            <div className="t-md muted" style={{ textAlign: 'center', padding: 18 }}>
              &gt; LOADING DECK COLLECTION<span className="blink">_</span>
            </div>
          ) : decks.length === 0 ? (
            <div className="t-md muted" style={{ textAlign: 'center', padding: 18 }}>
              &gt; NO PUBLIC DECKS UNDER THIS OWNER.
            </div>
          ) : (
            <DeckShelf
              decks={decks.slice(0, 60)}
              eloByDeckId={eloByDeckId}
              navigate={navigate}
            />
          )}
          {decks.length > 60 && (
            <div className="t-xs muted" style={{ textAlign: 'center', marginTop: 10 }}>
              &gt; SHOWING 60 / {decks.length}
            </div>
          )}
        </Panel>
      </div>
    </>
  )
}
