import { useState, useEffect } from 'react'
import { Panel } from './chrome'
import { api } from '../services/api'

// Backend rarity tiers (Common/Uncommon/Rare/Mythic/Secret) → the user-facing
// "common/uncommon/rare/epic/legendary" labels used by the badge UI.
const RARITY_LABEL = {
  common:    'COMMON',
  uncommon:  'UNCOMMON',
  rare:      'RARE',
  mythic:    'EPIC',
  secret:    'LEGENDARY',
}

// Rarity sort order — rarest first when picking the showcase strip.
const RARITY_RANK = {
  secret: 5, mythic: 4, rare: 3, uncommon: 2, common: 1,
}

function rarityClass(rarity) {
  return `badge-tile--${(rarity || 'common').toLowerCase()}`
}

function BadgeTile({ badge, awardedAt, size = 'md' }) {
  const r = (badge.rarity || 'common').toLowerCase()
  const stamped = awardedAt ? new Date(awardedAt) : null
  return (
    <div
      className={`badge-tile ${rarityClass(r)}${size === 'sm' ? ' badge-tile--sm' : ''}`}
      title={`${badge.name} — ${badge.description}${stamped ? `\nEarned ${stamped.toLocaleDateString()}` : ''}`}
    >
      <div className="badge-tile__icon">{badge.icon || '★'}</div>
      <div className="badge-tile__name">{badge.name}</div>
      {size !== 'sm' && (
        <>
          <div className="badge-tile__desc">{badge.description}</div>
          <div className="badge-tile__rarity">{RARITY_LABEL[r] || r.toUpperCase()}</div>
        </>
      )}
    </div>
  )
}

function useAchievements(owner) {
  const [snap, setSnap] = useState(null)
  const [error, setError] = useState(null)
  useEffect(() => {
    if (!owner) return
    let cancelled = false
    api.getAchievements(owner)
      .then(s => { if (!cancelled) setSnap(s) })
      .catch(e => { if (!cancelled) setError(e) })
    return () => { cancelled = true }
  }, [owner])
  return { snap, error }
}

export function AchievementsPanel({ owner, code = '04.AB', title }) {
  const { snap, error } = useAchievements(owner)
  if (!owner) return null

  const badges = snap?.badges || []
  const games = snap?.total_games || 0
  const wins = snap?.total_wins || 0
  const streak = snap?.current_win_streak || 0
  const opponents = snap?.opponents_faced || 0

  const sorted = [...badges].sort((a, b) => {
    const rd = (RARITY_RANK[(b.rarity || 'common').toLowerCase()] || 0)
              - (RARITY_RANK[(a.rarity || 'common').toLowerCase()] || 0)
    if (rd !== 0) return rd
    return (a.name || '').localeCompare(b.name || '')
  })

  const heading = title || `ACHIEVEMENTS / / ${owner.toUpperCase()}`

  return (
    <Panel code={code} title={heading} right={
      <span className="t-xs muted">{badges.length} EARNED</span>
    }>
      <div className="badge-stats">
        <span><span className="muted">GAMES</span> {games.toLocaleString()}</span>
        <span><span className="muted">WINS</span> {wins.toLocaleString()}</span>
        <span><span className="muted">STREAK</span> {streak}</span>
        <span><span className="muted">FACED</span> {opponents}</span>
      </div>

      {error && (
        <div className="t-xs" style={{ color: 'var(--danger)', padding: '8px 0' }}>
          &gt; FAILED TO LOAD ACHIEVEMENTS
        </div>
      )}

      {!error && badges.length === 0 && (
        <div className="t-md muted" style={{ padding: '18px 0', textAlign: 'center', letterSpacing: '0.04em' }}>
          &gt; NO BADGES EARNED YET — KEEP PLAYING!
        </div>
      )}

      {badges.length > 0 && (
        <div className="badge-grid">
          {sorted.map((b, i) => (
            <BadgeTile key={b.id || i} badge={b} awardedAt={b.awarded_at} />
          ))}
        </div>
      )}
    </Panel>
  )
}

export function BadgeShowcase({ owner, max = 5 }) {
  const { snap } = useAchievements(owner)
  if (!owner) return null

  const badges = snap?.badges || []
  if (badges.length === 0) {
    return (
      <div className="badge-showcase">
        <span className="t-xs muted">BADGES</span>
        <span className="t-xs muted-2">NONE YET</span>
      </div>
    )
  }

  const top = [...badges].sort((a, b) =>
    (RARITY_RANK[(b.rarity || 'common').toLowerCase()] || 0)
    - (RARITY_RANK[(a.rarity || 'common').toLowerCase()] || 0)
  ).slice(0, max)

  return (
    <div className="badge-showcase">
      <span className="t-xs muted">BADGES</span>
      <div className="badge-showcase__strip">
        {top.map((b, i) => (
          <BadgeTile key={b.id || i} badge={b} awardedAt={b.awarded_at} size="sm" />
        ))}
        {badges.length > max && (
          <span className="t-xs muted-2" style={{ alignSelf: 'center' }}>+{badges.length - max}</span>
        )}
      </div>
    </div>
  )
}

export default AchievementsPanel
