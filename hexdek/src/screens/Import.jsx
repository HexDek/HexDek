import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Panel, Tag, Btn, Tape } from '../components/chrome'
import { api } from '../services/api'
import { useAuth } from '../context/AuthContext'
import { trackEvent } from '../hooks/useAnalytics'

// Import — full-page deck-import flow at /import.
//
// The backend's POST /api/decks/import (alias of POST /api/decks) is
// format-tolerant: it accepts MTGO ("2 Sol Ring"), Arena (same shape
// with sideboard markers), or plain "Sol Ring" lines, plus an
// optional leading "COMMANDER: <Card>" line that pins the commander
// for Freya / showmatch consumers. parseDeckList strips quantity
// prefixes, set codes in parens, and ignores comments + blank lines.
//
// This page surfaces the commander as its own field for clarity:
// when the user fills it in we prepend the canonical "COMMANDER:"
// line (or replace an existing one) so the parser picks it up
// regardless of where the user pasted it.

const inputStyle = {
  width: '100%', padding: '8px 10px', background: 'var(--bg-2)',
  border: '1px solid var(--rule-2)', color: 'var(--ink)',
  fontFamily: 'inherit', fontSize: 11, letterSpacing: '0.06em',
  textTransform: 'uppercase', outline: 'none',
}

const SAMPLE = `COMMANDER: Tinybones, the Pickpocket
1 Swamp
1 Sol Ring
1 Dark Ritual
1 Thoughtseize
...`

// inferCommander pulls a "COMMANDER: <name>" line out of the deck list
// if the user pasted one, so we can pre-fill the commander field.
function inferCommander(text) {
  if (!text) return ''
  for (const raw of text.split('\n')) {
    const line = raw.trim()
    if (/^commander\s*:/i.test(line)) {
      return line.replace(/^commander\s*:/i, '').trim()
    }
  }
  return ''
}

// stripCommanderLine drops any existing COMMANDER: line so we can
// prepend our own without producing duplicates.
function stripCommanderLine(text) {
  return text
    .split('\n')
    .filter(line => !/^\s*commander\s*:/i.test(line))
    .join('\n')
}

// countDecklistLines returns a (cards, lines) tuple for the live
// preview header. Comment + blank lines don't count as cards.
function countDecklistLines(text) {
  let cards = 0
  let lines = 0
  for (const raw of text.split('\n')) {
    const line = raw.trim()
    if (!line) continue
    lines++
    if (line.startsWith('#') || line.startsWith('//')) continue
    if (/^commander\s*:/i.test(line)) continue
    if (/^sideboard|^maybeboard/i.test(line)) continue
    cards++
  }
  return { cards, lines }
}

export default function Import() {
  const navigate = useNavigate()
  const { user } = useAuth()

  // Default owner: same heuristic as Dashboard / DeckArchive — Firebase
  // doesn't carry a HexDek username, so we infer it from displayName /
  // email-prefix, with a localStorage override.
  const defaultOwner = useMemo(() => {
    if (!user) return ''
    return (
      localStorage.getItem('hexdek_owner')
      || user.displayName?.toLowerCase()
      || user.email?.split('@')[0]?.split('.')[0]
      || ''
    )
  }, [user])

  const [name, setName] = useState('')
  const [owner, setOwner] = useState(defaultOwner)
  const [commander, setCommander] = useState('')
  const [deckList, setDeckList] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState(null)

  const counts = useMemo(() => countDecklistLines(deckList), [deckList])

  // Detect commander from the textarea on each change so the field
  // stays in sync if the user pasted a list with COMMANDER: included.
  // Only auto-fills when the user hasn't typed their own value yet.
  const handleDeckListChange = (e) => {
    const text = e.target.value
    setDeckList(text)
    if (!commander) {
      const inferred = inferCommander(text)
      if (inferred) setCommander(inferred)
    }
  }

  const handleSubmit = async () => {
    setError(null)
    if (!deckList.trim()) {
      setError('PASTE A DECK LIST FIRST.')
      return
    }
    if (counts.cards < 5) {
      setError(`ONLY ${counts.cards} CARD${counts.cards === 1 ? '' : 'S'} DETECTED — DOUBLE-CHECK THE PASTE.`)
      return
    }

    // Normalize: if a commander was supplied, strip any existing
    // COMMANDER: line and prepend the canonical one.
    let payloadList = deckList
    const cmdrTrimmed = commander.trim()
    if (cmdrTrimmed) {
      payloadList = `COMMANDER: ${cmdrTrimmed}\n${stripCommanderLine(deckList)}`
    }

    setSubmitting(true)
    try {
      const result = await api.importDeckFull({
        name: name.trim() || 'Imported Deck',
        owner: owner.trim() || 'imported',
        deckList: payloadList,
      })
      trackEvent('import_deck_full', {
        name: name.trim() || 'Imported Deck',
        owner: owner.trim() || 'imported',
        cards: counts.cards,
      })
      // Backend returns { id, owner, ... }. Send the user to the new deck page.
      const newOwner = result?.owner || owner.trim() || 'imported'
      const newID = result?.id
      if (newID) {
        navigate(`/decks/${encodeURIComponent(newOwner)}/${encodeURIComponent(newID)}`)
      } else {
        // Fallback — listing page. Still better than staying on the form.
        navigate('/decks')
      }
    } catch (err) {
      setError((err && err.message) || 'IMPORT FAILED')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <>
      <Tape
        left="DECK IMPORT / / NEW SUBMISSION"
        mid={counts.cards > 0 ? `${counts.cards} CARDS DETECTED` : 'AWAITING PASTE'}
        right="DOC HX-110"
      />

      <div style={{ padding: '20px 30px', maxWidth: 920, margin: '0 auto' }}>
        <Panel code="IMP.A" title="DECK METADATA" solid>
          <div className="grid col-2" style={{ gap: 12 }}>
            <div>
              <div className="t-xs muted" style={{ marginBottom: 4 }}>DECK NAME</div>
              <input
                type="text"
                value={name}
                onChange={e => setName(e.target.value)}
                placeholder="TINYBONES, THE PICKPOCKET"
                style={inputStyle}
              />
              <div className="t-xs muted-2" style={{ marginTop: 4 }}>FILENAME-SAFE; A NUMERIC SUFFIX IS ADDED IF A DECK BY THIS NAME EXISTS.</div>
            </div>
            <div>
              <div className="t-xs muted" style={{ marginBottom: 4 }}>OWNER SLUG</div>
              <input
                type="text"
                value={owner}
                onChange={e => setOwner(e.target.value)}
                placeholder={defaultOwner ? defaultOwner.toUpperCase() : 'IMPORTED'}
                style={inputStyle}
              />
              <div className="t-xs muted-2" style={{ marginTop: 4 }}>USED IN THE URL: /decks/&lt;OWNER&gt;/&lt;ID&gt;.</div>
            </div>
          </div>
          <div style={{ marginTop: 12 }}>
            <div className="t-xs muted" style={{ marginBottom: 4 }}>COMMANDER (OPTIONAL)</div>
            <input
              type="text"
              value={commander}
              onChange={e => setCommander(e.target.value)}
              placeholder='AUTO-DETECTED FROM "COMMANDER:" LINE — OR TYPE TO OVERRIDE'
              style={inputStyle}
            />
            <div className="t-xs muted-2" style={{ marginTop: 4 }}>
              IF SET, THE BACKEND WRITES "COMMANDER: &lt;NAME&gt;" AS THE FIRST LINE OF THE STORED DECK FILE.
            </div>
          </div>
        </Panel>

        <div style={{ height: 14 }} />

        <Panel
          code="IMP.B"
          title={`DECK LIST / / ${counts.cards} CARDS · ${counts.lines} LINES`}
          right={<Tag solid>PASTE</Tag>}
        >
          <div className="t-xs muted" style={{ marginBottom: 4 }}>
            FORMATS ACCEPTED: MTGO ("2 SOL RING"), ARENA (SAME SHAPE + SIDEBOARD MARKERS), OR PLAIN "SOL RING" PER LINE.
          </div>
          <textarea
            value={deckList}
            onChange={handleDeckListChange}
            placeholder={SAMPLE}
            rows={20}
            style={{
              width: '100%', padding: '10px 12px', background: 'var(--bg-2)',
              border: '1px solid var(--rule-2)', color: 'var(--ink)',
              fontFamily: 'inherit', fontSize: 12, letterSpacing: '0.02em',
              outline: 'none', resize: 'vertical', lineHeight: 1.55,
            }}
            spellCheck={false}
          />
          <div className="t-xs muted-2" style={{ marginTop: 6, lineHeight: 1.55 }}>
            &gt; SET CODES IN PARENS LIKE "(M21)" ARE STRIPPED.<br />
            &gt; COMMENTS (#, //) AND BLANK LINES ARE IGNORED.<br />
            &gt; SIDEBOARD / MAYBEBOARD SECTIONS ARE NOT IMPORTED.
          </div>
        </Panel>

        {error && (
          <div className="panel" style={{ marginTop: 14, borderColor: 'var(--danger)' }}>
            <div className="panel-bd t-xs" style={{ color: 'var(--danger)', letterSpacing: '0.04em' }}>
              &gt; {error}
            </div>
          </div>
        )}

        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 18, gap: 10 }}>
          <div className="t-xs muted-2">
            DECK FILE WILL BE WRITTEN TO data/decks/{(owner.trim() || 'imported').toLowerCase()}/&lt;ID&gt;.TXT
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <Btn ghost onClick={() => navigate(-1)} arrow="←">CANCEL</Btn>
            <Btn solid onClick={handleSubmit} arrow={submitting ? '...' : '↗'}>
              {submitting ? 'IMPORTING' : 'IMPORT DECK'}
            </Btn>
          </div>
        </div>
      </div>
    </>
  )
}
