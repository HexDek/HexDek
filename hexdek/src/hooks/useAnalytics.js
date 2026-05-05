import { useEffect } from 'react'
import { useLocation } from 'react-router-dom'
import { getAnonId } from '../lib/anonId'
import { API_BASE } from '../services/api'

function gtag() {
  if (window.gtag) window.gtag(...arguments)
}

// Best-effort — telemetry must never block the UI or surface errors.
function postTelemetry(path, body) {
  try {
    fetch(`${API_BASE}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      keepalive: true,
    }).catch(() => {})
  } catch { /* ignore */ }
}

export function usePageTracking() {
  const location = useLocation()
  useEffect(() => {
    const fullPath = location.pathname + location.search
    gtag('event', 'page_view', {
      page_path: fullPath,
      page_title: document.title,
    })
    const anonId = getAnonId()
    if (!anonId) return
    postTelemetry('/api/telemetry/pageview', {
      anon_id:   anonId,
      path:      fullPath,
      timestamp: Date.now(),
      referrer:  typeof document !== 'undefined' ? document.referrer || '' : '',
    })
  }, [location])
}

// stitchSession links the current anon_id to an authenticated owner.
// Returns the fetch promise (callers may discard it; failures are silent).
export function stitchSession(owner) {
  const anonId = getAnonId()
  if (!anonId || !owner) return Promise.resolve(null)
  return fetch(`${API_BASE}/api/telemetry/stitch`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ anon_id: anonId, owner }),
    keepalive: true,
  }).catch(() => null)
}

export function trackEvent(action, params = {}) {
  gtag('event', action, params)
}
