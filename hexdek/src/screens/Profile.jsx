import { useState, useEffect } from 'react'
import { Panel, Btn, KV } from '../components/chrome'
import { useAuth } from '../context/AuthContext'

const inputStyle = {
  width: '100%', padding: '8px 10px', background: 'var(--bg-2)', border: '1px solid var(--rule-2)',
  color: 'var(--ink)', fontFamily: 'inherit', fontSize: 12, letterSpacing: '0.02em',
}

export default function Profile() {
  const { user } = useAuth()
  const [displayName, setDisplayName] = useState('')
  const [owner, setOwner] = useState('')
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    setDisplayName(localStorage.getItem('hexdek_display_name') || '')
    setOwner(localStorage.getItem('hexdek_owner') || '')
  }, [])

  const handleSave = () => {
    localStorage.setItem('hexdek_display_name', displayName.trim())
    localStorage.setItem('hexdek_owner', owner.trim())
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  return (
    <div style={{ padding: '20px 30px', maxWidth: 600, margin: '0 auto' }}>
      <Panel code="USR.0" title="USER PROFILE">
        <KV rows={[
          ['AUTH EMAIL', user?.email || '— not signed in —'],
          ['UID', user?.uid || '—'],
          ['STATUS', user ? 'AUTHENTICATED' : 'GUEST'],
        ]} />
      </Panel>

      <Panel code="USR.1" title="DISPLAY PREFERENCES" style={{ marginTop: 16 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>

          <div>
            <div className="t-xs muted" style={{ marginBottom: 4 }}>DISPLAY NAME</div>
            <input
              type="text"
              value={displayName}
              onChange={e => setDisplayName(e.target.value)}
              placeholder="How you want to appear in lobbies and reports"
              style={inputStyle}
            />
            <div className="t-xs muted-2" style={{ marginTop: 2 }}>
              Shown in spectator chats, party lobbies, and game reports.
            </div>
          </div>

          <div>
            <div className="t-xs muted" style={{ marginBottom: 4 }}>OWNER NAME</div>
            <input
              type="text"
              value={owner}
              onChange={e => setOwner(e.target.value)}
              placeholder="e.g. josh, kylie, blake..."
              style={inputStyle}
            />
            <div className="t-xs muted-2" style={{ marginTop: 2 }}>
              Controls which decks appear under "My Decks". Match the folder name in <code>data/decks/</code>.
            </div>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <Btn onClick={handleSave}>SAVE PREFERENCES</Btn>
            {saved && <span className="t-xs" style={{ color: 'var(--ok)' }}>● SAVED</span>}
          </div>

          <div className="t-xs muted-2">
            Preferences are stored locally in your browser only.
          </div>
        </div>
      </Panel>
    </div>
  )
}
