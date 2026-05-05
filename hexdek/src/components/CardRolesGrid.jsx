import { useState, useMemo } from 'react'
import { Panel, Tag } from './chrome'
import { cardArtUrl } from '../services/api'

const ROLE_ORDER = [
  'Threat',
  'Combo',
  'Finisher',
  'Tutor',
  'Draw',
  'Ramp',
  'Removal',
  'BoardWipe',
  'Counterspell',
  'Protection',
  'Stax',
  'Recursion',
  'Utility',
  'Land',
]

const ROLE_LABELS = {
  Ramp: 'RAMP',
  Draw: 'CARD DRAW',
  Removal: 'REMOVAL',
  BoardWipe: 'BOARD WIPES',
  Counterspell: 'COUNTERSPELLS',
  Tutor: 'TUTORS',
  Threat: 'THREATS',
  Combo: 'COMBO PIECES',
  Finisher: 'FINISHERS',
  Protection: 'PROTECTION',
  Stax: 'STAX',
  Recursion: 'RECURSION',
  Utility: 'UTILITY',
  Land: 'LANDS',
}

const ROLE_KIND = {
  Threat: 'bad',
  Combo: 'bad',
  Finisher: 'bad',
  Tutor: 'warn',
  Removal: 'warn',
  BoardWipe: 'warn',
  Counterspell: 'warn',
  Stax: 'warn',
  Ramp: 'ok',
  Draw: 'ok',
  Protection: 'ok',
  Recursion: 'ok',
}

function RoleThumb({ name, qty }) {
  const imgUrl = cardArtUrl(name)
  return (
    <div className="panel" style={{ padding: 0 }}>
      <div style={{ aspectRatio: '5/4', position: 'relative', overflow: 'hidden' }}>
        <img
          src={imgUrl}
          alt={name}
          style={{ width: '100%', height: '100%', objectFit: 'cover', filter: 'saturate(0.6) contrast(1.1)' }}
          onError={e => { e.target.style.display = 'none'; e.target.parentElement.classList.add('hatch') }}
        />
        {qty > 1 && (
          <span style={{ position: 'absolute', top: 4, right: 5, background: 'rgba(12,13,10,0.7)', padding: '0 3px', fontSize: 9, color: 'var(--ink)' }}>×{qty}</span>
        )}
      </div>
      <div style={{ padding: '3px 5px' }}>
        <div style={{ fontSize: 7, fontWeight: 700, letterSpacing: '0.04em', textTransform: 'uppercase', lineHeight: 1.1, minHeight: 14, overflow: 'hidden', textOverflow: 'ellipsis' }}>
          {name}
        </div>
      </div>
    </div>
  )
}

function ToggleTag({ active, onClick, children }) {
  return (
    <Tag
      solid={active}
      onClick={onClick}
      style={{ cursor: 'pointer', opacity: active ? 1 : 0.55 }}
    >
      {children}
    </Tag>
  )
}

export default function CardRolesGrid({ cards = [], cardRoles, code = '04.R' }) {
  const hasRoles = cardRoles && typeof cardRoles === 'object' && Object.keys(cardRoles).length > 0
  const [view, setView] = useState(hasRoles ? 'roles' : 'list')

  const groups = useMemo(() => {
    if (!hasRoles) return null
    const g = {}
    for (const [name, role] of Object.entries(cardRoles)) {
      if (!role) continue
      if (!g[role]) g[role] = []
      g[role].push(name)
    }
    for (const role of Object.keys(g)) g[role].sort((a, b) => a.localeCompare(b))
    return g
  }, [cardRoles, hasRoles])

  const orderedRoles = useMemo(() => {
    if (!groups) return []
    const known = ROLE_ORDER.filter(r => groups[r]?.length)
    const extra = Object.keys(groups).filter(r => !ROLE_ORDER.includes(r)).sort()
    return [...known, ...extra]
  }, [groups])

  const qtyByName = useMemo(() => {
    const map = {}
    for (const c of cards) {
      if (!c?.name) continue
      const clean = c.name.replace(/^COMMANDER:\s*/i, '').trim()
      map[clean] = (map[clean] || 0) + (c.quantity || 1)
    }
    return map
  }, [cards])

  const sortedCards = useMemo(() => {
    return [...cards]
      .map(c => ({ ...c, displayName: c.name?.replace(/^COMMANDER:\s*/i, '').trim() || c.name }))
      .filter(c => c.displayName)
      .sort((a, b) => a.displayName.localeCompare(b.displayName))
  }, [cards])

  const totalTagged = orderedRoles.reduce((s, r) => s + groups[r].length, 0)

  if (!hasRoles && !cards.length) return null

  const title = view === 'roles' && hasRoles
    ? `CARDS / / ${orderedRoles.length} ROLES / / ${totalTagged} TAGGED`
    : `CARDS / / ${sortedCards.length} ENTRIES`

  const toggle = (
    <span style={{ display: 'inline-flex', gap: 4 }}>
      {hasRoles && (
        <ToggleTag active={view === 'roles'} onClick={() => setView('roles')}>ROLES</ToggleTag>
      )}
      <ToggleTag active={view === 'list'} onClick={() => setView('list')}>LIST</ToggleTag>
    </span>
  )

  return (
    <Panel code={code} title={title} right={toggle}>
      {view === 'roles' && hasRoles ? (
        orderedRoles.map((role, idx) => {
          const names = groups[role]
          const label = ROLE_LABELS[role] || role.toUpperCase()
          const kind = ROLE_KIND[role] || null
          return (
            <div key={role} style={{ marginTop: idx === 0 ? 0 : 14 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6, paddingBottom: 4, borderBottom: '1px dashed var(--rule-2)' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <Tag solid kind={kind}>{label}</Tag>
                  <span className="t-xs muted">{names.length} {names.length === 1 ? 'CARD' : 'CARDS'}</span>
                </div>
              </div>
              <div className="grid col-5 gap-2">
                {names.map((name, i) => <RoleThumb key={i} name={name} qty={qtyByName[name]} />)}
              </div>
            </div>
          )
        })
      ) : (
        <div className="grid col-5 gap-2">
          {sortedCards.map((c, i) => (
            <RoleThumb key={i} name={c.displayName} qty={c.quantity} />
          ))}
        </div>
      )}
    </Panel>
  )
}
