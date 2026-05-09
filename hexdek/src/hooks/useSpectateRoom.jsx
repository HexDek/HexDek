import { useEffect, useRef, useState, useCallback } from 'react'

const API_BASE = import.meta.env.VITE_API_URL ?? ''

function wsUrl(roomId) {
  return API_BASE.replace(/^http/, 'ws') + `/ws/spectate/${roomId}`
}

// Exponential backoff matches useLiveSocket so both reconnect banners
// behave consistently. After MAX_ATTEMPTS the loop halts and waits for
// a manual reconnectNow() call from the banner.
const RECONNECT_BASE_MS = 1000
const RECONNECT_MAX_MS = 30000
const MAX_ATTEMPTS = 10

function backoffFor(attempt) {
  const exp = RECONNECT_BASE_MS * Math.pow(2, Math.max(0, attempt - 1))
  return Math.min(RECONNECT_MAX_MS, exp)
}

export function useSpectateRoom(roomId) {
  const [game, setGame] = useState(null)
  const [elo, setElo] = useState([])
  const [speed, setSpeed] = useState(1)
  const [viewers, setViewers] = useState(0)
  const [roomInfo, setRoomInfo] = useState(null)
  const [status, setStatus] = useState('disconnected')
  const [reconnectAttempt, setReconnectAttempt] = useState(0)
  const [nextRetryAt, setNextRetryAt] = useState(0)
  const wsRef = useRef(null)
  const reconnectRef = useRef(null)
  const attemptRef = useRef(0)
  const cancelledRef = useRef(false)
  const connectRef = useRef(null)

  const sendSpeed = useCallback((multiplier) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'speed', payload: { multiplier } }))
    }
  }, [])

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
    if (connectRef.current) connectRef.current()
  }, [])

  useEffect(() => {
    if (!roomId) return

    cancelledRef.current = false

    function connect() {
      if (cancelledRef.current) return
      if (wsRef.current?.readyState === WebSocket.OPEN) return
      if (reconnectRef.current) {
        clearTimeout(reconnectRef.current)
        reconnectRef.current = null
      }

      setStatus('contacting')
      setNextRetryAt(0)
      const ws = new WebSocket(wsUrl(roomId))
      wsRef.current = ws

      ws.onopen = () => {
        setStatus('live')
        attemptRef.current = 0
        setReconnectAttempt(0)
        setNextRetryAt(0)
      }

      ws.onmessage = (evt) => {
        try {
          const { type, payload } = JSON.parse(evt.data)
          switch (type) {
            case 'game': setGame(payload); break
            case 'elo': setElo(payload || []); break
            case 'speed': setSpeed(payload?.multiplier ?? 1); break
            case 'viewers': setViewers(payload?.count ?? 0); break
            case 'room_info':
              setRoomInfo(payload)
              setViewers(payload?.viewers ?? 0)
              setSpeed(payload?.speed ?? 1)
              break
            case 'pong': break
          }
        } catch {}
      }

      ws.onclose = () => {
        if (cancelledRef.current) return
        wsRef.current = null
        attemptRef.current += 1
        setReconnectAttempt(attemptRef.current)
        if (attemptRef.current > MAX_ATTEMPTS) {
          setStatus('failed')
          setNextRetryAt(0)
          return
        }
        setStatus('disconnected')
        const delay = backoffFor(attemptRef.current)
        setNextRetryAt(Date.now() + delay)
        reconnectRef.current = setTimeout(connect, delay)
      }

      ws.onerror = () => ws.close()
    }

    connectRef.current = connect
    connect()

    const ping = setInterval(() => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ type: 'ping' }))
      }
    }, 30000)

    return () => {
      cancelledRef.current = true
      clearInterval(ping)
      if (reconnectRef.current) clearTimeout(reconnectRef.current)
      if (wsRef.current) wsRef.current.close()
      connectRef.current = null
    }
  }, [roomId])

  return {
    game, elo, speed, viewers, roomInfo, status,
    reconnectAttempt, nextRetryAt, maxAttempts: MAX_ATTEMPTS,
    sendSpeed, reconnectNow,
  }
}
