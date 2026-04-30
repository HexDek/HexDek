import { useState } from 'react'
import { Panel, Btn } from '../components/chrome'
import { API_BASE } from '../services/api'

export default function BugReport() {
  const [type, setType] = useState('bug')
  const [page, setPage] = useState('')
  const [context, setContext] = useState('')
  const [symptom, setSymptom] = useState('')
  const [expected, setExpected] = useState('')
  const [contact, setContact] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [submitted, setSubmitted] = useState(false)
  const [error, setError] = useState(null)

  const handleSubmit = async () => {
    if (!symptom.trim()) {
      setError('Please describe what happened')
      return
    }
    setSubmitting(true)
    setError(null)
    try {
      const res = await fetch(`${API_BASE}/api/feedback`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ type, page, context, symptom, expected, contact }),
      })
      if (!res.ok) throw new Error(`Server error: ${res.status}`)
      setSubmitted(true)
    } catch (err) {
      setError(err.message)
    } finally {
      setSubmitting(false)
    }
  }

  if (submitted) {
    return (
      <div style={{ padding: '40px 30px', maxWidth: 600, margin: '0 auto', textAlign: 'center' }}>
        <Panel code="FB.OK" title="RECEIVED">
          <div className="t-md" style={{ lineHeight: 1.7, padding: 20 }}>
            <p style={{ fontSize: 18, fontWeight: 700 }}>REPORT LOGGED</p>
            <p style={{ marginTop: 12, color: 'var(--ink-2)' }}>
              We'll review this and follow up if you left contact info. Thanks for making HexDek better.
            </p>
          </div>
        </Panel>
      </div>
    )
  }

  const inputStyle = {
    width: '100%', padding: '8px 10px', background: 'var(--bg-2)', border: '1px solid var(--rule-2)',
    color: 'var(--ink)', fontFamily: 'inherit', fontSize: 12, letterSpacing: '0.02em',
  }

  return (
    <div style={{ padding: '20px 30px', maxWidth: 600, margin: '0 auto' }}>
      <Panel code="FB.0" title="BUG / SUGGESTION REPORT">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>

          <div>
            <div className="t-xs muted" style={{ marginBottom: 4 }}>TYPE</div>
            <div style={{ display: 'flex', gap: 8 }}>
              {['bug', 'suggestion'].map(t => (
                <div
                  key={t}
                  onClick={() => setType(t)}
                  style={{
                    padding: '6px 16px', cursor: 'pointer', fontSize: 11, letterSpacing: '0.05em',
                    border: type === t ? '1px solid var(--ok)' : '1px dashed var(--rule-2)',
                    background: type === t ? 'var(--bg-2)' : 'transparent',
                    color: type === t ? 'var(--ink)' : 'var(--ink-2)',
                  }}
                >
                  {t.toUpperCase()}
                </div>
              ))}
            </div>
          </div>

          <div>
            <div className="t-xs muted" style={{ marginBottom: 4 }}>WHICH PAGE / SCREEN</div>
            <input
              type="text"
              value={page}
              onChange={e => setPage(e.target.value)}
              placeholder="e.g. Spectator, Dashboard, Deck List..."
              style={inputStyle}
            />
          </div>

          <div>
            <div className="t-xs muted" style={{ marginBottom: 4 }}>CONTEXT — WHAT WERE YOU DOING</div>
            <textarea
              value={context}
              onChange={e => setContext(e.target.value)}
              placeholder="I was watching a live game and clicked on a commander..."
              rows={3}
              style={{ ...inputStyle, resize: 'vertical' }}
            />
          </div>

          <div>
            <div className="t-xs muted" style={{ marginBottom: 4 }}>
              {type === 'bug' ? 'WHAT HAPPENED (THE BUG)' : 'YOUR SUGGESTION'} *
            </div>
            <textarea
              value={symptom}
              onChange={e => setSymptom(e.target.value)}
              placeholder={type === 'bug'
                ? "The volcmap showed all zeros even though the game was running..."
                : "It would be cool if the leaderboard showed win streaks..."
              }
              rows={4}
              style={{ ...inputStyle, resize: 'vertical', borderColor: error ? 'var(--danger)' : 'var(--rule-2)' }}
            />
          </div>

          {type === 'bug' && (
            <div>
              <div className="t-xs muted" style={{ marginBottom: 4 }}>WHAT SHOULD HAVE HAPPENED</div>
              <textarea
                value={expected}
                onChange={e => setExpected(e.target.value)}
                placeholder="The volcmap should show the eval weights updating each turn..."
                rows={3}
                style={{ ...inputStyle, resize: 'vertical' }}
              />
            </div>
          )}

          <div>
            <div className="t-xs muted" style={{ marginBottom: 4 }}>CONTACT (OPTIONAL — FOR FOLLOWUP)</div>
            <input
              type="text"
              value={contact}
              onChange={e => setContact(e.target.value)}
              placeholder="email or Discord username"
              style={inputStyle}
            />
            <div className="t-xs muted-2" style={{ marginTop: 2 }}>We'll let you know when it's fixed</div>
          </div>

          {error && <div className="t-xs" style={{ color: 'var(--danger)' }}>{error}</div>}

          <Btn onClick={handleSubmit} disabled={submitting}>
            {submitting ? 'SUBMITTING...' : 'SUBMIT REPORT'}
          </Btn>
        </div>
      </Panel>
    </div>
  )
}
