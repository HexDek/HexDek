import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { Panel, KV, Bar, Tag, Btn, Tape, ConfidenceDots, ManaCurveChart, ColorPie, computeColorByCmc } from '../components/chrome'
import GlossaryTerm from '../components/GlossaryTerm'
import { ConsiderCuttingRationale, ValueEngineRationale, WinConditionRationale } from '../components/RationalePanels'
import CardRolesGrid from '../components/CardRolesGrid'
import CardLink from '../components/CardLink'
import CurseDisplay from '../components/CurseDisplay'
import MatchupsPanel from '../components/MatchupsPanel'
import ManaCost from '../components/ManaCost'
import { AchievementsPanel, BadgeShowcase } from '../components/AchievementsPanel'
import { toast } from '../components/Toast'
import { api, cardArtUrl, cardImageUrl, API_BASE } from '../services/api'
import { useArtContrast } from '../hooks/useArtContrast'
import { useLiveSocket } from '../hooks/useLiveSocket'
import { useAuth } from '../context/AuthContext'
import { trackEvent } from '../hooks/useAnalytics'
import { MOCK_DECK_ANALYSIS } from '../services/mock'
import { DeckPicker } from './DeckCompare'
import DeckExportModal from '../components/DeckExportModal'
import ContextBox from '../components/ContextBox'

// Brutalist stat-summary panel: mana curve, card-type breakdown, color
// pips. Computed entirely from the in-memory deck card list — no extra
// API roundtrip, so it renders instantly even when Freya analysis hasn't
// run yet. The deeper Freya-driven curve / color charts live in the
// ANALYSIS tab; this is the always-visible top-of-page summary.
const TYPE_BUCKETS = [
  // Highest priority first — a card lands in the first bucket whose
  // keyword appears in its type_line. Land beats everything (so artifact
  // lands count as lands), Creature beats Artifact/Enchantment (so
  // enchantment-creatures and artifact-creatures count as creatures —
  // matches EDHREC convention).
  { key: 'land',         label: 'LANDS',        match: /\bland\b/i,         color: '#8a9682' },
  { key: 'planeswalker', label: 'PLANESWALKERS', match: /\bplaneswalker\b/i, color: '#cda73c' },
  { key: 'creature',     label: 'CREATURES',    match: /\bcreature\b/i,     color: '#82C472' },
  { key: 'enchantment',  label: 'ENCHANTMENTS', match: /\benchantment\b/i,  color: '#b48ad6' },
  { key: 'artifact',     label: 'ARTIFACTS',    match: /\bartifact\b/i,     color: '#9aa6b8' },
  { key: 'sorcery',      label: 'SORCERIES',    match: /\bsorcery\b/i,      color: '#cc5c4a' },
  { key: 'instant',      label: 'INSTANTS',     match: /\binstant\b/i,      color: '#6e8fa0' },
]
const PIP_COLORS = { W: '#E0EBD3', U: '#6E8FA0', B: '#3a3628', R: '#CC5C4A', G: '#82C472' }

function computeDeckStats(cards) {
  const curve = [0, 0, 0, 0, 0, 0, 0, 0] // 0..6, 7+
  const types = Object.fromEntries(TYPE_BUCKETS.map(b => [b.key, 0]))
  let typesTotal = 0
  const pips = { W: 0, U: 0, B: 0, R: 0, G: 0 }
  let pipsTotal = 0

  for (const c of cards || []) {
    const qty = c.quantity || 1
    const typeStr = (c.type_line || (Array.isArray(c.types) ? c.types.join(' ') : '') || '').toLowerCase()
    const isLand = /\bland\b/.test(typeStr) || ((c.cmc ?? -1) === 0 && !c.mana_cost && !typeStr)

    // Mana curve — non-land only.
    if (!isLand) {
      const cmc = Math.max(0, Math.min(7, c.cmc ?? 0))
      curve[cmc] += qty
    }

    // Type bucket — first match wins.
    if (typeStr) {
      const bucket = TYPE_BUCKETS.find(b => b.match.test(typeStr))
      if (bucket) {
        types[bucket.key] += qty
        typesTotal += qty
      }
    } else if (isLand) {
      types.land += qty
      typesTotal += qty
    }

    // Color pips — count {W}{U}{B}{R}{G} in mana_cost, including hybrid
    // halves like {W/U} (each half scores once for its color).
    if (c.mana_cost) {
      const matches = c.mana_cost.match(/[WUBRG]/gi) || []
      for (const m of matches) {
        const k = m.toUpperCase()
        if (pips[k] != null) { pips[k] += qty; pipsTotal += qty }
      }
    }
  }

  return { curve, types, typesTotal, pips, pipsTotal }
}

function DeckStatsSummary({ cards }) {
  const { curve, types, typesTotal, pips, pipsTotal } = computeDeckStats(cards)
  const curveMax = Math.max(1, ...curve)
  const curveLabels = ['0', '1', '2', '3', '4', '5', '6', '7+']

  // Pie geometry — one circle, segments drawn as stroked arcs via
  // stroke-dasharray. circumference 2πr; r=15.9155 keeps circumference≈100
  // so dasharray values are simply percentages.
  const segments = TYPE_BUCKETS.map(b => ({
    bucket: b,
    count: types[b.key],
    pct: typesTotal > 0 ? (types[b.key] / typesTotal) * 100 : 0,
  })).filter(s => s.count > 0)
  let pieOffset = 25 // shift starting angle to 12 o'clock
  const pieSegs = segments.map(s => {
    const seg = { ...s, offset: pieOffset }
    pieOffset += s.pct
    return seg
  })

  const pipMax = Math.max(1, ...Object.values(pips))

  return (
    <Panel code="04.S" title="DECK STATS" right={<span className="t-xs muted">{cards?.length || 0} CARDS</span>}>
      <div className="deck-stats-summary">
        {/* Mana curve histogram */}
        <div className="deck-stats-summary__col">
          <div className="t-xs muted" style={{ marginBottom: 6 }}>MANA CURVE / / NONLAND CMC</div>
          <svg viewBox="0 0 200 90" preserveAspectRatio="none" style={{ width: '100%', height: 90, display: 'block', border: '1px solid var(--rule-2)' }}>
            {curve.map((n, i) => {
              const w = 200 / curve.length
              const x = i * w
              const h = (n / curveMax) * 70
              const y = 80 - h
              return (
                <g key={i}>
                  <rect x={x + 2} y={y} width={w - 4} height={h} fill="var(--accent, var(--ink))" />
                  {n > 0 && (
                    <text x={x + w / 2} y={y - 2} textAnchor="middle" fontSize="7" fill="var(--ink-2)" fontFamily="inherit">{n}</text>
                  )}
                  <text x={x + w / 2} y={88} textAnchor="middle" fontSize="8" fill="var(--ink-3)" fontFamily="inherit" letterSpacing="0.05em">{curveLabels[i]}</text>
                </g>
              )
            })}
          </svg>
        </div>

        {/* Card type breakdown — pie + legend */}
        <div className="deck-stats-summary__col">
          <div className="t-xs muted" style={{ marginBottom: 6 }}>CARD TYPES / / {typesTotal}</div>
          <div style={{ display: 'flex', gap: 10, alignItems: 'center' }}>
            <svg viewBox="0 0 42 42" style={{ width: 90, height: 90, flexShrink: 0 }}>
              <circle cx="21" cy="21" r="15.9155" fill="var(--bg-2, #181915)" stroke="var(--rule-2)" strokeWidth="0.4" />
              {pieSegs.map(s => (
                <circle
                  key={s.bucket.key}
                  cx="21" cy="21" r="15.9155"
                  fill="transparent"
                  stroke={s.bucket.color}
                  strokeWidth="9"
                  strokeDasharray={`${s.pct.toFixed(2)} ${(100 - s.pct).toFixed(2)}`}
                  strokeDashoffset={(100 - s.offset).toFixed(2)}
                  transform="rotate(-90 21 21)"
                >
                  <title>{`${s.bucket.label}: ${s.count} (${s.pct.toFixed(1)}%)`}</title>
                </circle>
              ))}
            </svg>
            <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr auto', gap: '2px 6px', flex: 1, fontSize: 9, alignContent: 'center' }}>
              {TYPE_BUCKETS.map(b => {
                const n = types[b.key]
                if (n === 0) return null
                const pct = typesTotal > 0 ? (n / typesTotal) * 100 : 0
                return (
                  <div key={b.key} style={{ display: 'contents' }}>
                    <span style={{ width: 8, height: 8, background: b.color, border: '1px solid var(--rule-2)', alignSelf: 'center' }} />
                    <span style={{ letterSpacing: '0.05em' }}>{b.label}</span>
                    <span style={{ fontVariantNumeric: 'tabular-nums', textAlign: 'right' }}>{n} · {pct.toFixed(0)}%</span>
                  </div>
                )
              })}
            </div>
          </div>
        </div>

        {/* Color pip distribution */}
        <div className="deck-stats-summary__col">
          <div className="t-xs muted" style={{ marginBottom: 6 }}>COLOR PIPS / / {pipsTotal}</div>
          {pipsTotal === 0 ? (
            <div className="t-xs muted-2" style={{ padding: '14px 0', textAlign: 'center' }}>— COLORLESS —</div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
              {Object.entries(pips).filter(([, n]) => n > 0).map(([color, n]) => {
                const pct = (n / pipsTotal) * 100
                const barW = (n / pipMax) * 100
                return (
                  <div key={color} style={{ display: 'grid', gridTemplateColumns: '14px 1fr 56px', alignItems: 'center', gap: 6 }}>
                    <span style={{ fontSize: 11, fontWeight: 700, textAlign: 'center' }}>{color}</span>
                    <div style={{ height: 10, border: '1px solid var(--rule-2)', background: 'var(--bg-2, rgba(0,0,0,0.2))', position: 'relative' }}>
                      <div style={{ position: 'absolute', inset: 0, width: `${barW}%`, background: PIP_COLORS[color], opacity: 0.85 }} />
                    </div>
                    <span className="t-xs" style={{ textAlign: 'right', fontVariantNumeric: 'tabular-nums' }}>{n} · {pct.toFixed(0)}%</span>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </Panel>
  )
}

const CardThumb = ({ name, cmc, score, compact }) => {
  const imgUrl = cardArtUrl(name)
  // Whole tile is a CardLink. underline=false because the click
  // affordance is the art tile itself; a dotted underline on a 5/7
  // image would be visual noise.
  if (compact) {
    return (
      <CardLink name={name} underline={false} style={{ display: 'block' }}>
        <div className="panel" style={{ padding: 0 }}>
          <div style={{ aspectRatio: '5/4', position: 'relative', overflow: 'hidden' }}>
            <img src={imgUrl} alt={name} style={{ width: '100%', height: '100%', objectFit: 'cover', filter: 'saturate(0.6) contrast(1.1)' }} onError={e => { e.target.style.display = 'none'; e.target.parentElement.classList.add('hatch') }} />
          </div>
          <div style={{ padding: '3px 5px' }}>
            <div style={{ fontSize: 7, fontWeight: 700, letterSpacing: '0.04em', textTransform: 'uppercase', lineHeight: 1.1, minHeight: 14, overflow: 'hidden', textOverflow: 'ellipsis' }}>{name}</div>
          </div>
        </div>
      </CardLink>
    )
  }
  return (
    <CardLink name={name} underline={false} style={{ display: 'block' }}>
      <div className="panel" style={{ padding: 0 }}>
        <div style={{ aspectRatio: '5/7', borderBottom: '1px solid var(--rule-2)', position: 'relative', overflow: 'hidden' }}>
          <img src={imgUrl} alt={name} style={{ width: '100%', height: '100%', objectFit: 'cover', filter: 'saturate(0.6) contrast(1.1)' }} onError={e => { e.target.style.display = 'none'; e.target.parentElement.classList.add('hatch') }} />
          <span style={{ position: 'absolute', top: 4, left: 5, background: 'rgba(12,13,10,0.6)', padding: '0 3px' }} className="t-xs muted-2">{cmc || ''}</span>
          {score && <span style={{ position: 'absolute', top: 4, right: 5, fontSize: 9, color: 'var(--ok)' }}>■{score}</span>}
        </div>
        <div style={{ padding: '5px 7px' }}>
          <div style={{ fontSize: 9, fontWeight: 700, letterSpacing: '0.04em', textTransform: 'uppercase', lineHeight: 1.2, minHeight: 24 }}>{name}</div>
        </div>
      </div>
    </CardLink>
  )
}

export default function DeckArchive() {
  const { owner, id } = useParams()
  const navigate = useNavigate()
  const [deck, setDeck] = useState(null)
  const [analysis, setAnalysis] = useState(null)
  const [loading, setLoading] = useState(true)
  const [analyzing, setAnalyzing] = useState(false)
  const [editing, setEditing] = useState(false)
  const [editText, setEditText] = useState('')
  const [saving, setSaving] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [comparePickerOpen, setComparePickerOpen] = useState(false)
  const [exportOpen, setExportOpen] = useState(false)
  const [versions, setVersions] = useState([])
  const [gauntlet, setGauntlet] = useState(null)
  const [curse, setCurse] = useState(null)
  const [achievements, setAchievements] = useState(null)
  const [editingName, setEditingName] = useState(false)
  const [nameDraft, setNameDraft] = useState('')
  const [savingName, setSavingName] = useState(false)
  const [winLinesExpanded, setWinLinesExpanded] = useState(false)
  const [cloning, setCloning] = useState(false)
  const [confirmClone, setConfirmClone] = useState(false)
  const [spawningRoom, setSpawningRoom] = useState(false)
  const [isFriend, setIsFriend] = useState(false)
  const [friendBusy, setFriendBusy] = useState(false)
  const [ownerFriendCount, setOwnerFriendCount] = useState(null)
  const [similarDecks, setSimilarDecks] = useState(null) // null=loading, []=resolved
  const [activeTab, setActiveTab] = useState('analysis')
  const { elo } = useLiveSocket()
  const { user } = useAuth()

  const startNameEdit = () => {
    setNameDraft(deck?.custom_name || deck?.commander || '')
    setEditingName(true)
  }

  const cancelNameEdit = () => {
    setEditingName(false)
    setNameDraft('')
  }

  const commitNameEdit = async () => {
    if (!owner || !id || savingName) return
    const trimmed = nameDraft.trim()
    const current = deck?.custom_name || ''
    if (trimmed === current) {
      cancelNameEdit()
      return
    }
    setSavingName(true)
    try {
      const updated = await api.patchDeck(`${owner}/${id}`, { name: trimmed })
      setDeck(d => ({ ...(d || {}), custom_name: updated.custom_name || '' }))
      trackEvent('rename_deck', { deck: `${owner}/${id}`, len: trimmed.length })
      setEditingName(false)
      toast.success('DECK RENAMED')
    } catch (err) {
      toast.error('RENAME FAILED')
    } finally {
      setSavingName(false)
    }
  }

  const userOwnerSlug = user
    ? (localStorage.getItem('hexdek_owner') || user.displayName?.toLowerCase() || user.email?.split('@')[0]?.split('.')[0] || '')
    : ''
  const isOwner = !!owner && !!userOwnerSlug && userOwnerSlug === owner.toLowerCase()
  const canFriend = !!user && !!userOwnerSlug && !!owner && !isOwner

  useEffect(() => {
    if (!canFriend) { setIsFriend(false); return }
    let cancelled = false
    api.listFriends(userOwnerSlug)
      .then(r => { if (!cancelled) setIsFriend((r.friends || []).includes(owner.toLowerCase())) })
      .catch(() => {})
    return () => { cancelled = true }
  }, [canFriend, owner, userOwnerSlug])

  // Pull the deck owner's friend count for the DECK SPECS panel. Refetches
  // when the owner changes or when this visitor's add/remove fires the
  // 'hexdek-friends-changed' event (mutual-add updates the owner's count too).
  useEffect(() => {
    if (!owner) { setOwnerFriendCount(null); return }
    let cancelled = false
    const load = () => {
      api.listFriends(owner)
        .then(r => { if (!cancelled) setOwnerFriendCount((r.friends || []).length) })
        .catch(() => { if (!cancelled) setOwnerFriendCount(null) })
    }
    load()
    const onChanged = () => load()
    window.addEventListener('hexdek-friends-changed', onChanged)
    return () => {
      cancelled = true
      window.removeEventListener('hexdek-friends-changed', onChanged)
    }
  }, [owner])

  const toggleFriend = async () => {
    if (!canFriend || friendBusy) return
    setFriendBusy(true)
    const target = owner.toLowerCase()
    const wasFriend = isFriend
    setIsFriend(!wasFriend) // optimistic
    try {
      if (wasFriend) await api.removeFriend(target, userOwnerSlug)
      else           await api.addFriend(target, userOwnerSlug)
      trackEvent(wasFriend ? 'remove_friend' : 'add_friend', { target })
      window.dispatchEvent(new CustomEvent('hexdek-friends-changed'))
      toast.success(wasFriend ? `UNFRIENDED ${target.toUpperCase()}` : `FRIEND ADDED · ${target.toUpperCase()}`)
    } catch {
      setIsFriend(wasFriend) // rollback
      toast.error(wasFriend ? 'UNFRIEND FAILED' : 'ADD FRIEND FAILED')
    } finally {
      setFriendBusy(false)
    }
  }

  const handleShare = async () => {
    if (!owner || !id) return
    const url = `${window.location.origin}/decks/${owner}/${id}`
    let copied = false
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(url)
        copied = true
      } else {
        const ta = document.createElement('textarea')
        ta.value = url
        ta.style.position = 'fixed'
        ta.style.opacity = '0'
        document.body.appendChild(ta)
        ta.select()
        copied = document.execCommand('copy')
        document.body.removeChild(ta)
      }
    } catch {}
    trackEvent('share_deck', { deck: `${owner}/${id}`, copied })
    if (copied) toast.success('SHARE LINK COPIED')
    else toast.error('COPY FAILED — ' + url, 5000)
  }

  const eloByDeckId = {}
  for (const e of elo) {
    if (e.deck_id) eloByDeckId[e.deck_id] = e
  }
  const deckKey = owner && id ? `${owner}/${id}` : null
  const deckElo = eloByDeckId[deckKey] || eloByDeckId[id] || null

  const fetchAnalysis = (ownerId, deckId) => {
    api.getDeckAnalysis(`${ownerId}/${deckId}`).then(data => {
      if (data.status === 'analyzing') {
        setAnalyzing(true)
      } else {
        setAnalysis(data)
        setAnalyzing(false)
      }
    }).catch(() => setAnalyzing(false))
  }

  useEffect(() => {
    if (!owner || !id) {
      setAnalysis(MOCK_DECK_ANALYSIS.tinybones)
      setLoading(false)
      return
    }
    Promise.allSettled([
      api.getDeck(`${owner}/${id}`),
      api.getDeckAnalysis(`${owner}/${id}`),
      api.getGauntlet(`${owner}/${id}`),
      api.getDeckCurse(`${owner}/${id}`),
      api.getAchievements(owner),
    ]).then(([deckRes, analysisRes, gauntletRes, curseRes, achievementsRes]) => {
      if (deckRes.status === 'fulfilled') setDeck(deckRes.value)
      if (analysisRes.status === 'fulfilled') {
        const data = analysisRes.value
        if (data.status === 'analyzing') {
          setAnalyzing(true)
        } else {
          setAnalysis(data)
        }
      }
      if (curseRes.status === 'fulfilled' && curseRes.value && curseRes.value.population) {
        setCurse(curseRes.value)
      }
      if (achievementsRes.status === 'fulfilled' && achievementsRes.value) {
        setAchievements(achievementsRes.value)
      }
      if (gauntletRes.status === 'fulfilled' && gauntletRes.value.status !== 'none') {
        setGauntlet(gauntletRes.value)
        if (gauntletRes.value.status === 'running') {
          const poll = () => {
            api.getGauntlet(`${owner}/${id}`).then(r => {
              setGauntlet(r)
              if (r.status === 'running') setTimeout(poll, 3000)
            })
          }
          setTimeout(poll, 3000)
        }
      }
      setLoading(false)
    })
  }, [owner, id])

  // Similar decks — independent fetch so the rest of the page renders
  // immediately. Server scans DecksDir and returns a ranked top-5.
  useEffect(() => {
    if (!owner || !id) { setSimilarDecks([]); return }
    let cancelled = false
    api.getSimilarDecks(`${owner}/${id}`, 5)
      .then(rows => { if (!cancelled) setSimilarDecks(Array.isArray(rows) ? rows : []) })
      .catch(() => { if (!cancelled) setSimilarDecks([]) })
    return () => { cancelled = true }
  }, [owner, id])

  // SSE listener — auto-refresh analysis when Freya completes.
  useEffect(() => {
    if (!owner || !id) return
    const es = new EventSource(`${API_BASE}/api/decks/${owner}/${id}/events`)
    es.addEventListener('freya_started', () => setAnalyzing(true))
    es.addEventListener('freya_complete', () => {
      api.getDeckAnalysis(`${owner}/${id}`).then(data => {
        setAnalysis(data)
        setAnalyzing(false)
      }).catch(() => setAnalyzing(false))
    })
    es.onerror = () => {}
    return () => es.close()
  }, [owner, id])

  const deckName = deck?.custom_name || deck?.commander || id?.replace(/_/g, ' ').toUpperCase() || 'DECK'

  useEffect(() => {
    if (!deckName) return
    const ownerLabel = owner ? ` · ${owner.toUpperCase()}` : ''
    document.title = `${deckName}${ownerLabel} — HEXDEK`
  }, [deckName, owner])

  const cardCount = deck?.card_count || deck?.cards?.length || 99
  const userBracket = deck?.bracket || '?'
  const wbs = analysis?.bracket || userBracket
  const wbsLabel = analysis?.bracket_label || ''
  const pls = analysis?.plays_like || null
  const plsLabel = analysis?.plays_like_label || ''
  const gameChangers = analysis?.game_changer_count ?? null
  const archetype = analysis?.archetype?.toUpperCase() || 'UNKNOWN'
  const summary = analysis?.gameplan_summary || ''
  const winLines = analysis?.win_lines || []
  const valueKeys = analysis?.value_engine_keys || []
  const evalWeights = analysis?.eval_weights || {}
  const cards = deck?.cards || []
  const manaBaseGrade = analysis?.mana_base_grade || null
  const keepableHandPct = analysis?.keepable_hand_pct ?? null
  const powerPercentile = analysis?.power_percentile ?? null
  const commanderSynergy = analysis?.commander_synergy ?? null
  const commanderThemes = analysis?.commander_themes || []
  const starCards = analysis?.star_cards || []
  // Prefer the structured rationale list when Freya has produced it; fall
  // back to the flat name list for older strategy.json files on disk.
  const cuttableCards = analysis?.cuttable_card_rationale || analysis?.cuttable_cards || []
  const valueChains = analysis?.value_chains || []
  const vulnerableTo = analysis?.vulnerable_to || []
  const finisherCards = analysis?.finisher_cards || []
  const comboNotes = analysis?.combo_notes || []
  const curveWarnings = analysis?.curve_warnings || []
  const colorMismatch = analysis?.color_mismatch || []
  const legality = analysis?.legality || null
  const gameChangerCards = analysis?.game_changer_cards || []
  const interactionAvgCmc = analysis?.interaction_avg_cmc ?? null
  const cheapInteraction = analysis?.cheap_interaction ?? null
  const emergentSynergies = analysis?.emergent_synergies || []
  const metaMatchups = analysis?.meta_matchups || []
  const cardRoles = analysis?.card_roles || null

  // Derive commander color identity for page theming. Prefer Freya's analysis
  // (authoritative), then commander mana cost, then any pip in the decklist.
  const colorIdentity = (() => {
    if (Array.isArray(analysis?.color_identity) && analysis.color_identity.length) {
      return [...analysis.color_identity].map(c => c.toUpperCase()).filter(c => 'WUBRG'.includes(c))
        .sort((a, b) => 'WUBRG'.indexOf(a) - 'WUBRG'.indexOf(b))
    }
    const ci = new Set()
    const scan = mc => {
      if (!mc) return
      const pips = mc.match(/\{([WUBRG])\}/gi) || []
      for (const p of pips) ci.add(p.replace(/[{}]/g, '').toUpperCase())
    }
    const cmdrName = deck?.commander_card
    if (cmdrName) {
      const cmdr = cards.find(c => c.name === cmdrName)
      if (cmdr) scan(cmdr.mana_cost)
    }
    if (ci.size === 0) for (const c of cards) scan(c.mana_cost)
    return Array.from(ci).sort((a, b) => 'WUBRG'.indexOf(a) - 'WUBRG'.indexOf(b))
  })()

  const pageTheme = (() => {
    // Per-color palette: rgba base for the wash, hex accent for highlights.
    const COLORS = {
      W: { base: '226, 218, 188', accent: '#d8c878' },
      U: { base: '34, 70, 110',   accent: '#5a8fbf' },
      B: { base: '36, 26, 42',    accent: '#9c6ab0' },
      R: { base: '78, 28, 22',    accent: '#cc5c4a' },
      G: { base: '36, 70, 36',    accent: '#7ac28a' },
    }
    const ids = colorIdentity.length ? colorIdentity : []
    if (ids.length === 0) {
      return { wash: 'linear-gradient(135deg, rgba(28,29,22,0.9), rgba(20,21,15,0.9))', accent: '#8a9682', label: 'COLORLESS' }
    }
    // Build a 135deg gradient across the colors. Single colors get a soft
    // top-left → bottom-right fade between two intensities of the same hue.
    let stops
    if (ids.length === 1) {
      const c = COLORS[ids[0]]
      stops = `rgba(${c.base}, 0.85) 0%, rgba(${c.base}, 0.35) 100%`
    } else {
      stops = ids.map((c, i) => {
        const pct = (i / (ids.length - 1)) * 100
        return `rgba(${COLORS[c].base}, 0.7) ${pct.toFixed(0)}%`
      }).join(', ')
    }
    // Pick accent by visual distinctiveness priority: R > G > U > B > W.
    const accentPriority = ['R', 'G', 'U', 'B', 'W']
    const accentColor = ids.find(c => accentPriority.includes(c))
      ? COLORS[accentPriority.find(c => ids.includes(c))].accent
      : '#8a9682'
    const COMBO_NAMES = {
      W: 'MONO WHITE', U: 'MONO BLUE', B: 'MONO BLACK', R: 'MONO RED', G: 'MONO GREEN',
      WU: 'AZORIUS', UB: 'DIMIR', BR: 'RAKDOS', RG: 'GRUUL', GW: 'SELESNYA',
      WB: 'ORZHOV', UR: 'IZZET', BG: 'GOLGARI', RW: 'BOROS', UG: 'SIMIC',
      WUB: 'ESPER', UBR: 'GRIXIS', BRG: 'JUND', RGW: 'NAYA', GWU: 'BANT',
      WBG: 'ABZAN', URW: 'JESKAI', BGU: 'SULTAI', RWB: 'MARDU', GUR: 'TEMUR',
      WUBR: 'YORE-TILLER', WUBG: 'WITCH-MAW', WURG: 'INK-TREADER', WBRG: 'DUNE-BROOD', UBRG: 'GLINT-EYE',
      WUBRG: 'FIVE-COLOR',
    }
    const label = COMBO_NAMES[ids.join('')] || ids.join('')
    return { wash: `linear-gradient(135deg, ${stops})`, accent: accentColor, label }
  })()

  const clientCurve = (() => {
    if (!cards.length) return null
    const dist = Array(8).fill(0)
    let totalCmc = 0, nonlandCount = 0, landCount = 0
    const demand = {}
    for (const c of cards) {
      const qty = c.quantity || 1
      const hasType = c.type_line || c.types
      const typeStr = (c.type_line || (c.types && c.types.join(' ')) || '').toLowerCase()
      const isLand = hasType ? /\bland\b/.test(typeStr) : ((c.cmc ?? -1) === 0 && !c.mana_cost)
      if (isLand) { landCount += qty; continue }
      const cmc = Math.min(c.cmc ?? 0, 7)
      dist[cmc] += qty
      totalCmc += (c.cmc ?? 0) * qty
      nonlandCount += qty
      if (c.mana_cost) {
        const pips = c.mana_cost.match(/\{([WUBRG])}/gi) || []
        for (const p of pips) {
          const color = p.replace(/[{}]/g, '')
          demand[color] = (demand[color] || 0) + qty
        }
      }
    }
    const avgCmc = nonlandCount > 0 ? totalCmc / nonlandCount : 0
    const peak = dist.indexOf(Math.max(...dist))
    const shape = peak <= 2 ? 'LOW CURVE' : peak <= 4 ? 'MID CURVE' : 'HIGH CURVE'
    return { distribution: dist, avg_cmc: avgCmc, curve_shape: shape, land_count: landCount, nonland_count: nonlandCount, demand }
  })()

  const curveData = analysis?.mana_curve || clientCurve
  const colorData = analysis?.color_balance?.demand || clientCurve?.demand

  const manaProduction = deck?.mana_production || (() => {
    if (!cards.length) return null
    const production = {}
    const basicMap = { plains: 'W', island: 'U', swamp: 'B', mountain: 'R', forest: 'G' }
    for (const c of cards) {
      const qty = c.quantity || 1
      const typeStr = (c.type_line || '').toLowerCase()
      if (!/\bland\b/.test(typeStr)) continue
      for (const [basic, color] of Object.entries(basicMap)) {
        if (typeStr.includes(basic)) {
          production[color] = (production[color] || 0) + qty
        }
      }
    }
    return production
  })()

  const demandColors = colorData ? Object.keys(colorData).filter(k => colorData[k] > 0) : []
  const isMultiColor = demandColors.length >= 2

  const cmdrCardName = deck?.commander_card || cards.find(c => c.name?.startsWith('COMMANDER:'))?.name?.replace('COMMANDER:', '').trim()
  const cmdrImageUrl = cmdrCardName
    ? cardArtUrl(cmdrCardName)
    : null
  const cmdrFullUrl = cmdrCardName ? cardImageUrl(cmdrCardName) : null
  const cmdrContrast = useArtContrast(cmdrImageUrl)

  if (loading) {
    return (
      <>
        <Tape left="DECK ARCHIVE / / LOADING" mid="" right="" />
        <div style={{ padding: 36, textAlign: 'center' }}>
          <div className="t-md muted">&gt; LOADING DECK DATA<span className="blink">_</span></div>
        </div>
      </>
    )
  }

  return (
    <div
      className="deck-archive-page"
      style={{
        '--page-wash': pageTheme.wash,
        '--accent': pageTheme.accent,
      }}
    >
      {/* Blown-up gaussian-blurred commander art behind everything —
          shared mechanism with CardPage via the .art-ambience class. */}
      {cmdrImageUrl && (
        <img
          className="art-ambience"
          src={cmdrImageUrl}
          alt=""
          aria-hidden="true"
        />
      )}

      <Tape
        left={`DECK ARCHIVE / / ${owner?.toUpperCase()} / / ${deckName}`}
        mid={pls ? `Plays Like B${pls} (Bracket B${wbs}) · ${pageTheme.label}` : `Bracket B${wbs} · ${pageTheme.label}`}
        right="EXPORT ↗ ANALYZE ↗"
      />

      <div
        className={`deck-hero ${cmdrImageUrl ? '' : 'hatch'}`}
        data-art-contrast={cmdrContrast || undefined}
        style={cmdrImageUrl
          ? { backgroundImage: `url(${cmdrImageUrl})`, ...(cmdrContrast ? { '--art-contrast': cmdrContrast } : null) }
          : undefined}
      >
        <div className="deck-hero__scrim" />
        <div className="deck-hero__corner deck-hero__corner--tl">04.HERO / / {pageTheme.label}</div>
        <div className="deck-hero__corner deck-hero__corner--tr">{owner?.toUpperCase()} / / {id}</div>
        <div className="deck-hero__actions">
          {canFriend && (
            <button
              type="button"
              className={`deck-hero__friend ${isFriend ? 'is-on' : ''}`}
              onClick={toggleFriend}
              disabled={friendBusy}
              title={isFriend ? `Unfriend ${owner.toUpperCase()}` : `Add ${owner.toUpperCase()} as a friend`}
            >
              <span>{isFriend ? '✓ FRIEND' : '+ ADD FRIEND'}</span>
            </button>
          )}
          {owner && id && (
            <button type="button" className="deck-hero__share" onClick={handleShare} title="Copy shareable link">
              <span>SHARE</span>
              <span className="arr">↗</span>
            </button>
          )}
          {owner && id && (
            <button type="button" className="deck-hero__share" onClick={() => setComparePickerOpen(true)} title="Compare against another deck">
              <span>COMPARE</span>
              <span className="arr">⇄</span>
            </button>
          )}
        </div>
        <div className="deck-hero__body">
          {cmdrFullUrl && (
            <div className="deck-hero__card">
              <img
                src={cmdrFullUrl}
                alt={cmdrCardName}
                className="deck-hero__card-img"
                onError={(e) => { e.target.style.display = 'none' }}
              />
            </div>
          )}
          <div style={{ flex: 1, minWidth: 0 }}>
          <div className="deck-hero__meta">
            <Tag solid>B{wbs}{wbsLabel ? ' · ' + wbsLabel : ''}</Tag>
            {pls && pls !== wbs && <Tag solid kind="warn">PLAYS LIKE B{pls}</Tag>}
            <Tag>{archetype}</Tag>
            {colorIdentity.length > 0 && <Tag>{colorIdentity.join('')}</Tag>}
          </div>
          <div className="deck-hero__title-row">
            {editingName ? (
              <input
                autoFocus
                className="deck-hero__title-input"
                value={nameDraft}
                maxLength={120}
                disabled={savingName}
                onChange={e => setNameDraft(e.target.value)}
                onBlur={commitNameEdit}
                onKeyDown={e => {
                  if (e.key === 'Enter') { e.preventDefault(); commitNameEdit() }
                  else if (e.key === 'Escape') { e.preventDefault(); cancelNameEdit() }
                }}
              />
            ) : (
              <>
                <h1 className="deck-hero__title">{deckName}</h1>
                {isOwner && (
                  <button
                    type="button"
                    className="deck-hero__rename"
                    onClick={startNameEdit}
                    title="Rename deck"
                    aria-label="Rename deck"
                  >✎</button>
                )}
              </>
            )}
          </div>
          {cmdrCardName && cmdrCardName.toUpperCase() !== deckName && (
            <div className="deck-hero__sub">{cmdrCardName}</div>
          )}
          {/* gameplan_summary hidden — Freya win-line detection needs accuracy pass */}
          </div>
        </div>
      </div>

      {/* Hero quick-actions context — explains the floating SHARE / COMPARE
          / FRIEND buttons in the hero. Dismissible so it disappears once
          the user has read it. */}
      {owner && id && (
        <div className="deck-hero__actions-context">
          <ContextBox id="deck.hero.actions">
            <strong>SHARE</strong> copies a public link to this deck page to your clipboard.
            {' '}<strong>COMPARE</strong> opens a side-by-side diff with another deck (overlap, color identity, archetype).
            {canFriend && <> <strong>+ ADD FRIEND</strong> follows {owner?.toUpperCase()} so their decks surface in your feed.</>}
          </ContextBox>
        </div>
      )}

      {/* Deck stats summary — always visible between hero and main columns. */}
      <div className="deck-stats-summary-row">
        <DeckStatsSummary cards={cards} />
      </div>

      <div className="archive-layout">
        <div className="archive-sidebar">
          <Panel code="04.A" title="DECK SPECS" solid>
            <KV rows={[
              ['OWNER', <Link to={`/profile/${owner}`} style={{ color: 'var(--ink)', textDecoration: 'none', borderBottom: '1px dotted var(--ink-3)' }}>{owner?.toUpperCase()}</Link>],
              ...(ownerFriendCount != null ? [['FRIENDS', String(ownerFriendCount)]] : []),
              ['CARDS', `${cardCount}`],
              ['BRACKET', `B${wbs}${wbsLabel ? ' ' + wbsLabel : ''}`, 'bracket'],
              ['PLAYS LIKE', pls ? `B${pls}${plsLabel ? ' ' + plsLabel : ''}${pls != wbs ? ' ⬆' : ''}` : '—', 'plays_like'],
              ['GAME CHANGERS', gameChangers != null ? `${gameChangers}` : '—', 'game_changers'],
              ['ARCHETYPE', archetype, 'archetype'],
              ...(legality ? [['LEGALITY', <span style={{ color: legality.valid ? 'var(--ok)' : 'var(--danger)', fontWeight: 700 }}>{legality.valid ? 'LEGAL' : 'ILLEGAL'}</span>, 'legality']] : []),
              ...(manaBaseGrade ? [['MANA BASE', manaBaseGrade, 'mana_base_grade']] : []),
              ...(powerPercentile != null ? [['POWER', `TOP ${powerPercentile}%`, 'power_percentile']] : []),
              ...(commanderSynergy != null ? [['CMDR SYNERGY', `${Math.round(commanderSynergy * 100)}%`, 'cmdr_synergy']] : []),
              ...(keepableHandPct != null ? [['KEEPABLE HANDS', `${Math.round(keepableHandPct)}%`, 'keepable_hands']] : []),
              ...(interactionAvgCmc != null ? [['INTERACTION CMC', `AVG ${Math.round(interactionAvgCmc * 10) / 10}`, 'interaction_avg_cmc']] : []),
              ...(cheapInteraction != null ? [['CHEAP REMOVAL', `${cheapInteraction} AT ≤2 CMC`, 'cheap_interaction']] : []),
            ]} />
            {commanderThemes.length > 0 && (
              <div style={{ marginTop: 8, display: 'flex', flexWrap: 'wrap', gap: 4 }}>
                {commanderThemes.map((t, i) => <Tag key={i}>{t.toUpperCase()}</Tag>)}
              </div>
            )}
            {deckElo && (
              <>
                <div className="hr" style={{ margin: '10px 0' }} />
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                  <GlossaryTerm term="confidence" compact>
                    <span className="t-xs muted">CONFIDENCE</span>
                  </GlossaryTerm>
                  <ConfidenceDots games={deckElo.games} showLabel size="lg" />
                </div>
                <KV rows={[
                  ['HexELO', <span className="punch">{Math.round(deckElo.hex_rating || 0)}</span>, 'hexelo'],
                  ['TS μ', <span className="t-xs muted-2">{Math.round(deckElo.mu || 0)}</span>, 'ts_mu'],
                  ['RECORD', <span><span style={{ color: 'var(--ok)' }}>{deckElo.wins}W</span> — <span style={{ color: 'var(--danger)' }}>{deckElo.losses}L</span></span>, 'record'],
                  ['WIN RATE', `${deckElo.win_rate}%`, 'win_rate'],
                  ['GAMES', `${deckElo.games?.toLocaleString()}`, 'games'],
                  ['DELTA', <span style={{ color: deckElo.delta >= 0 ? 'var(--ok)' : 'var(--danger)' }}>{deckElo.delta >= 0 ? '+' : ''}{Math.round(deckElo.delta)}</span>, 'delta'],
                ]} />
              </>
            )}
            <div className="hr" style={{ margin: '10px 0' }} />
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              {owner && id && (
                <>
                  <ContextBox id="deck.edit" compact>Opens an editor with the full deck list. Saving re-runs Freya analysis on the new list.</ContextBox>
                  <Btn arrow="↗" onClick={() => {
                    if (editing) return
                    const lines = cards.map(c => {
                      const cmdr = deck?.commander_card
                      if (cmdr && c.name === cmdr) return `COMMANDER: ${c.name}`
                      return c.quantity > 1 ? `${c.quantity} ${c.name}` : `1 ${c.name}`
                    })
                    setEditText(lines.join('\n'))
                    setEditing(true)
                    api.getDeckVersions(`${owner}/${id}`).then(setVersions).catch(() => {})
                  }}>EDIT DECK</Btn>
                </>
              )}
              <ContextBox id="deck.export" compact>Downloads the decklist in your chosen format (Moxfield, Arena, plain text).</ContextBox>
              <Btn ghost arrow="↗" onClick={() => {
                if (!cards.length) return
                setExportOpen(true)
              }}>EXPORT</Btn>
              {analyzing && <Tag solid kind="info">ANALYZING...</Tag>}
              {owner && id && (
                <>
                  <ContextBox id="deck.forge" compact>Opens this deck in the Forge — interactive playtester for testing draws, mulligans, and lines.</ContextBox>
                  <Btn ghost arrow="↗" onClick={() => navigate(`/forge?deck=${owner}/${id}`)}>OPEN IN FORGE</Btn>
                </>
              )}
              {owner && id && !isOwner && user && (
                <>
                  <ContextBox id="deck.clone" compact>Copies this deck into your account so you can edit and tune your own version. The clone re-runs Freya analysis on import.</ContextBox>
                  {!confirmClone ? (
                    <Btn solid arrow="⎘" onClick={() => setConfirmClone(true)}>
                      CLONE DECK
                    </Btn>
                  ) : (
                    <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                      <Btn
                        solid
                        arrow="⎘"
                        disabled={cloning}
                        onClick={() => {
                          if (cloning) return
                          setCloning(true)
                          trackEvent('clone_deck', { deck: `${owner}/${id}` })
                          api.cloneDeck(`${owner}/${id}`).then(res => {
                            toast.success('DECK CLONED — RUNNING FREYA')
                            navigate(`/decks/${res.owner}/${res.id}`)
                          }).catch(err => {
                            if (err?.status === 401) toast.error('SIGN IN TO CLONE')
                            else if (err?.status === 429) toast.error('CLONE LIMIT REACHED — TRY AGAIN IN AN HOUR')
                            else if (err?.status === 400) toast.error(err.message || 'CLONE REJECTED')
                            else if (err?.status === 404) toast.error('SOURCE DECK NOT FOUND')
                            else toast.error('CLONE FAILED')
                            setCloning(false)
                            setConfirmClone(false)
                          })
                        }}
                      >
                        {cloning ? 'CLONING (FREYA RUNNING)...' : 'CONFIRM CLONE'}
                      </Btn>
                      <Btn ghost arrow="✕" onClick={() => setConfirmClone(false)} disabled={cloning}>CANCEL</Btn>
                    </div>
                  )}
                </>
              )}
              {owner && id && !isOwner && !user && (
                <>
                  <ContextBox id="deck.clone" compact>Sign in to clone this deck into your own collection — Freya will re-analyze the copy on import.</ContextBox>
                  <Btn ghost arrow="↗" onClick={() => navigate('/login')}>SIGN IN TO CLONE</Btn>
                </>
              )}
              {owner && id && (
                <>
                  <div className="hr" style={{ margin: '4px 0' }} />
                  {!confirmDelete ? (
                    <>
                      <ContextBox id="deck.delete" compact tone="danger">Permanently removes this deck and its analysis. This cannot be undone.</ContextBox>
                      <Btn ghost onClick={() => setConfirmDelete(true)} style={{ color: 'var(--danger)', borderColor: 'var(--danger)' }}>DELETE DECK</Btn>
                    </>
                  ) : (
                    <>
                      <ContextBox compact tone="danger">Final confirmation — CONFIRM deletes the deck for good. CANCEL backs out.</ContextBox>
                      <div style={{ display: 'flex', gap: 6 }}>
                        <Btn solid onClick={() => {
                          api.deleteDeck(`${owner}/${id}`).then(() => navigate('/decks')).catch(() => setConfirmDelete(false))
                        }} style={{ flex: 1, background: 'var(--danger)', borderColor: 'var(--danger)' }}>CONFIRM</Btn>
                        <Btn ghost onClick={() => setConfirmDelete(false)} style={{ flex: 1 }}>CANCEL</Btn>
                      </div>
                    </>
                  )}
                </>
              )}
            </div>
            {owner && (
              <>
                <div className="hr" style={{ margin: '10px 0' }} />
                <BadgeShowcase owner={owner} />
              </>
            )}
          </Panel>

          {/* MATCHUPS — head-to-head record per opposing commander
              from showmatch_game history. Best/worst leaderboards
              gate on a min-games threshold so 1-0 small samples don't
              dominate the rankings. */}
          <MatchupsPanel owner={owner} id={id} />

          {/* SIMILAR DECKS — server-ranked by shared-card overlap with
              bonuses for matching commander / archetype / bracket. The
              endpoint already drops noise (≤10 shared cards and no
              bonus); an empty response means we genuinely have nothing
              to recommend yet. */}
          <Panel
            code="04.SIM"
            title={`SIMILAR DECKS / / ${similarDecks == null ? '…' : similarDecks.length}`}
            right={similarDecks && similarDecks.length > 0 ? <Tag solid>{similarDecks.length}</Tag> : null}
          >
            {similarDecks == null ? (
              <div className="t-xs muted" style={{ padding: '10px 0', textAlign: 'center' }}>
                &gt; SCANNING DECK INDEX<span className="blink">_</span>
              </div>
            ) : similarDecks.length === 0 ? (
              <div className="t-xs muted" style={{ padding: '10px 0', textAlign: 'center', lineHeight: 1.6 }}>
                &gt; NO SIMILAR DECKS FOUND.
              </div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                {similarDecks.map((d, i) => {
                  const cmdrArt = d.commander_card ? cardArtUrl(d.commander_card) : null
                  const showName = (d.commander || d.name || d.id || '').toUpperCase()
                  const tags = []
                  if (d.same_commander) tags.push('CMDR')
                  if (d.same_archetype) tags.push('ARCHE')
                  if (d.same_bracket)   tags.push(`B${d.bracket}`)
                  return (
                    <Link
                      key={`${d.owner}/${d.id}`}
                      to={`/decks/${d.owner}/${d.id}`}
                      style={{
                        display: 'grid',
                        gridTemplateColumns: '52px 1fr',
                        gap: 8,
                        padding: 4,
                        border: '1px solid var(--rule-2)',
                        textDecoration: 'none',
                        color: 'var(--ink)',
                        background: i === 0 ? 'color-mix(in srgb, var(--accent) 10%, transparent)' : 'transparent',
                      }}
                      title={`${showName} · ${d.shared_cards} shared`}
                    >
                      <div
                        className={cmdrArt ? '' : 'hatch'}
                        style={{
                          width: 52, height: 40, overflow: 'hidden',
                          backgroundImage: cmdrArt ? `url(${cmdrArt})` : undefined,
                          backgroundSize: 'cover', backgroundPosition: 'center 30%',
                          filter: 'saturate(0.6) contrast(1.05)',
                        }}
                      />
                      <div style={{ minWidth: 0 }}>
                        <div className="t-xs" style={{
                          fontWeight: 700, letterSpacing: '0.04em',
                          overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
                        }}>
                          {showName}
                        </div>
                        <div className="t-xs muted-2" style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                          {(d.owner || '').toUpperCase()}
                        </div>
                        <div className="t-xs" style={{
                          marginTop: 2, display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap',
                        }}>
                          <span style={{ color: 'var(--ok)', fontWeight: 700, fontVariantNumeric: 'tabular-nums' }}>
                            {d.shared_cards} SHARED
                          </span>
                          {tags.map(t => (
                            <span key={t} style={{
                              fontSize: 8, letterSpacing: '0.08em', padding: '0 4px',
                              border: '1px solid color-mix(in srgb, var(--accent) 50%, var(--rule-2))',
                              color: 'var(--ink-2)',
                            }}>{t}</span>
                          ))}
                        </div>
                      </div>
                    </Link>
                  )
                })}
              </div>
            )}
          </Panel>
        </div>

        <div className="archive-main">
          {/* Tab bar */}
          <div className="deck-tabs">
            <button type="button" className={`deck-tab ${activeTab === 'analysis' ? 'active' : ''}`} onClick={() => setActiveTab('analysis')}>ANALYSIS</button>
            <button type="button" className={`deck-tab ${activeTab === 'decklist' ? 'active' : ''}`} onClick={() => setActiveTab('decklist')}>DECK LIST</button>
            <button type="button" className={`deck-tab ${activeTab === 'achievements' ? 'active' : ''}`} onClick={() => setActiveTab('achievements')}>ACHIEVEMENTS</button>
          </div>

          {/* Edit mode — always visible regardless of tab */}
          {editing && (
            <Panel code="04.X" title="EDIT DECK LIST" right={
              <span className="t-xs" style={{ color: 'var(--warn)' }}>EDITING</span>
            }>
              <textarea
                value={editText}
                onChange={e => setEditText(e.target.value)}
                style={{
                  width: '100%', minHeight: 300, padding: 10,
                  background: 'var(--bg-2, rgba(0,0,0,0.3))', border: '1px solid var(--rule-2)',
                  color: 'var(--ink)', fontFamily: 'inherit', fontSize: 11,
                  letterSpacing: '0.04em', lineHeight: 1.6, resize: 'vertical',
                }}
                spellCheck={false}
              />
              <ContextBox id="deck.edit-save" style={{ marginTop: 10 }}>
                <strong>SAVE UPDATE</strong> writes a new version of the deck and re-runs Freya analysis.
                {' '}<strong>CANCEL</strong> discards your edits.
              </ContextBox>
              <div style={{ display: 'flex', gap: 8 }}>
                <Btn solid onClick={() => {
                  if (!editText.trim() || saving) return
                  setSaving(true)
                  api.updateDeck(`${owner}/${id}`, editText).then(updated => {
                    setEditing(false)
                    setSaving(false)
                    setAnalyzing(true)
                    api.getDeck(`${owner}/${id}`).then(setDeck)
                    api.getDeckVersions(`${owner}/${id}`).then(setVersions).catch(() => {})
                  }).catch(() => setSaving(false))
                }}>{saving ? 'SAVING...' : 'SAVE UPDATE'}</Btn>
                <Btn ghost onClick={() => { setEditing(false); setSaving(false) }}>CANCEL</Btn>
              </div>
              {versions.length > 0 && (
                <div style={{ marginTop: 12 }}>
                  <div className="t-xs muted" style={{ marginBottom: 6 }}>VERSION HISTORY</div>
                  {versions.slice(0, 10).map((v, i) => (
                    <div key={i} style={{ display: 'flex', justifyContent: 'space-between', padding: '3px 0', borderBottom: '1px dotted var(--rule)' }}>
                      <span className="t-xs">V{v.version}</span>
                      <span className="t-xs muted">{v.saved_at ? new Date(v.saved_at).toLocaleDateString() : ''}</span>
                    </div>
                  ))}
                </div>
              )}
            </Panel>
          )}

          {/* === ANALYSIS TAB === */}
          {activeTab === 'analysis' && <>
          <Panel code="04.C" title="FREYA / / ENGINE ANALYSIS" right={<Tag solid>Bracket B{wbs}{pls && pls !== wbs ? ` → Plays Like B${pls}` : ''}</Tag>}>
            {!analysis ? (
              <div style={{ padding: '20px 0', textAlign: 'center' }}>
                <div className="t-md muted" style={{ lineHeight: 1.8, textTransform: 'uppercase', letterSpacing: '0.04em' }}>
                  {analyzing ? (
                    <>&gt; FREYA ENGINE ANALYZING DECK<span className="blink">_</span><br />&gt; DETECTING COMBOS, SYNERGIES, WIN LINES...<br />&gt; THIS MAY TAKE A FEW SECONDS</>
                  ) : (
                    <>&gt; NO FREYA ANALYSIS ON FILE<br />&gt; RUN <span style={{ color: 'var(--ink)' }}>HEXDEK-FREYA</span> TO GENERATE STRATEGY REPORT<br />&gt; BRACKET, ARCHETYPE, WIN LINES, EVAL WEIGHTS<span className="blink">_</span></>
                  )}
                </div>
              </div>
            ) : (
              <div className="analysis-grid">
                <div>
                  <div className="t-xs muted">ARCHETYPE</div>
                  <div className="t-2xl" style={{ fontWeight: 700, marginTop: 2 }}>{archetype}</div>
                </div>
                <div className="analysis-weights">
                  <div className="t-xs muted">EVAL WEIGHTS</div>
                  {Object.entries(evalWeights).slice(0, 6).map(([k, v], i) => (
                    <div key={i} style={{ display: 'grid', gridTemplateColumns: '100px 1fr 36px', alignItems: 'center', gap: 6, marginTop: 6 }}>
                      <span className="t-xs" style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{k.replace(/_/g, ' ').toUpperCase()}</span>
                      <Bar value={v * 100} />
                      <span className="t-xs muted text-right">{Math.round(v * 100) / 100}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </Panel>

          {/* Gauntlet button — prominent, right under Freya */}
          {owner && id && (
            <div>
              <ContextBox id="deck.run-actions">
                <strong>RUN GAUNTLET (500)</strong> queues 500 AI-vs-AI games against bracket-matched meta decks on the server. Win rate, ELO delta, and best/worst matchups land in the GAUNTLET REPORT panel below; takes a few minutes.
                {' '}<strong>SPECTATE LIVE</strong> spawns a fresh 4-player room with this deck and opens the live spectator view — you can watch every decision as the AI plays it out.
                {' '}<strong>TEST VARIANT</strong> opens a scratch editor where you can swap cards and rerun Freya analysis without overwriting the saved deck.
              </ContextBox>
              <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap' }}>
              <Btn solid arrow="▶" onClick={() => {
                if (gauntlet?.status === 'running') return
                trackEvent('start_gauntlet', { deck: `${owner}/${id}`, games: 500 })
                api.startGauntlet(`${owner}/${id}`, 500).then(() => {
                  const poll = () => {
                    api.getGauntlet(`${owner}/${id}`).then(r => {
                      setGauntlet(r)
                      if (r.status === 'running') setTimeout(poll, 3000)
                    })
                  }
                  setTimeout(poll, 2000)
                })
                setGauntlet({ status: 'running', games: 0, target: 500, win_rate: 0 })
              }}>{gauntlet?.status === 'running' ? 'GAUNTLET RUNNING...' : 'RUN GAUNTLET (500)'}</Btn>
              <Btn solid arrow="▶" onClick={() => {
                if (spawningRoom) return
                setSpawningRoom(true)
                trackEvent('spawn_spectate_room', { deck: `${owner}/${id}` })
                api.spawnSpectateRoom(`${owner}/${id}`).then(r => {
                  setSpawningRoom(false)
                  if (r.room_id) navigate(`/spectate/${r.room_id}`)
                }).catch(() => setSpawningRoom(false))
              }}>{spawningRoom ? 'SPAWNING...' : 'SPECTATE LIVE'}</Btn>
              <Btn ghost arrow="▶">TEST VARIANT</Btn>
              </div>
            </div>
          )}

          {gauntlet && gauntlet.status !== 'none' && (
            <Panel code="04.G" title="GAUNTLET REPORT" right={
              <Tag solid kind={gauntlet.status === 'complete' ? 'ok' : null}>
                {gauntlet.status === 'running' ? `${gauntlet.games}/${gauntlet.target}` : gauntlet.status?.toUpperCase()}
              </Tag>
            }>
              {gauntlet.status === 'running' ? (
                <div style={{ padding: '16px 0', textAlign: 'center' }}>
                  <div className="t-md muted" style={{ lineHeight: 1.8, textTransform: 'uppercase', letterSpacing: '0.04em' }}>
                    &gt; GAUNTLET IN PROGRESS<span className="blink">_</span><br />
                    &gt; {gauntlet.games?.toLocaleString()} / {gauntlet.target?.toLocaleString()} GAMES ({gauntlet.win_rate || 0}% WIN RATE)
                  </div>
                  <Bar value={gauntlet.games / gauntlet.target * 100} />
                </div>
              ) : gauntlet.status === 'complete' ? (
                <div>
                  <div className="grid col-3" style={{ gap: 14, marginBottom: 14 }}>
                    <div>
                      <div className="t-xs muted">WIN RATE</div>
                      <div className="t-2xl" style={{ fontWeight: 700, color: gauntlet.win_rate >= 25 ? 'var(--ok)' : 'var(--danger)' }}>{gauntlet.win_rate}%</div>
                    </div>
                    <div>
                      <div className="t-xs muted">RECORD</div>
                      <div className="t-2xl" style={{ fontWeight: 700 }}><span style={{ color: 'var(--ok)' }}>{gauntlet.wins}W</span> — <span style={{ color: 'var(--danger)' }}>{gauntlet.losses}L</span></div>
                    </div>
                    <div>
                      <div className="t-xs muted">ELO DELTA</div>
                      <div className="t-2xl" style={{ fontWeight: 700, color: gauntlet.elo_delta >= 0 ? 'var(--ok)' : 'var(--danger)' }}>
                        {gauntlet.elo_delta >= 0 ? '+' : ''}{Math.round(gauntlet.elo_delta)}
                      </div>
                    </div>
                  </div>
                  <KV rows={[
                    ['GAMES', `${gauntlet.games?.toLocaleString()}`],
                    ['AVG TURNS', `${gauntlet.avg_turns}`],
                    ['ELO', `${gauntlet.elo_start} → ${gauntlet.elo_end}`],
                  ]} />
                  {gauntlet.top_beaten?.length > 0 && (
                    <>
                      <div className="hr" style={{ margin: '8px 0' }} />
                      <div className="t-xs muted" style={{ marginBottom: 4 }}>MOST BEATEN</div>
                      {gauntlet.top_beaten.map((b, i) => (
                        <div key={i} className="t-xs" style={{ color: 'var(--ok)', padding: '1px 0' }}>&gt; {b}</div>
                      ))}
                    </>
                  )}
                  {gauntlet.top_lost_to?.length > 0 && (
                    <>
                      <div className="hr" style={{ margin: '8px 0' }} />
                      <div className="t-xs muted" style={{ marginBottom: 4 }}>MOST LOST TO</div>
                      {gauntlet.top_lost_to.map((b, i) => (
                        <div key={i} className="t-xs" style={{ color: 'var(--danger)', padding: '1px 0' }}>&gt; {b}</div>
                      ))}
                    </>
                  )}
                </div>
              ) : gauntlet.status === 'error' ? (
                <div className="t-xs" style={{ color: 'var(--danger)', padding: '10px 0' }}>
                  &gt; GAUNTLET ERROR — deck may not be loaded in the engine pool. Try again or contact support.
                </div>
              ) : null}
            </Panel>
          )}

          {/* Mana Curve + Color Balance */}
          {curveData && (
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }} className="curve-grid">
              <Panel code="04.M" title="MANA CURVE">
                <ManaCurveChart
                  distribution={curveData.distribution}
                  avgCmc={curveData.avg_cmc}
                  curveShape={curveData.curve_shape}
                  warnings={curveData.warnings}
                  landCount={curveData.land_count}
                  nonlandCount={curveData.nonland_count}
                  colorByCmc={computeColorByCmc(cards)}
                />
              </Panel>
              <Panel code="04.N" title="COLOR BALANCE">
                <ColorPie demand={colorData} />
                {isMultiColor && manaProduction && colorData && (() => {
                  const MANA_COLORS = { W: '#E0EBD3', U: '#6E8FA0', B: '#3a3628', R: '#CC5C4A', G: '#82C472', C: '#8A9682' }
                  const allColors = [...new Set([...Object.keys(colorData), ...Object.keys(manaProduction)])].filter(k => (colorData[k] || 0) > 0).sort()
                  const totalProd = allColors.reduce((s, k) => s + (manaProduction[k] || 0), 0)
                  const totalDem = allColors.reduce((s, k) => s + (colorData[k] || 0), 0)
                  if (totalProd === 0 || totalDem === 0) return null
                  return (
                    <div style={{ marginTop: 12 }}>
                      <div className="t-xs muted" style={{ marginBottom: 6 }}>PRODUCTION vs DEMAND</div>
                      {allColors.map(color => {
                        const prodPct = Math.round(((manaProduction[color] || 0) / totalProd) * 100)
                        const demPct = Math.round(((colorData[color] || 0) / totalDem) * 100)
                        const diff = prodPct - demPct
                        const ok = diff >= -3
                        return (
                          <div key={color} style={{ marginBottom: 6 }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 2 }}>
                              <span className="t-xs" style={{ fontWeight: 700 }}>{color}</span>
                              <span className="t-xs" style={{ color: ok ? 'var(--ok)' : 'var(--danger)' }}>
                                {prodPct}% / {demPct}%{diff !== 0 ? ` (${diff > 0 ? '+' : ''}${diff})` : ''}
                              </span>
                            </div>
                            <div style={{ display: 'flex', gap: 1, height: 6 }}>
                              <div style={{ width: `${prodPct}%`, height: '100%', background: MANA_COLORS[color] || 'var(--ink-3)', opacity: 0.9, borderRadius: 1 }} title={`Production: ${prodPct}% (${manaProduction[color] || 0} sources)`} />
                            </div>
                            <div style={{ display: 'flex', gap: 1, height: 3, marginTop: 1 }}>
                              <div style={{ width: `${demPct}%`, height: '100%', background: 'var(--ink-3)', opacity: 0.4, borderRadius: 1 }} title={`Demand: ${demPct}% (${colorData[color] || 0} pips)`} />
                            </div>
                          </div>
                        )
                      })}
                      <div className="t-xs muted" style={{ marginTop: 4 }}>% OF SOURCES / % OF PIPS</div>
                    </div>
                  )
                })()}
                {analysis?.color_balance?.warnings?.length > 0 && (
                  <div style={{ marginTop: 8, display: 'flex', flexWrap: 'wrap', gap: 4 }}>
                    {analysis.color_balance.warnings.map((w, i) => <Tag key={i} kind="warn" solid>{w}</Tag>)}
                  </div>
                )}
              </Panel>
            </div>
          )}

          {/* Win lines */}
          {winLines.length > 0 && (() => {
            const WINLINE_CAP = 8
            const visible = winLinesExpanded ? winLines : winLines.slice(0, WINLINE_CAP)
            const hidden = winLines.length - WINLINE_CAP
            return (
            <Panel code="04.D" title={`WIN LINES / / ${winLines.length} DETECTED`}>
              {visible.map((wl, i) => {
                const kindMap = { finisher: 'bad', combat: 'warn', commander_damage: 'ok', combo: 'bad', synergy: null }
                const symbols = ['α', 'β', 'γ', 'δ', 'ε', 'ζ']
                return (
                  <div key={i} className="winline-row" style={{ padding: '10px 0', borderBottom: i < visible.length - 1 ? '1px dashed var(--rule-2)' : 'none' }}>
                    <div style={{ fontSize: 24, fontWeight: 700, color: kindMap[wl.type] === 'bad' ? 'var(--danger)' : kindMap[wl.type] === 'warn' ? 'var(--warn)' : kindMap[wl.type] === 'ok' ? 'var(--ok)' : 'var(--ink)' }}>
                      {symbols[i] || '·'}
                    </div>
                    <Tag kind={kindMap[wl.type]} solid>{wl.type?.toUpperCase()}</Tag>
                    <div>
                      <div className="t-md" style={{ fontWeight: 700 }}>{wl.pieces?.join(' + ')}</div>
                      {wl.tutor_paths && (
                        <div className="t-xs muted" style={{ marginTop: 2 }}>
                          TUTORS: {wl.tutor_paths.map(t => t.tutor).join(', ')}
                        </div>
                      )}
                    </div>
                  </div>
                )
              })}
              {!winLinesExpanded && hidden > 0 && (
                <button
                  type="button"
                  onClick={() => setWinLinesExpanded(true)}
                  style={{ width: '100%', padding: '10px 0', marginTop: 6, background: 'none', border: '1px dashed var(--rule-2)', color: 'var(--ink-2)', fontFamily: 'inherit', fontSize: 11, fontWeight: 700, letterSpacing: '0.06em', textTransform: 'uppercase', cursor: 'pointer' }}
                >
                  SHOW {hidden} MORE WIN LINE{hidden === 1 ? '' : 'S'} ↓
                </button>
              )}
              {winLinesExpanded && winLines.length > WINLINE_CAP && (
                <button
                  type="button"
                  onClick={() => setWinLinesExpanded(false)}
                  style={{ width: '100%', padding: '10px 0', marginTop: 6, background: 'none', border: '1px dashed var(--rule-2)', color: 'var(--ink-2)', fontFamily: 'inherit', fontSize: 11, fontWeight: 700, letterSpacing: '0.06em', textTransform: 'uppercase', cursor: 'pointer' }}
                >
                  COLLAPSE ↑
                </button>
              )}
            </Panel>
            )
          })()}

          {/* Win condition rationale — explains detection logic per line */}
          <WinConditionRationale winLines={winLines} />

          {/* Legality violations */}
          {legality && !legality.valid && (
            <Panel code="04.L" title="LEGALITY VIOLATIONS" right={<Tag kind="bad" solid>ILLEGAL</Tag>}>
              {legality.errors?.map((e, i) => (
                <div key={i} className="t-xs" style={{ color: 'var(--danger)', padding: '2px 0' }}>&gt; {e}</div>
              ))}
              {legality.warnings?.map((w, i) => (
                <div key={i} className="t-xs" style={{ color: 'var(--warn)', padding: '2px 0' }}>&gt; {w}</div>
              ))}
            </Panel>
          )}

          {/* Warnings: curve, color, combo */}
          {(curveWarnings.length > 0 || colorMismatch.length > 0 || comboNotes.length > 0) && (
            <Panel code="04.W" title="WARNINGS" right={<Tag kind="warn" solid>{curveWarnings.length + colorMismatch.length + comboNotes.length}</Tag>}>
              {curveWarnings.map((w, i) => (
                <div key={`c${i}`} className="t-xs" style={{ color: 'var(--warn)', padding: '2px 0' }}>&gt; CURVE: {w}</div>
              ))}
              {colorMismatch.map((w, i) => (
                <div key={`m${i}`} className="t-xs" style={{ color: 'var(--warn)', padding: '2px 0' }}>&gt; COLOR: {w}</div>
              ))}
              {comboNotes.map((w, i) => (
                <div key={`n${i}`} className="t-xs" style={{ color: 'var(--ink-2)', padding: '2px 0' }}>&gt; COMBO: {w}</div>
              ))}
            </Panel>
          )}

          {/* Meta matchups */}
          {metaMatchups.length > 0 && (
            <Panel code="04.MM" title={`META POSITIONING / / ${archetype}`}>
              <div style={{ display: 'grid', gap: 0 }}>
                {metaMatchups.map((m, i) => {
                  const ratingColor = m.rating === 'favored' ? 'var(--ok)' : m.rating === 'unfavored' ? 'var(--danger)' : 'var(--ink-2)'
                  const ratingSymbol = m.rating === 'favored' ? '▲' : m.rating === 'unfavored' ? '▼' : '—'
                  return (
                    <div key={i} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '6px 0', borderBottom: i < metaMatchups.length - 1 ? '1px dotted var(--rule)' : 'none' }}>
                      <div>
                        <span className="t-xs" style={{ fontWeight: 700 }}>vs {m.archetype?.toUpperCase()}</span>
                        {m.reason && <div className="t-xs muted" style={{ marginTop: 1 }}>{m.reason}</div>}
                      </div>
                      <Tag solid kind={m.rating === 'favored' ? 'ok' : m.rating === 'unfavored' ? 'bad' : null}>
                        {ratingSymbol} {m.rating?.toUpperCase()}
                      </Tag>
                    </div>
                  )
                })}
              </div>
            </Panel>
          )}

          {/* Vulnerable to */}
          {vulnerableTo.length > 0 && (
            <Panel code="04.V" title="VULNERABLE TO">
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
                {vulnerableTo.map((v, i) => <Tag key={i} kind="warn" solid>{v.toUpperCase()}</Tag>)}
              </div>
            </Panel>
          )}

          {/* Star cards */}
          {starCards.length > 0 && (
            <Panel code="04.S" title={`STAR CARDS / / ${starCards.length}`}>
              <div className="grid col-5 gap-2">
                {starCards.slice(0, 10).map((name, i) => (
                  <CardThumb key={i} name={name} score="★" />
                ))}
              </div>
            </Panel>
          )}

          {/* Finisher cards */}
          {finisherCards.length > 0 && (
            <Panel code="04.K" title={`WIN CONDITIONS / / ${finisherCards.length}`}>
              <div className="grid col-5 gap-2">
                {finisherCards.slice(0, 10).map((name, i) => (
                  <CardThumb key={i} name={name} />
                ))}
              </div>
            </Panel>
          )}

          {/* Value engine keys */}
          {valueKeys.length > 0 && (
            <Panel code="04.E" title={`VALUE ENGINE / / ${valueKeys.length} KEY CARDS`}>
              <div className="grid col-5 gap-2">
                {valueKeys.slice(0, 10).map((name, i) => (
                  <CardThumb key={i} name={name} />
                ))}
              </div>
            </Panel>
          )}

          {/* Value engine rationale — explains why each engine was identified */}
          <ValueEngineRationale chains={valueChains} />

          {/* Game Changer cards */}
          {gameChangerCards.length > 0 && (
            <Panel code="04.GC" title={`GAME CHANGERS / / ${gameChangerCards.length}`} right={<Tag kind="bad" solid>B4+</Tag>}>
              <div className="grid col-5 gap-2">
                {gameChangerCards.map((name, i) => (
                  <CardThumb key={i} name={name} />
                ))}
              </div>
            </Panel>
          )}

          {/* Emergent synergies */}
          {emergentSynergies.length > 0 && (
            <Panel code="04.H" title={`EMERGENT SYNERGIES / / ${emergentSynergies.length} DISCOVERED`}>
              {emergentSynergies.slice(0, 12).map((syn, i) => (
                <div key={i} style={{ padding: '6px 0', borderBottom: i < emergentSynergies.length - 1 ? '1px dashed var(--rule-2)' : 'none', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <div className="t-md" style={{ fontWeight: 700 }}>{syn.cards?.join(' + ')}</div>
                    {syn.effect_pattern && <div className="t-xs muted" style={{ marginTop: 2 }}>{syn.effect_pattern}</div>}
                  </div>
                  <div style={{ textAlign: 'right', whiteSpace: 'nowrap' }}>
                    <Tag solid kind={syn.tier >= 3 ? 'ok' : null}>T{syn.tier}</Tag>
                    {syn.observation_count > 0 && <span className="t-xs muted" style={{ marginLeft: 6 }}>{syn.observation_count}× seen</span>}
                  </div>
                </div>
              ))}
            </Panel>
          )}

          {/* Cuttable cards rationale (replaces older thumbnail-only panel) */}
          <ConsiderCuttingRationale cuts={cuttableCards} />

          {/* Tutor targets */}
          {analysis?.tutor_targets && (
            <Panel code="04.F" title="TUTOR TARGETS">
              <KV rows={analysis.tutor_targets.map((t, i) => [`TARGET.${i + 1}`, t])} />
            </Panel>
          )}

          {curse && (
            <div className="archive-curse-section">
              <CurseDisplay
                curse={curse}
                isOwner={isOwner}
                deckId={deckKey}
                onConstraintsChange={(constraints) => setCurse(c => ({ ...(c || {}), constraints }))}
              />
            </div>
          )}

          </>}

          {/* === DECK LIST TAB === */}
          {activeTab === 'decklist' && <>
          {cards.length > 0 && (
            <CardRolesGrid cards={cards} cardRoles={cardRoles} />
          )}

          {cards.length > 0 && (
            <Panel code="04.B" title={`FULL CARD LIST / / ${cards.length} ENTRIES`}>
              <div>
                {cards.map((c, i) => {
                  const linkName = (c.name || '').replace(/^COMMANDER:\s*/i, '').trim()
                  return (
                    <div key={i} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 8, padding: '3px 0', borderBottom: i < cards.length - 1 ? '1px dotted var(--rule)' : 'none' }}>
                      <CardLink name={linkName} className="t-xs" style={{ borderBottom: 'none', minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {c.name}
                      </CardLink>
                      <span style={{ display: 'flex', alignItems: 'center', gap: 6, flexShrink: 0 }}>
                        {c.mana_cost && <ManaCost cost={c.mana_cost} size={12} gap={1} />}
                        <span className="t-xs muted">{c.quantity > 1 ? `×${c.quantity}` : ''}</span>
                      </span>
                    </div>
                  )
                })}
              </div>
            </Panel>
          )}
          </>}

          {/* === ACHIEVEMENTS TAB === */}
          {activeTab === 'achievements' && <>
          {achievements && (achievements.badges?.length > 0 || achievements.total_games > 0) ? (
            <Panel
              code="04.ACH"
              title={`ACHIEVEMENTS / / ${owner?.toUpperCase() || ''}`}
              right={<Tag solid kind={achievements.badges?.length > 0 ? 'ok' : null}>{achievements.badges?.length || 0} EARNED</Tag>}
            >
              {(achievements.total_games > 0 || achievements.opponents_faced > 0) && (
                <KV rows={[
                  ['GAMES', `${achievements.total_games?.toLocaleString() || 0}`],
                  ['WINS', `${achievements.total_wins?.toLocaleString() || 0}`],
                  ['STREAK', `${achievements.current_win_streak || 0} (BEST ${achievements.max_win_streak || 0})`],
                  ['OPPONENTS', `${achievements.opponents_faced?.toLocaleString() || 0}`],
                ]} />
              )}
              {achievements.badges?.length > 0 && (
                <>
                  <div className="hr" style={{ margin: '10px 0' }} />
                  {(() => {
                    const RARITY_COLOR = {
                      common:   { border: '#8a9682', bg: 'rgba(138,150,130,0.06)', label: 'COMMON' },
                      uncommon: { border: '#6e8fa0', bg: 'rgba(110,143,160,0.08)', label: 'UNCOMMON' },
                      rare:     { border: '#d8c878', bg: 'rgba(216,200,120,0.10)', label: 'RARE' },
                      mythic:   { border: '#cc5c4a', bg: 'rgba(204,92,74,0.12)', label: 'MYTHIC' },
                      secret:   { border: '#9c6ab0', bg: 'rgba(156,106,176,0.14)', label: 'SECRET' },
                    }
                    const catalogById = {}
                    for (const b of (achievements.catalog || [])) catalogById[b.id] = b
                    return (
                      <div style={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(auto-fill, minmax(140px, 1fr))',
                        gap: 8,
                      }}>
                        {achievements.badges.map(badge => {
                          const def = catalogById[badge.id] || badge
                          const palette = RARITY_COLOR[def.rarity] || RARITY_COLOR.common
                          return (
                            <div
                              key={badge.id}
                              title={`${def.name}\n${def.description}`}
                              style={{
                                border: `2px solid ${palette.border}`,
                                background: palette.bg,
                                padding: '8px 10px',
                                display: 'flex',
                                flexDirection: 'column',
                                gap: 4,
                              }}
                            >
                              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                <span style={{ fontSize: 22, lineHeight: 1 }}>{def.icon}</span>
                                <span className="t-xs" style={{ color: palette.border, letterSpacing: '0.06em', fontWeight: 700 }}>{palette.label}</span>
                              </div>
                              <div className="t-xs" style={{ fontWeight: 700, letterSpacing: '0.04em' }}>{def.name}</div>
                              <div className="t-xs muted" style={{ lineHeight: 1.3 }}>{def.description}</div>
                              <div className="t-xs muted-2" style={{ marginTop: 2 }}>
                                {badge.awarded_at ? new Date(badge.awarded_at).toLocaleDateString() : ''}
                              </div>
                            </div>
                          )
                        })}
                      </div>
                    )
                  })()}
                </>
              )}
            </Panel>
          ) : (
            <Panel code="04.ACH" title="ACHIEVEMENTS">
              <div className="t-xs muted" style={{ padding: '20px 0', textAlign: 'center', lineHeight: 1.8 }}>
                &gt; NO ACHIEVEMENTS EARNED YET.<br />
                &gt; RUN GAMES TO UNLOCK BADGES.
              </div>
            </Panel>
          )}
          </>}
        </div>
      </div>
      {exportOpen && (
        <DeckExportModal
          deck={deck}
          deckId={id}
          onClose={() => setExportOpen(false)}
        />
      )}
      {comparePickerOpen && (
        <DeckPicker
          excludeKey={`${owner}/${id}`}
          onClose={() => setComparePickerOpen(false)}
          onPick={(d) => {
            setComparePickerOpen(false)
            navigate(`/compare/${owner}/${id}/${d.owner}/${d.id}`)
          }}
        />
      )}
    </div>
  )
}
