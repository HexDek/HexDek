import { useState } from 'react'
import { Btn } from './chrome'
import { api } from '../services/api'
import { trackEvent } from '../hooks/useAnalytics'

export default function ImportModal({ onClose, onImported }) {
  const [name, setName] = useState('')
  const [owner, setOwner] = useState('')
  const [deckList, setDeckList] = useState('')
  const [importing, setImporting] = useState(false)
  const [error, setError] = useState(null)

  const handleSubmit = async () => {
    if (!deckList.trim()) {
      setError('DECK LIST REQUIRED')
      return
    }
    setImporting(true)
    setError(null)
    try {
      trackEvent('import_deck', { name: name || 'Imported Deck', owner: owner || 'imported' })
      await api.importDeck(name || 'Imported Deck', owner || 'imported', deckList)
      onImported && onImported()
      onClose()
    } catch (err) {
      setError(err.message)
    } finally {
      setImporting(false)
    }
  }

  return (
    <div style={{
      position: 'fixed', inset: 0, zIndex: 1000,
      background: 'rgba(0,0,0,0.7)', display: 'flex', alignItems: 'center', justifyContent: 'center',
    }} onClick={onClose}>
      <div className="panel import-modal" onClick={e => e.stopPropagation()}>
        <div className="panel-hd">
          <span>DECK IMPORT / / PASTE LIST</span>
          <span style={{ cursor: 'pointer' }} onClick={onClose}>X</span>
        </div>
        <div className="panel-bd" style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div className="grid col-2" style={{ gap: 10 }}>
            <div>
              <div className="t-xs muted" style={{ marginBottom: 4 }}>DECK NAME</div>
              <input
                type="text"
                value={name}
                onChange={e => setName(e.target.value)}
                placeholder="TINYBONES, THE PICKPOCKET"
                style={{
                  width: '100%', padding: '8px 10px', background: 'var(--bg-2)',
                  border: '1px solid var(--rule-2)', color: 'var(--ink)',
                  fontFamily: 'inherit', fontSize: 11, letterSpacing: '0.06em',
                  textTransform: 'uppercase', outline: 'none',
                }}
              />
            </div>
            <div>
              <div className="t-xs muted" style={{ marginBottom: 4 }}>OWNER</div>
              <input
                type="text"
                value={owner}
                onChange={e => setOwner(e.target.value)}
                placeholder="IMPORTED"
                style={{
                  width: '100%', padding: '8px 10px', background: 'var(--bg-2)',
                  border: '1px solid var(--rule-2)', color: 'var(--ink)',
                  fontFamily: 'inherit', fontSize: 11, letterSpacing: '0.06em',
                  textTransform: 'uppercase', outline: 'none',
                }}
              />
            </div>
          </div>
          <div>
            <div className="t-xs muted" style={{ marginBottom: 4 }}>DECK LIST (1 CARD NAME PER LINE)</div>
            <textarea
              value={deckList}
              onChange={e => setDeckList(e.target.value)}
              placeholder={"COMMANDER: Tinybones, the Pickpocket\n1 Swamp\n1 Dark Ritual\n1 Sol Ring\n1 Thoughtseize\n..."}
              rows={14}
              style={{
                width: '100%', padding: '8px 10px', background: 'var(--bg-2)',
                border: '1px solid var(--rule-2)', color: 'var(--ink)',
                fontFamily: 'inherit', fontSize: 11, letterSpacing: '0.04em',
                outline: 'none', resize: 'vertical',
              }}
            />
          </div>
          <div className="t-xs muted-2">
            FORMAT: "COMMANDER: Card Name" ON FIRST LINE, THEN "1 Card Name" OR JUST "Card Name" PER LINE.
            SIDEBOARD/MAYBEBOARD SECTIONS ARE IGNORED. COMMENTS (#) AND BLANK LINES IGNORED.
          </div>
          {error && <div className="t-xs" style={{ color: 'var(--danger)' }}>&gt; ERROR: {error}</div>}
          <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
            <Btn sm ghost onClick={onClose}>CANCEL</Btn>
            <Btn sm solid onClick={handleSubmit} arrow={importing ? '...' : '↗'}>
              {importing ? 'IMPORTING' : 'IMPORT'}
            </Btn>
          </div>
        </div>
      </div>
    </div>
  )
}
