import { createContext, useContext, useEffect, useRef, useState, useCallback } from 'react'

const API_BASE = import.meta.env.VITE_API_URL ?? ''
const WS_URL = API_BASE.replace(/^http/, 'ws') + '/ws/live'
const CACHE_KEY = 'hexdek_live_cache'

function loadCached() {
  try {
    const raw = sessionStorage.getItem(CACHE_KEY)
    if (!raw) return null
    const c = JSON.parse(raw)
    if (c.stats && c.ts && c.stats.games_per_min > 0) {
      const elapsed = (Date.now() - c.ts) / 1000
      if (elapsed < 3600) {
        c.stats = { ...c.stats, games_played: Math.round(c.stats.games_played + (c.stats.games_per_min / 60) * elapsed) }
      }
    }
    return c
  } catch { return null }
}

function saveCache(patch) {
  try {
    const prev = JSON.parse(sessionStorage.getItem(CACHE_KEY) || '{}')
    sessionStorage.setItem(CACHE_KEY, JSON.stringify({ ...prev, ...patch, ts: Date.now() }))
  } catch {}
}

const LiveCtx = createContext(null)

// Exponential backoff: BASE * 2^(attempt-1), capped at MAX. After
// MAX_ATTEMPTS the auto-retry stops and only manual reconnectNow()
// will re-arm the loop.
const RECONNECT_BASE_MS = 1000
const RECONNECT_MAX_MS = 30000
const MAX_ATTEMPTS = 10

function backoffFor(attempt) {
  const exp = RECONNECT_BASE_MS * Math.pow(2, Math.max(0, attempt - 1))
  return Math.min(RECONNECT_MAX_MS, exp)
}

// status: 'disconnected' | 'contacting' | 'initializing' | 'live' | 'failed'
export function LiveProvider({ children }) {
  const cached = useRef(loadCached()).current
  const [game, setGame] = useState(null)
  const [elo, setElo] = useState(cached?.elo || [])
  const [stats, setStats] = useState(cached?.stats || null)
  const [history, setHistory] = useState([])
  const [speed, setSpeed] = useState(1)
  const [status, setStatus] = useState('disconnected')
  const [reconnectAttempt, setReconnectAttempt] = useState(0)
  const [nextRetryAt, setNextRetryAt] = useState(0)
  const wsRef = useRef(null)
  const reconnectRef = useRef(null)
  const gotFirstStats = useRef(!!cached?.stats)
  const attemptRef = useRef(0)

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return
    if (reconnectRef.current) {
      clearTimeout(reconnectRef.current)
      reconnectRef.current = null
    }

    setStatus('contacting')
    setNextRetryAt(0)
    const ws = new WebSocket(WS_URL)
    wsRef.current = ws

    ws.onopen = () => {
      setStatus('initializing')
      attemptRef.current = 0
      setReconnectAttempt(0)
      setNextRetryAt(0)
    }

    ws.onmessage = (evt) => {
      try {
        const { type, payload } = JSON.parse(evt.data)
        switch (type) {
          case 'game': setGame(payload); break
          case 'elo': setElo(payload || []); saveCache({ elo: payload }); break
          case 'stats':
            setStats(payload)
            saveCache({ stats: payload })
            if (!gotFirstStats.current) {
              gotFirstStats.current = true
              setStatus('live')
            }
            break
          case 'history': setHistory(payload || []); break
          case 'speed': setSpeed(payload?.multiplier ?? 1); break
          case 'pong': break
        }
        if (gotFirstStats.current && type !== 'pong') {
          setStatus('live')
        }
      } catch {}
    }

    ws.onclose = () => {
      wsRef.current = null
      gotFirstStats.current = false
      attemptRef.current += 1
      setReconnectAttempt(attemptRef.current)
      if (attemptRef.current > MAX_ATTEMPTS) {
        // Give up auto-retry; user must hit "Reconnect Now". Do NOT
        // schedule another setTimeout — the banner reads `status` to
        // gate the manual button.
        setStatus('failed')
        setNextRetryAt(0)
        return
      }
      setStatus('disconnected')
      const delay = backoffFor(attemptRef.current)
      setNextRetryAt(Date.now() + delay)
      reconnectRef.current = setTimeout(connect, delay)
    }

    ws.onerror = () => {
      ws.close()
    }
  }, [])

  // Manual override — clears any pending backoff timer and reconnects
  // immediately. Resets the attempt counter so the next failure starts
  // backoff from scratch instead of continuing the prior 30s ceiling.
  const reconnectNow = useCallback(() => {
    if (reconnectRef.current) {
      clearTimeout(reconnectRef.current)
      reconnectRef.current = null
    }
    if (wsRef.current) {
      try { wsRef.current.close() } catch {}
      wsRef.current = null
    }
    attemptRef.current = 0
    setReconnectAttempt(0)
    setNextRetryAt(0)
    connect()
  }, [connect])

  useEffect(() => {
    connect()

    const ping = setInterval(() => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ type: 'ping' }))
      }
    }, 30000)

    return () => {
      clearInterval(ping)
      if (reconnectRef.current) clearTimeout(reconnectRef.current)
      if (wsRef.current) wsRef.current.close()
    }
  }, [connect])

  return (
    <LiveCtx.Provider value={{
      game, elo, stats, history, speed, status,
      reconnectAttempt, nextRetryAt, maxAttempts: MAX_ATTEMPTS,
      reconnectNow,
    }}>
      {children}
    </LiveCtx.Provider>
  )
}

export function useLiveSocket() {
  const ctx = useContext(LiveCtx)
  if (!ctx) throw new Error('useLiveSocket must be inside LiveProvider')
  return ctx
}
