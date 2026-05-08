import { useEffect, useRef, useState, useCallback } from 'react'

const API_BASE = import.meta.env.VITE_API_URL ?? ''

function wsUrl(roomId) {
  return API_BASE.replace(/^http/, 'ws') + `/ws/spectate/${roomId}`
}

const RECONNECT_DELAY_MS = 2000

export function useSpectateRoom(roomId) {
  const [game, setGame] = useState(null)
  const [elo, setElo] = useState([])
  const [speed, setSpeed] = useState(1)
  const [viewers, setViewers] = useState(0)
  const [roomInfo, setRoomInfo] = useState(null)
  const [status, setStatus] = useState('disconnected')
  const wsRef = useRef(null)
  const reconnectRef = useRef(null)

  const sendSpeed = useCallback((multiplier) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'speed', payload: { multiplier } }))
    }
  }, [])

  useEffect(() => {
    if (!roomId) return

    let cancelled = false

    function connect() {
      if (cancelled) return
      if (wsRef.current?.readyState === WebSocket.OPEN) return

      setStatus('contacting')
      const ws = new WebSocket(wsUrl(roomId))
      wsRef.current = ws

      ws.onopen = () => {
        setStatus('live')
        if (reconnectRef.current) {
          clearTimeout(reconnectRef.current)
          reconnectRef.current = null
        }
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
        if (cancelled) return
        setStatus('disconnected')
        wsRef.current = null
        reconnectRef.current = setTimeout(connect, RECONNECT_DELAY_MS)
      }

      ws.onerror = () => ws.close()
    }

    connect()

    const ping = setInterval(() => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ type: 'ping' }))
      }
    }, 30000)

    return () => {
      cancelled = true
      clearInterval(ping)
      if (reconnectRef.current) clearTimeout(reconnectRef.current)
      if (wsRef.current) wsRef.current.close()
    }
  }, [roomId])

  return { game, elo, speed, viewers, roomInfo, status, sendSpeed }
}
