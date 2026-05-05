import { useState, useEffect, useRef, useCallback } from 'react'
import { createPortal } from 'react-dom'
import { useNavigate } from 'react-router-dom'
import { cardImageUrl } from '../services/api'

const MTG_COLORS = {
  W: 'rgba(248, 231, 185, 0.35)',
  U: 'rgba(14, 104, 171, 0.35)',
  B: 'rgba(21, 11, 0, 0.45)',
  R: 'rgba(211, 32, 42, 0.35)',
  G: 'rgba(0, 115, 62, 0.35)',
}

function colorTint(colors) {
  if (!colors || colors.length === 0) return 'rgba(30, 30, 30, 0.85)'
  if (colors.length === 1) return MTG_COLORS[colors[0]] || 'rgba(30, 30, 30, 0.85)'
  const c1 = MTG_COLORS[colors[0]] || 'rgba(30, 30, 30, 0.85)'
  const c2 = MTG_COLORS[colors[1]] || 'rgba(30, 30, 30, 0.85)'
  return `linear-gradient(135deg, ${c1}, ${c2})`
}

const cache = new Map()

function fetchCard(name) {
  const key = (name || '').toLowerCase()
  if (!key) return Promise.resolve(null)
  if (cache.has(key)) return cache.get(key)
  const url = `https://api.scryfall.com/cards/named?exact=${encodeURIComponent(name)}`
  const p = fetch(url)
    .then(r => r.ok ? r.json() : null)
    .then(d => {
      if (!d) return null
      return {
        name: d.name,
        colors: d.colors || d.color_identity || [],
        scryfall_uri: d.scryfall_uri,
      }
    })
    .catch(() => null)
  cache.set(key, p)
  return p
}

function CardPopupBody({ name, onClose, onNavigate }) {
  const [data, setData] = useState(null)
  const imgUrl = cardImageUrl(name)
  const cardRef = useRef(null)

  useEffect(() => {
    let alive = true
    fetchCard(name).then(d => { if (alive) setData(d) })
    return () => { alive = false }
  }, [name])

  useEffect(() => {
    const esc = (e) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', esc)
    return () => document.removeEventListener('keydown', esc)
  }, [onClose])

  const tint = data ? colorTint(data.colors) : 'rgba(30, 30, 30, 0.85)'
  const bgStyle = tint.startsWith('linear')
    ? { backgroundImage: `${tint}, linear-gradient(rgba(0,0,0,0.7), rgba(0,0,0,0.7))` }
    : { backgroundColor: tint }

  return createPortal(
    <div
      style={{
        position: 'fixed', inset: 0, zIndex: 1100,
        display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
        backdropFilter: 'blur(6px)',
        ...bgStyle,
      }}
      onClick={onClose}
    >
      <div
        ref={cardRef}
        onClick={(e) => e.stopPropagation()}
        style={{
          display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 12,
          maxWidth: 'min(90vw, 360px)', maxHeight: '85vh',
        }}
      >
        {imgUrl && (
          <img
            src={imgUrl}
            alt={name}
            style={{
              width: '100%', maxHeight: '70vh', objectFit: 'contain',
              borderRadius: 12, boxShadow: '0 8px 32px rgba(0,0,0,0.6)',
            }}
            onError={(e) => { e.target.style.display = 'none' }}
          />
        )}
        <button
          type="button"
          onClick={(e) => { e.stopPropagation(); onNavigate() }}
          style={{
            background: 'var(--inv-bg)', color: 'var(--inv-ink)',
            border: '1px solid var(--ink)', padding: '8px 16px',
            fontFamily: 'inherit', fontSize: 11, fontWeight: 700,
            letterSpacing: '0.1em', cursor: 'pointer', textTransform: 'uppercase',
          }}
        >
          VIEW CARD PAGE →
        </button>
      </div>
    </div>,
    document.body,
  )
}

export function useCardPopup(name) {
  const [open, setOpen] = useState(false)
  const triggerRef = useRef(null)
  const navigate = useNavigate()

  const openPopup = useCallback((e) => {
    if (e) { e.preventDefault(); e.stopPropagation() }
    setOpen(true)
  }, [])

  const close = useCallback(() => {
    setOpen(false)
  }, [])

  const goToCardPage = useCallback(() => {
    close()
    navigate(`/cards/${encodeURIComponent(name)}`)
  }, [close, name, navigate])

  const triggerProps = {
    ref: triggerRef,
    onClick: openPopup,
    onTouchStart: openPopup,
  }

  const popup = open && name ? (
    <CardPopupBody
      name={name}
      onClose={close}
      onNavigate={goToCardPage}
    />
  ) : null

  return { triggerProps, popup, isOpen: open }
}

export default function CardPopupTrigger({ name, children, as: As = 'span', style, ...rest }) {
  const { triggerProps, popup } = useCardPopup(name)
  return (
    <>
      <As {...rest} {...triggerProps} style={{ cursor: 'pointer', ...style }}>
        {children}
      </As>
      {popup}
    </>
  )
}
