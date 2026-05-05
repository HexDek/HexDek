import { useNavigate } from 'react-router-dom'
import { Btn } from './chrome'

export default function SignInPrompt({ onClose }) {
  const navigate = useNavigate()
  return (
    <div style={{
      position: 'fixed', inset: 0, zIndex: 1000,
      background: 'rgba(0,0,0,0.7)', display: 'flex', alignItems: 'center', justifyContent: 'center',
    }} onClick={onClose}>
      <div className="panel" onClick={e => e.stopPropagation()} style={{ maxWidth: 420, width: '100%' }}>
        <div className="panel-hd">
          <span>SIGN-IN REQUIRED</span>
          <span style={{ cursor: 'pointer' }} onClick={onClose}>X</span>
        </div>
        <div className="panel-bd" style={{ display: 'flex', flexDirection: 'column', gap: 14, padding: 18 }}>
          <div className="t-md" style={{ lineHeight: 1.6, textTransform: 'uppercase', letterSpacing: '0.04em' }}>
            &gt; SIGN IN TO UPLOAD YOUR DECK.
            <br />
            &gt; MAGIC-LINK AUTH. NO PASSWORD.
            <br />
            &gt; YOUR BUILDS, YOUR ELO, YOUR ARCHIVE.
          </div>
          <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
            <Btn sm ghost onClick={onClose}>CANCEL</Btn>
            <Btn sm solid arrow="↗" onClick={() => { onClose(); navigate('/login') }}>SIGN IN</Btn>
          </div>
        </div>
      </div>
    </div>
  )
}
