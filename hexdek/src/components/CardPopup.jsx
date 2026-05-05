import { useState, useEffect, useRef, useCallback } from 'react'
import { createPortal } from 'react-dom'
import { useNavigate } from 'react-router-dom'
import { cardArtUrl } from '../services/api'

// CardPopup — hover (desktop) / tap (mobile) preview for any card name.
//
// Two integration points:
//   <CardPopupTrigger name="Sol Ring">…</CardPopupTrigger>   — wraps children
//   const { triggerProps, popup } = useCardPopup('Sol Ring') — props spread
//
// Card data (oracle text / type / mana cost) is pulled from Scryfall's
// JSON endpoint and memoized in a module-level Map so repeated hovers of
// the same card share a single network request. Card art uses the
// existing /api/card-art/{name} proxy (cached server-side, theme-aware
// fallback) rather than hitting Scryfall directly.
//
// Touch detection is per-event (no userAgent sniff): a touchstart on the
// trigger latches the popup open until a tap outside dismisses it. Mouse
// hover behaves normally.

const cache = new Map() // name(lowercased) -> Promise<CardData|null>

function fetchCard(name) {
  const key = (name || '').toLowerCase()
  if (!key) return Promise.resolve(null)
  if (cache.has(key)) return cache.get(key)
  const url = `https://api.scryfall.com/cards/named?exact=${encodeURIComponent(name)}`
  const p = fetch(url)
    .then(r => r.ok ? r.json() : null)
    .then(d => {
      if (!d) return null
      const face = (Array.isArray(d.card_faces) && d.card_faces.length) ? d.card_faces[0] : d
      return {
        name: d.name,
        type_line: d.type_line || face.type_line || '',
        mana_cost: d.mana_cost || face.mana_cost || '',
        oracle_text: d.oracle_text || face.oracle_text || '',
        cmc: d.cmc,
      }
    })
    .catch(() => null)
  cache.set(key, p)
  return p
}

const POPUP_W = 280
const POPUP_H_MAX = 460
const VIEWPORT_PAD = 12

// effectiveWidth — the popup is normally POPUP_W, but on narrow viewports
// (mobile portrait) we shrink to the available width minus padding so the
// popup never overflows the screen. Returns both the width and the
// computed position.
function clampPosition(rect) {
  const vw = window.innerWidth
  const vh = window.innerHeight
  const effW = Math.min(POPUP_W, vw - VIEWPORT_PAD * 2)
  // Below mobile breakpoint, anchor to the bottom of the viewport so it
  // doesn't get pushed off-screen by a trigger near the top edge of a
  // long page; this matches the iOS/Android quick-look pattern.
  if (vw < 480) {
    const left = Math.max(VIEWPORT_PAD, Math.round((vw - effW) / 2))
    const maxH = Math.min(POPUP_H_MAX, vh - VIEWPORT_PAD * 2)
    const top = Math.max(VIEWPORT_PAD, vh - maxH - VIEWPORT_PAD)
    return { top, left, width: effW }
  }
  // Desktop: place to the right of the trigger by default, fall back to left.
  let left = rect.right + 8
  if (left + effW > vw - VIEWPORT_PAD) {
    left = rect.left - effW - 8
  }
  if (left < VIEWPORT_PAD) {
    left = Math.max(VIEWPORT_PAD, vw - effW - VIEWPORT_PAD)
  }
  let top = rect.top
  if (top + POPUP_H_MAX > vh - VIEWPORT_PAD) {
    top = Math.max(VIEWPORT_PAD, vh - POPUP_H_MAX - VIEWPORT_PAD)
  }
  if (top < VIEWPORT_PAD) top = VIEWPORT_PAD
  return { top, left, width: effW }
}

function CardPopupBody({ name, position, onClose, onNavigate }) {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const popupRef = useRef(null)

  useEffect(() => {
    let alive = true
    setLoading(true)
    fetchCard(name).then(d => {
      if (!alive) return
      setData(d)
      setLoading(false)
    })
    return () => { alive = false }
  }, [name])

  // Tap-outside dismiss for the latched (touch) state. Mouse hover uses
  // its own onMouseLeave on the trigger so this is a no-op there.
  useEffect(() => {
    const handler = (e) => {
      if (!popupRef.current) return
      if (popupRef.current.contains(e.target)) return
      onClose()
    }
    document.addEventListener('mousedown', handler)
    document.addEventListener('touchstart', handler, { passive: true })
    return () => {
      document.removeEventListener('mousedown', handler)
      document.removeEventListener('touchstart', handler)
    }
  }, [onClose])

  return createPortal(
    <div
      ref={popupRef}
      role="tooltip"
      style={{
        position: 'fixed',
        top: position.top,
        left: position.left,
        width: position.width || POPUP_W,
        maxWidth: 'calc(100vw - 24px)',
        maxHeight: 'min(70vh, 460px)',
        background: 'var(--bg)',
        border: '1px solid var(--ink)',
        boxShadow: '4px 4px 0 0 var(--rule-2)',
        zIndex: 1100,
        display: 'flex',
        flexDirection: 'column',
        fontSize: 11,
      }}
    >
      <div
        className="hatch"
        style={{ aspectRatio: '5/4', borderBottom: '1px solid var(--rule-2)', overflow: 'hidden', position: 'relative' }}
      >
        <img
          src={cardArtUrl(name)}
          alt={name}
          loading="lazy"
          style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }}
          onError={(e) => { e.target.style.display = 'none' }}
        />
      </div>
      <div style={{ padding: '8px 10px', display: 'flex', flexDirection: 'column', gap: 6, overflow: 'auto' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', gap: 8 }}>
          <span style={{ fontSize: 13, fontWeight: 700, lineHeight: 1.2, letterSpacing: '0.02em' }}>
            {data?.name || name}
          </span>
          {data?.mana_cost && (
            <span
              style={{
                fontFamily: 'inherit',
                fontSize: 10,
                letterSpacing: '0.04em',
                border: '1px solid var(--rule-2)',
                padding: '1px 5px',
                whiteSpace: 'nowrap',
              }}
            >
              {data.mana_cost}
            </span>
          )}
        </div>
        {loading ? (
          <div className="t-xs muted">LOADING…</div>
        ) : data ? (
          <>
            {data.type_line && (
              <div className="t-xs" style={{ color: 'var(--ink-2)', textTransform: 'uppercase', letterSpacing: '0.06em' }}>
                {data.type_line}
              </div>
            )}
            {data.oracle_text && (
              <div
                style={{
                  fontSize: 11,
                  lineHeight: 1.4,
                  color: 'var(--ink)',
                  whiteSpace: 'pre-wrap',
                  borderTop: '1px dashed var(--rule)',
                  paddingTop: 6,
                }}
              >
                {data.oracle_text}
              </div>
            )}
          </>
        ) : (
          <div className="t-xs" style={{ color: 'var(--danger)' }}>NOT FOUND ON SCRYFALL</div>
        )}
        <button
          type="button"
          onMouseDown={(e) => { e.preventDefault(); e.stopPropagation() }}
          onClick={(e) => {
            e.stopPropagation()
            onNavigate()
          }}
          style={{
            marginTop: 4,
            background: 'var(--inv-bg)',
            color: 'var(--inv-ink)',
            border: '1px solid var(--ink)',
            padding: '5px 8px',
            fontFamily: 'inherit',
            fontSize: 10,
            fontWeight: 700,
            letterSpacing: '0.1em',
            textAlign: 'left',
            cursor: 'pointer',
          }}
        >
          VIEW CARD PAGE ↗
        </button>
      </div>
    </div>,
    document.body,
  )
}

export function useCardPopup(name) {
  const [open, setOpen] = useState(false)
  const [latched, setLatched] = useState(false) // true after touch tap
  const [position, setPosition] = useState({ top: 0, left: 0, width: POPUP_W })
  const triggerRef = useRef(null)
  const navigate = useNavigate()

  const place = useCallback(() => {
    if (!triggerRef.current) return
    setPosition(clampPosition(triggerRef.current.getBoundingClientRect()))
  }, [])

  const onMouseEnter = useCallback(() => {
    if (latched) return
    place()
    setOpen(true)
  }, [latched, place])

  const onMouseLeave = useCallback(() => {
    if (latched) return
    setOpen(false)
  }, [latched])

  const onTouchStart = useCallback(() => {
    place()
    setLatched(true)
    setOpen(true)
  }, [place])

  const close = useCallback(() => {
    setOpen(false)
    setLatched(false)
  }, [])

  const goToCardPage = useCallback(() => {
    close()
    navigate(`/cards/${encodeURIComponent(name)}`)
  }, [close, name, navigate])

  // Reposition on scroll/resize while open.
  useEffect(() => {
    if (!open) return
    const onUpdate = () => place()
    window.addEventListener('scroll', onUpdate, true)
    window.addEventListener('resize', onUpdate)
    return () => {
      window.removeEventListener('scroll', onUpdate, true)
      window.removeEventListener('resize', onUpdate)
    }
  }, [open, place])

  const triggerProps = {
    ref: triggerRef,
    onMouseEnter,
    onMouseLeave,
    onTouchStart,
  }

  const popup = open && name ? (
    <CardPopupBody
      name={name}
      position={position}
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
      <As {...rest} {...triggerProps} style={{ cursor: 'help', ...style }}>
        {children}
      </As>
      {popup}
    </>
  )
}
