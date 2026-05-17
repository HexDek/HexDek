import { useState, useId, useRef, useEffect } from 'react'
import { getGlossaryTerm } from '../data/glossary'

// GlossaryTerm — wraps any label or stat with a tap-to-expand inline
// explanation. Pass `term` (an id from src/data/glossary.js) and the
// rendered label as children; the component adds a subtle dotted
// underline plus a tiny ? indicator and toggles a disclosure panel
// directly below on click/tap. Designed to be mobile-first: tap once
// to expand, tap again (or anywhere else) to collapse. No hover-only
// state.
//
// If `term` resolves to no glossary entry, the component degrades to
// rendering the children unchanged so it's safe to sprinkle freely
// while the glossary is still being filled in.
//
// Props:
//   term      — glossary id (string)
//   children  — the visible label / stat
//   inline    — render the disclosure inline (default) vs. as a popover
//   compact   — smaller indicator, used inside table cells
//   onOpen    — optional callback fired when expanded (analytics hook)

export default function GlossaryTerm({
  term,
  children,
  inline = true,
  compact = false,
  onOpen,
}) {
  const entry = getGlossaryTerm(term)
  const [open, setOpen] = useState(false)
  // flip = anchor panel to the right edge of the trigger instead of the
  // left edge. Set when the trigger sits in the right half of the viewport
  // so a 320px-wide panel doesn't overflow the right margin.
  const [flip, setFlip] = useState(false)
  const wrapRef = useRef(null)
  const panelId = useId()

  useEffect(() => {
    if (!open) return
    const onDocClick = (e) => {
      if (wrapRef.current && !wrapRef.current.contains(e.target)) {
        setOpen(false)
      }
    }
    const onKey = (e) => {
      if (e.key === 'Escape') setOpen(false)
    }
    document.addEventListener('mousedown', onDocClick)
    document.addEventListener('touchstart', onDocClick)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('mousedown', onDocClick)
      document.removeEventListener('touchstart', onDocClick)
      document.removeEventListener('keydown', onKey)
    }
  }, [open])

  if (!entry) return <>{children}</>

  const toggle = (e) => {
    e.stopPropagation()
    e.preventDefault()
    const next = !open
    if (next && wrapRef.current) {
      // Compute panel anchor side BEFORE opening. If the trigger sits past
      // the viewport midpoint, flip the anchor to the right edge so the
      // 320px panel grows leftward into the viewport.
      const rect = wrapRef.current.getBoundingClientRect()
      const vw = window.innerWidth || document.documentElement.clientWidth || 0
      setFlip(rect.left + 320 > vw - 12)
    }
    setOpen(next)
    if (next && onOpen) onOpen(term)
  }

  return (
    <span
      ref={wrapRef}
      className={`gloss ${compact ? 'gloss--compact' : ''} ${open ? 'gloss--open' : ''}`}
    >
      <button
        type="button"
        className="gloss-trigger"
        aria-expanded={open}
        aria-controls={panelId}
        aria-label={`Define ${entry.title}`}
        onClick={toggle}
        onTouchEnd={(e) => {
          // Some mobile browsers fire both touchend + click; prevent
          // the double-toggle by handling once and stopping the click.
          e.preventDefault()
          toggle(e)
        }}
      >
        <span className="gloss-label">{children}</span>
        <span className="gloss-mark" aria-hidden="true">?</span>
      </button>
      {open && (
        <span
          id={panelId}
          role="region"
          aria-label={`${entry.title} explanation`}
          className={`gloss-panel ${inline ? 'gloss-panel--inline' : 'gloss-panel--popover'} ${flip ? 'gloss-panel--flip' : ''}`}
        >
          <span className="gloss-panel-title">{entry.title}</span>
          <span className="gloss-panel-body">{entry.body}</span>
          {entry.source && (
            <span className="gloss-panel-source">{entry.source}</span>
          )}
        </span>
      )}
    </span>
  )
}
