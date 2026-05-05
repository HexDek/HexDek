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

function RoleThumb({ name }) {
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
      </div>
      <div style={{ padding: '3px 5px' }}>
        <div style={{ fontSize: 7, fontWeight: 700, letterSpacing: '0.04em', textTransform: 'uppercase', lineHeight: 1.1, minHeight: 14, overflow: 'hidden', textOverflow: 'ellipsis' }}>
          {name}
        </div>
      </div>
    </div>
  )
}

export default function CardRolesGrid({ cardRoles, code = '04.R' }) {
  if (!cardRoles || typeof cardRoles !== 'object') return null

  const groups = {}
  for (const [name, role] of Object.entries(cardRoles)) {
    if (!role) continue
    if (!groups[role]) groups[role] = []
    groups[role].push(name)
  }
  for (const role of Object.keys(groups)) {
    groups[role].sort((a, b) => a.localeCompare(b))
  }

  const knownRoles = ROLE_ORDER.filter(r => groups[r]?.length)
  const extraRoles = Object.keys(groups).filter(r => !ROLE_ORDER.includes(r)).sort()
  const orderedRoles = [...knownRoles, ...extraRoles]

  if (orderedRoles.length === 0) return null

  const totalTagged = orderedRoles.reduce((s, r) => s + groups[r].length, 0)

  return (
    <Panel code={code} title={`CARD ROLES / / ${orderedRoles.length} GROUPS / / ${totalTagged} TAGGED`}>
      {orderedRoles.map((role, idx) => {
        const cards = groups[role]
        const label = ROLE_LABELS[role] || role.toUpperCase()
        const kind = ROLE_KIND[role] || null
        return (
          <div key={role} style={{ marginTop: idx === 0 ? 0 : 14 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6, paddingBottom: 4, borderBottom: '1px dashed var(--rule-2)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <Tag solid kind={kind}>{label}</Tag>
                <span className="t-xs muted">{cards.length} {cards.length === 1 ? 'CARD' : 'CARDS'}</span>
              </div>
            </div>
            <div className="grid col-5 gap-2">
              {cards.map((name, i) => <RoleThumb key={i} name={name} />)}
            </div>
          </div>
        )
      })}
    </Panel>
  )
}
