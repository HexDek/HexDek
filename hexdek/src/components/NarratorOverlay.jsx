import { useMemo } from 'react'
import { findGameChangerInText } from '../data/gameChangers'

// NarratorOverlay — plain-English commentary strip.
//
// Reads the engine's per-game log array (each entry has turn, seat,
// action, detail, kind) and renders the last 10 narrated events as a
// semi-transparent bar pinned to the bottom of the spectator view.
// This is the foundation for Twitch / OBS overlay integration: the
// generated sentences are stable, plain English (no ALL-CAPS terminal
// styling on the prose itself, so streams can pipe them through TTS
// or chat without further processing).
//
// The component is purely derived from props — no state, no fetches.
// Mount once at the bottom of the spectator view; pass game.log and
// game.seats.

const MAX_EVENTS = 10

// Tone classes drive the left-edge accent stripe. Loud events (combat,
// game-changers, eliminations) get a colored stripe so streamers can
// pick them out at a glance.
const TONE = {
  combat:    'narrator-row--combat',
  cast:      'narrator-row--cast',
  elim:      'narrator-row--elim',
  changer:   'narrator-row--changer',
  event:     'narrator-row--event',
}

function playerLabel(seat, seats) {
  const s = seats?.[seat]
  if (s?.commander) return s.commander.split('//')[0].trim()
  return `Player ${seat + 1}`
}

// Strip the "<COMMANDER> " prefix the backend prepends to most action
// strings. Returns the action text without the actor, so we can
// substitute a friendlier player label in front.
function stripActor(action, actor) {
  if (!action) return ''
  const upperActor = (actor || '').toUpperCase()
  if (upperActor && action.startsWith(upperActor + ' ')) {
    return action.slice(upperActor.length + 1)
  }
  return action
}

// narrate transforms one LogEntry into { text, tone } for display.
// Returns null when the entry doesn't surface in the narrator (turn
// markers, internal events, etc.) so the consumer can skip it.
function narrate(entry, seats) {
  if (!entry || !entry.action) return null
  const actor = playerLabel(entry.seat, seats)
  const actorRaw = seats?.[entry.seat]?.commander?.split('//')[0].trim() || ''
  const tail = stripActor(entry.action, actorRaw)
  const lower = tail.toLowerCase()
  const detail = entry.detail || ''

  // Game-changer detection wins over the kind switch — the engine fires
  // the same enter_battlefield kind for any permanent ETB; we only want
  // to call out the meta-warping ones.
  const gc = findGameChangerInText(entry.action)
  if (gc) {
    return {
      text: `GAME CHANGER · ${gc.toUpperCase()} hits the battlefield (${actor}).`,
      tone: 'changer',
    }
  }

  switch (entry.kind) {
    case 'combat': {
      // Engine emits "X ATTACKS WITH N CREATURE(S)" or
      // "X ATTACKS WITH N CREATURE(S) [→ Y]". Re-cast as a sentence.
      // P/T isn't in the log payload today; if added later, splice it in.
      const m = lower.match(/attacks with (\d+) creature/)
      if (m) {
        const n = parseInt(m[1], 10)
        const target = lower.includes('→') ? lower.split('→')[1].trim() : null
        return {
          text: target
            ? `${actor} attacks ${target.replace(/\.$/, '')} with ${n} creature${n === 1 ? '' : 's'}.`
            : `${actor} attacks with ${n} creature${n === 1 ? '' : 's'}.`,
          tone: 'combat',
        }
      }
      return { text: `${actor} ${tail.toLowerCase()}`, tone: 'combat' }
    }
    case 'cast': {
      // tail is "CASTS <NAME>". Narrate as "X casts NAME."
      const m = tail.match(/^CASTS (.+)$/i)
      if (m) return { text: `${actor} casts ${titleCase(m[1])}.`, tone: 'cast' }
      return { text: `${actor} casts a spell.`, tone: 'cast' }
    }
    case 'land': {
      const m = tail.match(/^PLAYS LAND: (.+)$/i)
      if (m) return { text: `${actor} plays ${titleCase(m[1])}.`, tone: 'event' }
      return null
    }
    case 'enter_battlefield':
    case 'etb': {
      const m = tail.match(/(?:→ ETB|ETB|ENTERS):? *(.+)$/i)
      const name = m ? titleCase(m[1]) : detail || 'a permanent'
      return { text: `${name} enters under ${actor}.`, tone: 'event' }
    }
    case 'elimination':
    case 'player_lost': {
      // detail often carries the reason ("commander damage", "lethal", …).
      const reason = detail ? ` (${detail})` : ''
      return { text: `${actor} eliminated${reason}.`, tone: 'elim' }
    }
    case 'damage': {
      const m = tail.match(/DEALS (\d+) DAMAGE TO (.+)$/i)
      if (m) return {
        text: `${actor} deals ${m[1]} damage to ${titleCase(m[2])}${detail ? ` (${detail})` : ''}.`,
        tone: 'event',
      }
      return null
    }
    case 'counter': {
      const m = tail.match(/^COUNTERS (.+)$/i)
      if (m) return { text: `${actor} counters ${titleCase(m[1])}.`, tone: 'cast' }
      return null
    }
    case 'removal': {
      const m = tail.match(/^(DESTROYS|EXILES|SACRIFICES) (.+)$/i)
      if (m) {
        const verb = m[1].toLowerCase().replace(/s$/, 'es').replace(/ses$/, 'ces')
        return { text: `${actor} ${m[1].toLowerCase()} ${titleCase(m[2])}.`, tone: 'event' }
      }
      return null
    }
    case 'token': {
      const m = tail.match(/^CREATES TOKEN: (.+)$/i)
      if (m) return { text: `${actor} creates a ${titleCase(m[1])} token.`, tone: 'event' }
      return null
    }
    case 'reanimate': {
      const m = tail.match(/^REANIMATES (.+)$/i)
      if (m) return { text: `${actor} reanimates ${titleCase(m[1])}.`, tone: 'changer' }
      return null
    }
    case 'extra_turn':
      return { text: `${actor} takes an extra turn.`, tone: 'changer' }
    case 'mill': {
      const m = tail.match(/MILLS (\d+) CARDS?/i)
      if (m) return { text: `${actor} mills ${m[1]} card${m[1] === '1' ? '' : 's'}.`, tone: 'event' }
      return null
    }
    case 'draw':
    case 'life':
    case 'trigger':
    case 'activate':
    case 'search':
    default:
      // Skip the chatter; the four spec'd kinds plus changer/etb/etc.
      // are already covered. Returning null keeps the strip uncluttered.
      return null
  }
}

function titleCase(s) {
  if (!s) return ''
  return s.toLowerCase().replace(/\b\w/g, c => c.toUpperCase())
}

export { narrate }

export default function NarratorOverlay({ log, seats }) {
  const lines = useMemo(() => {
    if (!Array.isArray(log) || log.length === 0) return []
    const out = []
    // Walk newest → oldest so we don't waste work narrating events
    // that'll be cropped before reaching the top of the strip.
    for (let i = log.length - 1; i >= 0 && out.length < MAX_EVENTS; i--) {
      const n = narrate(log[i], seats)
      if (!n) continue
      out.push({ ...n, turn: log[i].turn, key: `${i}-${log[i].kind}` })
    }
    return out.reverse() // oldest first so newest is at the bottom (most-visible) edge
  }, [log, seats])

  if (lines.length === 0) return null

  return (
    <div className="narrator-overlay" role="log" aria-live="polite" aria-atomic="false">
      <div className="narrator-overlay-hd">
        <span>NARRATOR / / LIVE COMMENTARY</span>
        <span className="narrator-overlay-count">{lines.length} / {MAX_EVENTS}</span>
      </div>
      <div className="narrator-overlay-rows">
        {lines.map((l, idx) => (
          <div
            key={l.key}
            className={`narrator-row ${TONE[l.tone] || ''}${idx === lines.length - 1 ? ' narrator-row--latest' : ''}`}
          >
            <span className="narrator-row-turn">T{l.turn}</span>
            <span className="narrator-row-text">{l.text}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
