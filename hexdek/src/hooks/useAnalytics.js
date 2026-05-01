import { useEffect } from 'react'
import { useLocation } from 'react-router-dom'

function gtag() {
  if (window.gtag) window.gtag(...arguments)
}

export function usePageTracking() {
  const location = useLocation()
  useEffect(() => {
    gtag('event', 'page_view', {
      page_path: location.pathname + location.search,
      page_title: document.title,
    })
  }, [location])
}

export function trackEvent(action, params = {}) {
  gtag('event', action, params)
}
