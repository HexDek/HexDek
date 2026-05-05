import { useState, useEffect } from 'react'
import { NavLink, Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Crops } from './chrome'
import SearchBar from './SearchBar'
import { ToastHost } from './Toast'
import { useAuth } from '../context/AuthContext'
import { useLiveSocket } from '../hooks/useLiveSocket'
import { useTranslation } from '../i18n'

const PUBLIC_NAV = [
  { to: '/decks', key: 'nav.decks' },
  { to: '/leaderboard', key: 'nav.leaderboard' },
  { to: '/spectate', key: 'nav.spectate' },
]

const AUTH_NAV = [
  { to: '/decks', key: 'nav.decks' },
  { to: '/leaderboard', key: 'nav.leaderboard' },
  { to: '/spectate', key: 'nav.spectate' },
  { to: '/operator', key: 'nav.operator' },
  { to: '/friends', key: 'nav.friends' },
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
  const { status: wsStatus } = useLiveSocket()
  const navigate = useNavigate()
  const location = useLocation()
  const nav = user ? AUTH_NAV : PUBLIC_NAV
  const [theme, toggleTheme] = useTheme()
  const { t, locale, setLocale, availableLocales } = useTranslation()

  const isNavActive = (to) => {
    return location.pathname === to.split('?')[0]
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
            <NavLink to={user ? '/dash' : '/'} className="brand" style={{ textDecoration: 'none' }}>{t('app.brand')}</NavLink>
            <nav>
              {nav.map(n => (
                <NavLink
                  key={n.key}
                  to={n.to}
                  className={isNavActive(n.to) ? 'on' : ''}
                >
                  {t(n.key)}
                </NavLink>
              ))}
            </nav>
          </div>
          <div className="appbar-controls">
            <SearchBar />
            <button onClick={toggleTheme}>{theme === 'dark' ? '☽' : '☀'}</button>
            {availableLocales.length > 1 && (
              <select
                value={locale}
                onChange={e => setLocale(e.target.value)}
                aria-label="Change language"
                style={{ outline: 'none' }}
              >
                {availableLocales.map(l => (
                  <option key={l} value={l} style={{ background: 'var(--bg)', color: 'var(--ink)' }}>
                    {l.toUpperCase()}
                  </option>
                ))}
              </select>
            )}
            {!loading && (
              user ? (
                <a onClick={handleLogout}>LOGOUT</a>
              ) : (
                <NavLink to="/login" style={{ color: 'var(--accent)', textDecoration: 'none', fontWeight: 700 }}>SIGN IN ↗</NavLink>
              )
            )}
          </div>
        </div>

        <div style={{ flex: 1, overflow: 'auto', overflowX: 'hidden', display: 'flex', flexDirection: 'column' }}>
          <Outlet />
        </div>

        <div className="statusbar" style={{ flexDirection: 'column', gap: 0, padding: 0 }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '4px 14px', borderBottom: '1px solid var(--rule)' }}>
            <span>
              <span style={{ color: wsStatus === 'live' ? 'var(--ok)' : wsStatus === 'disconnected' ? 'var(--danger)' : 'var(--warn)' }}>●</span>
              {' '}
              {wsStatus === 'live' && '+ + +  HEXDEK CORE READY  + + +'}
              {wsStatus === 'contacting' && 'CONTACTING SERVER...'}
              {wsStatus === 'initializing' && 'INITIALIZING...'}
              {wsStatus === 'disconnected' && 'DISCONNECTED — RECONNECTING...'}
            </span>
            <span>OPEN SOURCE / / DONATIONS-POWERED / / NO ADS</span>
            <span>{user ? `USR.${user.email?.split('@')[0]?.toUpperCase()}` : 'GUEST'}</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 14, padding: '4px 14px' }}>
            <NavLink to="/about" style={{ color: 'var(--ink-2)', textDecoration: 'none', fontSize: 9, letterSpacing: '0.08em', fontWeight: 700 }}>ABOUT</NavLink>
            <NavLink to="/feedback" style={{ color: 'var(--danger)', textDecoration: 'none', fontSize: 9, letterSpacing: '0.08em', fontWeight: 700 }}>BUG / SUGGESTION</NavLink>
            <NavLink to="/donations" style={{ color: 'var(--ok)', textDecoration: 'none', fontSize: 9, letterSpacing: '0.08em', fontWeight: 700 }}>DONATE ♥</NavLink>
            <a href="https://discord.gg/Mz2ueRFXds" target="_blank" rel="noopener noreferrer" style={{ color: 'var(--ink-2)', textDecoration: 'none', fontSize: 9, letterSpacing: '0.08em', fontWeight: 700 }}>DISCORD</a>
          </div>
        </div>
      </div>
      <ToastHost />
    </div>
  )
}
