import { useEffect, useState } from 'react'

// useArtContrast — given a card-art URL, compute whether the top of the
// image is dark or light, so overlaid text can flip color/shadow to stay
// readable. Returns 'light' | 'dark' | null (null = unknown: still
// loading, or the image was cross-origin without CORS headers and we
// couldn't read pixels).
//
// Sampling strategy:
//   1. Draw the image's top 30% rect into a small offscreen canvas
//      (64×N, so we don't decode the full image more than once).
//   2. Average luminance per Rec. 601: (0.299R + 0.587G + 0.114B) / 255.
//   3. Threshold at 0.5 — any value at or above is 'light', below is
//      'dark'. The threshold is intentionally high so darkly-tinted art
//      with bright highlights still classifies as 'dark' (most card
//      illustrations sit well below 0.5).
//
// Results are memoised in sessionStorage so navigating back to the same
// card / deck page doesn't redecode and re-sample every time.

const MEMORY_CACHE = new Map() // url -> 'light' | 'dark' | 'unknown'
const STORAGE_PREFIX = 'art-contrast:v1:'

function readCache(url) {
  if (MEMORY_CACHE.has(url)) return MEMORY_CACHE.get(url)
  try {
    const v = sessionStorage.getItem(STORAGE_PREFIX + url)
    if (v === 'light' || v === 'dark' || v === 'unknown') {
      MEMORY_CACHE.set(url, v)
      return v
    }
  } catch {}
  return null
}

function writeCache(url, value) {
  MEMORY_CACHE.set(url, value)
  try { sessionStorage.setItem(STORAGE_PREFIX + url, value) } catch {}
}

function classify(url) {
  return new Promise((resolve) => {
    const img = new Image()
    // Required so the canvas read isn't tainted. The /api/card-art proxy
    // forwards CORS-friendly Scryfall image responses; if a future call
    // points at a host without permissive headers, getImageData throws
    // SecurityError below and we resolve to 'unknown'.
    img.crossOrigin = 'anonymous'
    img.onload = () => {
      try {
        const w = 64
        const h = Math.max(8, Math.round((img.naturalHeight / img.naturalWidth) * w * 0.3))
        const canvas = document.createElement('canvas')
        canvas.width = w
        canvas.height = h
        const ctx = canvas.getContext('2d', { willReadFrequently: true })
        // Draw only the top 30% of the source image into the canvas.
        const srcH = Math.max(1, Math.round(img.naturalHeight * 0.3))
        ctx.drawImage(img, 0, 0, img.naturalWidth, srcH, 0, 0, w, h)
        const data = ctx.getImageData(0, 0, w, h).data
        let sum = 0
        let n = 0
        for (let i = 0; i < data.length; i += 4) {
          const a = data[i + 3]
          if (a < 16) continue // skip near-transparent pixels
          const r = data[i]
          const g = data[i + 1]
          const b = data[i + 2]
          sum += 0.299 * r + 0.587 * g + 0.114 * b
          n++
        }
        if (n === 0) { resolve('unknown'); return }
        const lum = sum / n / 255
        resolve(lum >= 0.5 ? 'light' : 'dark')
      } catch {
        resolve('unknown')
      }
    }
    img.onerror = () => resolve('unknown')
    img.src = url
  })
}

export function useArtContrast(imageUrl) {
  const [contrast, setContrast] = useState(() => {
    if (!imageUrl) return null
    const cached = readCache(imageUrl)
    if (cached === 'light' || cached === 'dark') return cached
    return null
  })

  useEffect(() => {
    if (!imageUrl) { setContrast(null); return }
    const cached = readCache(imageUrl)
    if (cached === 'light' || cached === 'dark') {
      setContrast(cached)
      return
    }
    if (cached === 'unknown') {
      setContrast(null)
      return
    }
    let alive = true
    classify(imageUrl).then(result => {
      writeCache(imageUrl, result)
      if (!alive) return
      setContrast(result === 'unknown' ? null : result)
    })
    return () => { alive = false }
  }, [imageUrl])

  return contrast
}

export default useArtContrast
