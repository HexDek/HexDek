import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Panel, Tag, Tape, Btn } from '../components/chrome'
import { api, cardArtUrl } from '../services/api'
import { useAuth } from '../context/AuthContext'
import { useArtContrast } from '../hooks/useArtContrast'
import { countryFlagEmoji } from '../lib/flag'
import { toast } from '../components/Toast'

// Friends — /friends. Lightweight pub model: see who you've added,
// peek at their decks and ELO via PublicProfile, search for new
// players to add. No feed, no notifications.
//
// Data flow:
//   - api.listFriends(userOwner)              → ['slug', 'slug', ...]
//   - api.getOwnerProfiles(allSlugs)          → { slug: 'US', ... } for flags
//   - api.getDecks({ owner: slug }) per row   → first deck = commander art tile
//   - api.search(query) filtered to .owners  → search hits
//
// Per-friend deck fetches are sequential by useEffect identity so a
// long friend list doesn't slam the API on mount; results are kept in
// a module-level cache so navigating back doesn't refetch.

const COVER_CACHE = new Map() // owner -> { commander, deckCount }

function deriveOwner(user) {
  if (typeof localStorage !== 'undefined') {
    const stored = localStorage.getItem('hexdek_owner')
    if (stored) return stored.toLowerCase()
  }
  if (user?.displayName) return user.displayName.toLowerCase()
  if (user?.email) return user.email.split('@')[0].split('.')[0].toLowerCase()
  return ''
}

function FriendTile({ owner, onUnfriend, onOpen, country }) {
  const cached = COVER_CACHE.get(owner)
  const [cover, setCover] = useState(cached || null)

  useEffect(() => {
    if (cached) return
    let alive = true
    api.getDecks({ owner })
      .then(decks => {
        if (!alive) return
        const list = Array.isArray(decks) ? decks : []
        const first = list[0]
        const c = {
          commander: first?.commander_card || first?.commander || '',
          deckCount: list.length,
        }
        COVER_CACHE.set(owner, c)
        setCover(c)
      })
      .catch(() => {})
    return () => { alive = false }
  }, [owner, cached])

  const artUrl = cover?.commander ? cardArtUrl(cover.commander) : null
  const artContrast = useArtContrast(artUrl)
  const flag = countryFlagEmoji(country)
  const titleStyle = artContrast === 'light'
    ? { color: '#0c0d0a', textShadow: '0 1px 3px rgba(255,255,255,0.85)' }
    : { color: '#f4f0e6', textShadow: '0 1px 3px rgba(0,0,0,0.9)' }

  return (
    <div
      data-art-contrast={artContrast || undefined}
      style={{
        background: 'var(--panel)',
        border: '1px solid var(--rule-2)',
        display: 'flex',
        flexDirection: 'column',
        cursor: 'pointer',
        ...(artContrast ? { '--art-contrast': artContrast } : null),
      }}
      onClick={() => onOpen(owner)}
    >
      <div
        className={artUrl ? '' : 'hatch'}
        style={{ aspectRatio: '5/4', position: 'relative', overflow: 'hidden', background: 'var(--bg-2)' }}
      >
        {artUrl && (
          <img
            src={artUrl}
            alt=""
            loading="lazy"
            style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }}
            onError={(e) => { e.target.style.display = 'none'; e.target.parentElement.classList.add('hatch') }}
          />
        )}
        <div
          style={{
            position: 'absolute', inset: 0,
            background: artContrast === 'light'
              ? 'linear-gradient(to bottom, rgba(255,255,255,0) 45%, rgba(255,255,255,0.7) 100%)'
              : 'linear-gradient(to bottom, rgba(0,0,0,0) 45%, rgba(0,0,0,0.78) 100%)',
            pointerEvents: 'none',
          }}
        />
        <button
          type="button"
          onClick={(e) => { e.stopPropagation(); onUnfriend(owner) }}
          title={`Unfriend ${owner.toUpperCase()}`}
          style={{
            position: 'absolute', top: 6, right: 6,
            background: 'rgba(0,0,0,0.65)', color: 'var(--ink)',
            border: '1px solid var(--rule-2)',
            padding: '2px 6px', fontSize: 9, fontWeight: 700,
            letterSpacing: '0.08em', cursor: 'pointer',
            fontFamily: 'inherit',
          }}
        >✕</button>
        <div
          style={{
            position: 'absolute', bottom: 8, left: 10, right: 10,
            ...titleStyle,
          }}
        >
          <div style={{ fontSize: 14, fontWeight: 700, lineHeight: 1.15, letterSpacing: '0.02em', display: 'flex', alignItems: 'center', gap: 6 }}>
            {flag && <span style={{ fontSize: 16, lineHeight: 1 }}>{flag}</span>}
            {owner.toUpperCase()}
          </div>
          {cover?.commander && (
            <div style={{ fontSize: 10, marginTop: 2, opacity: 0.85 }}>
              {cover.commander}
            </div>
          )}
        </div>
      </div>
      <div style={{
        padding: '6px 10px',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        borderTop: '1px solid var(--rule-2)',
        fontSize: 10,
        letterSpacing: '0.06em',
      }}>
        <span className="muted">{cover ? `${cover.deckCount} DECKS` : 'LOADING…'}</span>
        <span style={{ color: 'var(--accent)' }}>OPEN ↗</span>
      </div>
    </div>
  )
}

function SearchHit({ slug, isFriend, isSelf, country, onAdd, navigate }) {
  const flag = countryFlagEmoji(country)
  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      padding: '8px 10px',
      borderBottom: '1px dashed var(--rule-2)',
      gap: 10,
    }}>
      <span
        onClick={() => navigate(`/profile/${slug}`)}
        style={{ cursor: 'pointer', fontWeight: 700, letterSpacing: '0.04em', display: 'flex', alignItems: 'center', gap: 6 }}
      >
        {flag && <span style={{ fontSize: 14, lineHeight: 1 }}>{flag}</span>}
        {slug.toUpperCase()}
        {isSelf && <span className="t-xs muted" style={{ marginLeft: 4 }}>(YOU)</span>}
      </span>
      {isSelf ? (
        <span className="t-xs muted">—</span>
      ) : isFriend ? (
        <Tag solid kind="ok">FRIEND ✓</Tag>
      ) : (
        <Btn sm arrow="+" onClick={() => onAdd(slug)}>ADD FRIEND</Btn>
      )}
    </div>
  )
}

export default function Friends() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const userOwner = useMemo(() => deriveOwner(user), [user])

  const [friends, setFriends] = useState([])
  const [loading, setLoading] = useState(true)
  const [countries, setCountries] = useState({}) // owner -> 'US'
  const [query, setQuery] = useState('')
  const [results, setResults] = useState([])
  const [searching, setSearching] = useState(false)

  // Initial friend list.
  useEffect(() => {
    if (!userOwner) { setFriends([]); setLoading(false); return }
    let alive = true
    setLoading(true)
    api.listFriends(userOwner)
      .then(r => {
        if (!alive) return
        setFriends(Array.isArray(r) ? r : (r?.friends || []))
      })
      .catch(() => { if (alive) setFriends([]) })
      .finally(() => { if (alive) setLoading(false) })
    return () => { alive = false }
  }, [userOwner])

  // Country flags for friends + search hits, batched.
  useEffect(() => {
    const wanted = new Set(friends)
    for (const r of results) wanted.add(r.label)
    const missing = [...wanted].filter(o => o && !(o in countries))
    if (missing.length === 0) return
    api.getOwnerProfiles(missing)
      .then(map => setCountries(prev => ({ ...prev, ...map })))
      .catch(() => {})
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [friends, results])

  // Debounced player search.
  useEffect(() => {
    const q = query.trim()
    if (q.length < 2) { setResults([]); return }
    let alive = true
    setSearching(true)
    const t = setTimeout(() => {
      api.search(q, 12)
        .then(r => {
          if (!alive) return
          const owners = (r?.results?.owners || []).filter(x => x.kind === 'owner')
          setResults(owners)
        })
        .catch(() => { if (alive) setResults([]) })
        .finally(() => { if (alive) setSearching(false) })
    }, 250)
    return () => { alive = false; clearTimeout(t) }
  }, [query])

  const friendsSet = useMemo(() => new Set(friends.map(f => f.toLowerCase())), [friends])

  const handleAdd = (target) => {
    if (!userOwner || !target) return
    api.addFriend(target, userOwner)
      .then(() => {
        setFriends(prev => prev.includes(target) ? prev : [...prev, target])
        toast.success(`ADDED ${target.toUpperCase()}`)
      })
      .catch(() => toast.error('ADD FAILED'))
  }

  const handleUnfriend = (target) => {
    if (!userOwner || !target) return
    api.removeFriend(target, userOwner)
      .then(() => {
        setFriends(prev => prev.filter(f => f !== target))
        toast.info(`REMOVED ${target.toUpperCase()}`)
      })
      .catch(() => toast.error('REMOVE FAILED'))
  }

  const openProfile = (owner) => navigate(`/profile/${owner}`)

  if (!userOwner) {
    return (
      <>
        <Tape left="FRIENDS / / NO OWNER" mid="SET A DISPLAY NAME" right="DOC HX-510" />
        <div style={{ padding: 36, textAlign: 'center' }}>
          <div className="t-md muted">&gt; SIGN IN OR SET HEXDEK_OWNER FIRST.</div>
        </div>
      </>
    )
  }

  return (
    <>
      <Tape
        left={`FRIENDS / / ${userOwner.toUpperCase()}`}
        mid={loading ? 'LOADING' : `${friends.length} CONNECTIONS`}
        right="DOC HX-510"
      />

      <div style={{ padding: 18, flex: 1, display: 'flex', flexDirection: 'column', gap: 14, overflow: 'auto' }}>
        {/* Find players */}
        <Panel
          code="FR.0"
          title="FIND PLAYERS"
          right={<span className="t-xs muted">{searching ? 'SEARCHING…' : `${results.length} HITS`}</span>}
        >
          <div className="panel" style={{ padding: 0, marginBottom: 8, borderStyle: query ? 'solid' : 'dashed' }}>
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="SEARCH PLAYERS BY NAME..."
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
          {query.trim().length < 2 ? (
            <div className="t-xs muted" style={{ padding: '6px 0' }}>
              &gt; TYPE 2+ CHARACTERS TO SEARCH
            </div>
          ) : results.length === 0 && !searching ? (
            <div className="t-xs muted" style={{ padding: '6px 0' }}>
              &gt; NO PLAYERS MATCHED.
            </div>
          ) : (
            <div>
              {results.map(r => (
                <SearchHit
                  key={r.label}
                  slug={r.label}
                  isFriend={friendsSet.has(r.label.toLowerCase())}
                  isSelf={r.label.toLowerCase() === userOwner}
                  country={countries[r.label]}
                  onAdd={handleAdd}
                  navigate={navigate}
                />
              ))}
            </div>
          )}
        </Panel>

        {/* Friend list */}
        <Panel
          code="FR.1"
          title={`MY FRIENDS / / ${friends.length}`}
        >
          {loading ? (
            <div className="t-md muted" style={{ textAlign: 'center', padding: 18 }}>
              &gt; LOADING<span className="blink">_</span>
            </div>
          ) : friends.length === 0 ? (
            <div className="t-md muted" style={{ textAlign: 'center', padding: 18 }}>
              &gt; NO FRIENDS ADDED YET. SEARCH ABOVE TO ADD ONE.
            </div>
          ) : (
            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))',
              gap: 14,
            }}>
              {friends.map(f => (
                <FriendTile
                  key={f}
                  owner={f}
                  country={countries[f]}
                  onUnfriend={handleUnfriend}
                  onOpen={openProfile}
                />
              ))}
            </div>
          )}
        </Panel>
      </div>
    </>
  )
}
