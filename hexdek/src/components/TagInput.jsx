import { useState, useEffect, useRef, useCallback } from 'react'
import { api } from '../services/api'

// TagInput — chip-style tag editor with autocomplete suggestions backed
// by /api/tags. Used by ImportModal (during deck creation) and
// DeckArchive (when editing an existing deck). Mirrors the
// WorkshopAddCard pattern in DeckArchive.jsx: 200ms debounced query,
// keyboard + click to pick, focus-blur dropdown.
//
// Tags are normalized to lowercase / trimmed before being added so the
// chip the user sees matches what the server will store.
//
// Props:
//   value:    string[] — current tags (controlled)
//   onChange: (next: string[]) => void
//   owner:    string   — scope suggestions to this owner; pass '*' for
//                       the global pool, or omit to use the signed-in
//                       owner inferred from the X-HexDek-Owner header.
//   max:      number   — soft cap (default 16, matches MaxTagsPerDeck)
//   placeholder, label: cosmetics
export default function TagInput({
  value = [],
  onChange,
  owner,
  max = 16,
  placeholder = 'add tag — type and press Enter',
  label,
  disabled = false,
}) {
  const [q, setQ] = useState('')
  const [results, setResults] = useState([])
  const [focused, setFocused] = useState(false)
  const inputRef = useRef(null)

  const tags = Array.isArray(value) ? value : []
  const atMax = tags.length >= max

  // Fetch suggestions on debounce. Empty query returns the most-used
  // tags overall (the server sorts by count desc), so the user sees
  // their existing tag vocabulary as soon as they focus the input.
  useEffect(() => {
    if (!focused) return
    let cancelled = false
    const t = setTimeout(() => {
      api.getTagSuggestions({ q: q.trim(), owner, limit: 8 })
        .then(rows => {
          if (cancelled) return
          const list = Array.isArray(rows) ? rows : []
          const filtered = list.filter(r => !tags.includes(r.tag))
          setResults(filtered)
        })
        .catch(() => { if (!cancelled) setResults([]) })
    }, 200)
    return () => { cancelled = true; clearTimeout(t) }
  }, [q, focused, owner, tags])

  const add = useCallback((raw) => {
    const t = String(raw || '').trim().toLowerCase()
    if (!t) return
    if (t.length > 32) return
    if (tags.includes(t)) return
    if (tags.length >= max) return
    onChange?.([...tags, t])
    setQ('')
    setResults([])
  }, [tags, onChange, max])

  const remove = useCallback((t) => {
    onChange?.(tags.filter(x => x !== t))
  }, [tags, onChange])

  const onKeyDown = (e) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault()
      // Pull the top suggestion when it's an exact-prefix match the
      // user is likely targeting; otherwise treat the input as a free
      // new tag (lets the user invent vocabulary the autocomplete
      // hasn't seen yet — the whole point of the field on day one).
      const top = results[0]
      const typed = q.trim().toLowerCase()
      if (top && top.tag.startsWith(typed) && typed.length > 0) {
        add(top.tag)
      } else {
        add(typed)
      }
    } else if (e.key === 'Backspace' && q === '' && tags.length > 0) {
      remove(tags[tags.length - 1])
    } else if (e.key === 'Escape') {
      setQ('')
      setResults([])
      inputRef.current?.blur()
    }
  }

  return (
    <div style={{ position: 'relative' }}>
      {label && (
        <label className="import-modal__label" style={{ display: 'block', marginBottom: 4 }}>
          {label}
        </label>
      )}
      <div
        style={{
          display: 'flex', flexWrap: 'wrap', gap: 4, alignItems: 'center',
          padding: '4px 6px',
          background: 'var(--bg-2, rgba(0,0,0,0.3))',
          border: '1px solid var(--rule-2)',
          minHeight: 30,
          opacity: disabled ? 0.5 : 1,
        }}
        onClick={() => inputRef.current?.focus()}
      >
        {tags.map(t => (
          <span
            key={t}
            style={{
              display: 'inline-flex', alignItems: 'center', gap: 4,
              padding: '2px 6px',
              background: 'var(--panel)',
              border: '1px solid var(--rule)',
              fontSize: 10,
              letterSpacing: '0.04em',
              textTransform: 'uppercase',
            }}
          >
            {t}
            {!disabled && (
              <button
                type="button"
                onClick={(e) => { e.stopPropagation(); remove(t) }}
                aria-label={`remove ${t}`}
                style={{
                  background: 'none', border: 'none',
                  color: 'var(--muted-2, #888)', cursor: 'pointer',
                  padding: 0, fontSize: 12, lineHeight: 1,
                }}
              >×</button>
            )}
          </span>
        ))}
        <input
          ref={inputRef}
          type="text"
          value={q}
          onChange={e => setQ(e.target.value)}
          onFocus={() => setFocused(true)}
          onBlur={() => setTimeout(() => setFocused(false), 150)}
          onKeyDown={onKeyDown}
          placeholder={atMax ? `max ${max} tags` : (tags.length === 0 ? placeholder : '')}
          disabled={disabled || atMax}
          spellCheck={false}
          style={{
            flex: '1 1 100px', minWidth: 80,
            background: 'transparent', border: 'none', outline: 'none',
            color: 'var(--ink)', fontFamily: 'inherit', fontSize: 11,
            letterSpacing: '0.04em',
            padding: '4px 2px',
          }}
        />
      </div>

      {focused && results.length > 0 && (
        <div style={{
          position: 'absolute', top: '100%', left: 0, right: 0, zIndex: 10,
          background: 'var(--panel)', border: '1px solid var(--rule-2)',
          borderTop: 'none', maxHeight: 200, overflowY: 'auto',
        }}>
          {results.map(r => (
            <div
              key={r.tag}
              onMouseDown={(e) => { e.preventDefault(); add(r.tag) }}
              style={{
                padding: '5px 10px', cursor: 'pointer', fontSize: 11,
                borderBottom: '1px solid var(--rule)',
                display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                letterSpacing: '0.04em', textTransform: 'uppercase',
              }}
              onMouseEnter={e => e.currentTarget.style.background = 'var(--bg-2, rgba(255,255,255,0.04))'}
              onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
            >
              <span>{r.tag}</span>
              <span className="t-xs muted">{r.count}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
