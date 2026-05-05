import { useState, useEffect, useRef, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../services/api'

const KIND_GLYPH = {
  deck: '▣',
  commander: '◆',
  owner: '◉',
  card: '▢',
}

const KIND_LABEL = {
  deck: 'DECK',
  commander: 'CMDR',
  owner: 'PLAYER',
  card: 'CARD',
}

function flattenResults(payload) {
  if (!payload) return []
  if (Array.isArray(payload.top) && payload.top.length) return payload.top
  const r = payload.results || {}
  return [
    ...(r.owners || []),
    ...(r.commanders || []),
    ...(r.decks || []),
    ...(r.cards || []),
  ]
}

function resolveHref(item) {
  if (!item) return null
  if (item.kind === 'deck' && item.owner && item.id) {
    return `/decks/${encodeURIComponent(item.owner)}/${encodeURIComponent(item.id)}`
  }
  if (item.kind === 'commander') {
    return `/decks?tab=all&q=${encodeURIComponent(item.label)}`
  }
  if (item.kind === 'owner') {
    return `/decks?tab=all&q=${encodeURIComponent(item.label)}`
  }
  if (item.kind === 'card') {
    return `/decks?tab=all&q=${encodeURIComponent(item.label)}`
  }
  return null
}

export default function SearchBar() {
  const [query, setQuery] = useState('')
  const [items, setItems] = useState([])
  const [open, setOpen] = useState(false)
  const [active, setActive] = useState(0)
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const containerRef = useRef(null)
  const inputRef = useRef(null)
  const reqIdRef = useRef(0)

  useEffect(() => {
    const trimmed = query.trim()
    if (trimmed.length < 2) {
      setItems([])
      setLoading(false)
      return
    }
    setLoading(true)
    const reqId = ++reqIdRef.current
    const handle = setTimeout(async () => {
      try {
        const data = await api.search(trimmed, 5)
        if (reqId !== reqIdRef.current) return
        const flat = flattenResults(data)
        setItems(flat)
        setActive(0)
      } catch {
        if (reqId !== reqIdRef.current) return
        setItems([])
      } finally {
        if (reqId === reqIdRef.current) setLoading(false)
      }
    }, 180)
    return () => clearTimeout(handle)
  }, [query])

  useEffect(() => {
    const onDocClick = (e) => {
      if (containerRef.current && !containerRef.current.contains(e.target)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', onDocClick)
    return () => document.removeEventListener('mousedown', onDocClick)
  }, [])

  useEffect(() => {
    const onKey = (e) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        inputRef.current?.focus()
        setOpen(true)
      }
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [])

  const choose = useCallback((item) => {
    if (!item) return
    const href = resolveHref(item)
    setOpen(false)
    setQuery('')
    setItems([])
    if (href) navigate(href)
  }, [navigate])

  const onKeyDown = (e) => {
    if (!open && (e.key === 'ArrowDown' || e.key === 'Enter')) {
      setOpen(true)
    }
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setActive(i => Math.min(i + 1, Math.max(items.length - 1, 0)))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setActive(i => Math.max(i - 1, 0))
    } else if (e.key === 'Enter') {
      e.preventDefault()
      if (items[active]) choose(items[active])
    } else if (e.key === 'Escape') {
      setOpen(false)
      inputRef.current?.blur()
    }
  }

  const showDropdown = open && query.trim().length >= 2

  return (
    <div ref={containerRef} className="searchbar" style={{ position: 'relative' }}>
      <span className="searchbar__icon" aria-hidden>⌖</span>
      <input
        ref={inputRef}
        type="text"
        value={query}
        onChange={e => { setQuery(e.target.value); setOpen(true) }}
        onFocus={() => setOpen(true)}
        onKeyDown={onKeyDown}
        placeholder="SEARCH DECKS · CMDRS · PLAYERS · CARDS"
        aria-label="Universal search"
        spellCheck={false}
        autoComplete="off"
        className="searchbar__input"
      />
      <span className="searchbar__hint">⌘K</span>
      {showDropdown && (
        <div className="searchbar__dropdown">
          {loading && items.length === 0 && (
            <div className="searchbar__row searchbar__row--note">SCANNING…</div>
          )}
          {!loading && items.length === 0 && (
            <div className="searchbar__row searchbar__row--note">NO RESULTS</div>
          )}
          {items.map((item, i) => (
            <button
              key={`${item.kind}-${item.label}-${item.owner || ''}-${item.id || ''}-${i}`}
              type="button"
              onMouseEnter={() => setActive(i)}
              onClick={() => choose(item)}
              className={`searchbar__row ${i === active ? 'is-active' : ''}`}
            >
              <span className="searchbar__glyph" aria-hidden>
                {KIND_GLYPH[item.kind] || '·'}
              </span>
              <span className="searchbar__kind">{KIND_LABEL[item.kind] || item.kind?.toUpperCase()}</span>
              <span className="searchbar__label">{item.label}</span>
              {item.sub && <span className="searchbar__sub">{item.sub}</span>}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
