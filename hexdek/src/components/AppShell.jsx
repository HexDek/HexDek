import { useState, useEffect } from 'react'
import { NavLink, Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Crops } from './chrome'
import SearchBar from './SearchBar'
import { useAuth } from '../context/AuthContext'

const PUBLIC_NAV = [
  { to: '/decks', label: 'DECKS' },
  { to: '/leaderboard', label: 'RANKINGS' },
  { to: '/spectate', label: 'SPECTATE' },
]

const AUTH_NAV = [
  { to: '/decks', label: 'DECKS' },
  { to: '/leaderboard', label: 'RANKINGS' },
  { to: '/spectate', label: 'SPECTATE' },
  { to: '/decks?tab=mine', label: 'MY DECKS' },
]

function useTheme() {
  const [theme, setTheme] = useState(() => {
    try { return localStorage.getItem('hexdek.theme') || 'dark' } catch { return 'dark' }
  })
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
    try { localStorage.setItem('hexdek.theme', theme) } catch {}
  }, [theme])
  return [theme, () => setTheme(t => t === 'dark' ? 'light' : 'dark')]
}

export default function AppShell() {
  const { user, loading, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const nav = user ? AUTH_NAV : PUBLIC_NAV
  const [theme, toggleTheme] = useTheme()

  const isNavActive = (to) => {
    const [path, query = ''] = to.split('?')
    if (location.pathname !== path) return false
    if (query) {
      const params = new URLSearchParams(query)
      const current = new URLSearchParams(location.search)
      for (const [k, v] of params) if (current.get(k) !== v) return false
      return true
    }
    // Plain path (e.g. /decks): active only when no `tab` query is set, so MY DECKS
    // and DECKS don't both light up.
    return !new URLSearchParams(location.search).get('tab')
  }

  const handleLogout = async () => {
    await logout()
    navigate('/')
  }

  return (
    <div style={{ height: '100vh', background: 'var(--bg)', position: 'relative', overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
      <span className="grain" />
      <div className="frame" style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        <Crops />

        <div className="appbar">
          <div className="flex items-center gap-4">
            <NavLink to={user ? '/dash' : '/'} className="brand" style={{ textDecoration: 'none' }}>HEXDEK//</NavLink>
            <nav>
              {nav.map(n => (
                <NavLink
                  key={n.label}
                  to={n.to}
                  className={isNavActive(n.to) ? 'on' : ''}
                >
                  {n.label}
                </NavLink>
              ))}
            </nav>
          </div>
          <SearchBar />
          <span style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            SYS.BUILD 25.04.28 · HEXDEK V0.10D
            <button onClick={toggleTheme} className="btn--sm btn--ghost" style={{
              padding: '2px 6px', border: '1px solid var(--rule-2)', background: 'transparent',
              color: 'var(--ink-2)', cursor: 'pointer', fontFamily: 'inherit', fontSize: 9,
              letterSpacing: '0.08em', textTransform: 'uppercase',
            }}>{theme === 'dark' ? '☽ DARK' : '☀ LIGHT'}</button>
          </span>
          {!loading && (
            <span style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              {user ? (
                <>
                  <NavLink to="/profile" className="t-xs" style={{ color: 'var(--ok)', textDecoration: 'none' }}>● {user.email?.split('@')[0]?.toUpperCase()}</NavLink>
                  <a onClick={handleLogout} style={{ cursor: 'pointer', fontSize: 9, letterSpacing: '0.1em', color: 'var(--ink-2)' }}>LOGOUT</a>
                </>
              ) : (
                <NavLink to="/login" style={{ fontSize: 10, letterSpacing: '0.1em', color: 'var(--accent)', textDecoration: 'none', fontWeight: 700, border: '1px solid var(--rule-2)', padding: '3px 8px' }}>SIGN IN ↗</NavLink>
              )}
            </span>
          )}
        </div>

        <div style={{ flex: 1, overflow: 'auto', display: 'flex', flexDirection: 'column' }}>
          <Outlet />
        </div>

        <div className="statusbar">
          <span>+ + +  HEXDEK CORE READY  + + +</span>
          <NavLink to="/about" style={{ color: 'var(--ink-2)', textDecoration: 'none', fontSize: 9, letterSpacing: '0.08em', fontWeight: 700 }}>ABOUT</NavLink>
          <NavLink to="/feedback" style={{ color: 'var(--danger)', textDecoration: 'none', fontSize: 9, letterSpacing: '0.08em', fontWeight: 700 }}>BUG / SUGGESTION</NavLink>
          <NavLink to="/donations" style={{ color: 'var(--ok)', textDecoration: 'none', fontSize: 9, letterSpacing: '0.08em', fontWeight: 700 }}>DONATE ♥</NavLink>
          <a href="https://discord.gg/Mz2ueRFXds" target="_blank" rel="noopener noreferrer" style={{ color: 'var(--ink-2)', textDecoration: 'none', fontSize: 9, letterSpacing: '0.08em', fontWeight: 700 }}>DISCORD</a>
          <span>OPEN SOURCE / / DONATIONS-POWERED / / NO ADS</span>
          <span>{user ? `USR.${user.email?.split('@')[0]?.toUpperCase()}` : 'GUEST'}</span>
        </div>
      </div>
    </div>
  )
}
