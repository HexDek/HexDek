import { useEffect, useRef } from 'react'

// Shared modal keyboard plumbing:
//   - closes on Escape
//   - traps Tab/Shift+Tab inside the modal panel
//   - restores focus to the element that was active before the modal opened
//
// Usage:
//   const panelRef = useModalKeyboard({ onClose })
//   return <div ref={panelRef} role="dialog" aria-modal="true">...</div>
//
// Skip the trap (e.g. when an inner widget owns its own keyboard) by
// passing `trapFocus: false`. Escape and focus-restore still apply.

const FOCUSABLE_SELECTOR = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled]):not([type="hidden"])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(',')

export function useModalKeyboard({ onClose, trapFocus = true } = {}) {
  const panelRef = useRef(null)
  const prevActiveRef = useRef(null)

  useEffect(() => {
    prevActiveRef.current = document.activeElement
    return () => {
      const prev = prevActiveRef.current
      if (prev && typeof prev.focus === 'function' && document.contains(prev)) {
        prev.focus()
      }
    }
  }, [])

  useEffect(() => {
    const onKey = (e) => {
      if (e.key === 'Escape') {
        e.stopPropagation()
        onClose?.()
        return
      }
      if (!trapFocus || e.key !== 'Tab') return
      const panel = panelRef.current
      if (!panel) return
      const focusable = panel.querySelectorAll(FOCUSABLE_SELECTOR)
      if (focusable.length === 0) {
        e.preventDefault()
        return
      }
      const first = focusable[0]
      const last = focusable[focusable.length - 1]
      const active = document.activeElement
      if (e.shiftKey) {
        if (active === first || !panel.contains(active)) {
          e.preventDefault()
          last.focus()
        }
      } else {
        if (active === last) {
          e.preventDefault()
          first.focus()
        }
      }
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [onClose, trapFocus])

  return panelRef
}
