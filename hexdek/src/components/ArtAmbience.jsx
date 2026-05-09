import { useEffect, useRef, useState } from 'react'
import { cardArtUrl } from '../services/api'

// Crossfade backdrop for the spectator views. The previous backdrop
// stays mounted underneath while the new one fades in on top, so the
// blurred color-identity bleed morphs into the next commander instead
// of snapping. Fade duration is kept in lockstep with the
// .art-ambience-fade-in keyframes in index.css.
const FADE_MS = 1200

export default function ArtAmbience({ name }) {
  const [layers, setLayers] = useState(() => name ? [{ name, id: 0 }] : [])
  const counterRef = useRef(0)

  useEffect(() => {
    if (!name) return
    setLayers(prev => {
      if (prev.length && prev[prev.length - 1].name === name) return prev
      counterRef.current += 1
      return [...prev.slice(-1), { name, id: counterRef.current }]
    })
  }, [name])

  useEffect(() => {
    if (layers.length <= 1) return
    const t = setTimeout(() => setLayers(prev => prev.slice(-1)), FADE_MS + 50)
    return () => clearTimeout(t)
  }, [layers])

  if (!layers.length) return null
  return layers.map((l, i) => {
    const url = cardArtUrl(l.name)
    if (!url) return null
    const isTop = i === layers.length - 1
    const isFadingIn = isTop && layers.length > 1
    return (
      <img
        key={l.id}
        className={`art-ambience${isFadingIn ? ' art-ambience-fade-in' : ''}`}
        src={url}
        alt=""
        aria-hidden="true"
      />
    )
  })
}
