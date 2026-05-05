import { createContext, useContext, useState, useEffect, useRef } from 'react'
import { onAuthChange, signOutUser } from '../lib/firebase'
import { stitchSession } from '../hooks/useAnalytics'

const AuthContext = createContext(null)

function ownerSlug(u) {
  if (!u) return ''
  if (typeof window !== 'undefined') {
    try {
      const stored = window.localStorage.getItem('hexdek_owner')
      if (stored) return stored.toLowerCase()
    } catch { /* ignore */ }
  }
  const slug = u.displayName?.toLowerCase()
    || u.email?.split('@')[0]?.split('.')[0]
    || ''
  return slug.toLowerCase()
}

const DEV_USER = {
  uid: 'dev-local',
  email: 'dev@localhost',
  displayName: 'DEV OPERATOR',
}

const isLocalhost = typeof window !== 'undefined' &&
  (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')

export function AuthProvider({ children }) {
  const [user, setUser] = useState(isLocalhost ? DEV_USER : null)
  const [loading, setLoading] = useState(!isLocalhost)
  const stitchedFor = useRef(null)

  useEffect(() => {
    if (isLocalhost) return
    const unsub = onAuthChange((u) => {
      setUser(u)
      setLoading(false)
    })
    return unsub
  }, [])

  // Temporal Pincer — stitch the anonymous browser id to the owner the
  // first time auth resolves (and again if the owner changes). Idempotent
  // on the server side via INSERT OR REPLACE on (anon_id, owner).
  useEffect(() => {
    const slug = ownerSlug(user)
    if (!slug) {
      stitchedFor.current = null
      return
    }
    if (stitchedFor.current === slug) return
    stitchedFor.current = slug
    stitchSession(slug)
  }, [user])

  const logout = async () => {
    await signOutUser()
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, loading, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be inside AuthProvider')
  return ctx
}
