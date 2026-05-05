// i18n foundation for hexdek.dev.
//
// Locale catalogs are JSON files under src/locales/. Add a new language
// by dropping a new file (e.g. src/locales/de.json) and importing it
// into the LOCALES map below.
//
// Usage from a component:
//
//   import { useTranslation } from '../i18n'
//   const { t, locale, setLocale, availableLocales } = useTranslation()
//   return <span>{t('nav.decks')}</span>
//
// Keys are dotted paths into the JSON (`nav.decks` → catalog.nav.decks).
// Missing keys fall back to English, then to the literal key string.
// {var} placeholders are interpolated from the second argument:
//
//   t('hello', { name: 'wiedeman' })  // "Hello, wiedeman"

import { useSyncExternalStore } from 'react'

import en from './locales/en.json'
import es from './locales/es.json'

// Add new languages here. Keep keys ISO 639-1 (en, de, ja, ...).
const LOCALES = { en, es }
const FALLBACK = 'en'
const STORAGE_KEY = 'hexdek.locale'

let currentLocale = (() => {
  try {
    const saved = typeof localStorage !== 'undefined' && localStorage.getItem(STORAGE_KEY)
    if (saved && LOCALES[saved]) return saved
  } catch {}
  return FALLBACK
})()

const subscribers = new Set()

function notify() {
  for (const fn of subscribers) fn()
}

function subscribe(fn) {
  subscribers.add(fn)
  return () => subscribers.delete(fn)
}

function getSnapshot() {
  return currentLocale
}

export function setLocale(locale) {
  if (!LOCALES[locale] || locale === currentLocale) return
  currentLocale = locale
  try { localStorage.setItem(STORAGE_KEY, locale) } catch {}
  notify()
}

export function getLocale() {
  return currentLocale
}

export function availableLocales() {
  return Object.keys(LOCALES)
}

// lookup walks a dotted path through a catalog object. Returns undefined
// for any missing segment.
function lookup(catalog, key) {
  const parts = key.split('.')
  let cursor = catalog
  for (const part of parts) {
    if (cursor == null || typeof cursor !== 'object') return undefined
    cursor = cursor[part]
  }
  return cursor
}

// interpolate replaces {var} tokens with values from vars. Tokens
// without a matching var are left as-is so missing data is visible.
function interpolate(str, vars) {
  if (!vars || typeof str !== 'string') return str
  return str.replace(/\{(\w+)\}/g, (m, name) => (vars[name] != null ? String(vars[name]) : m))
}

// t resolves a key against the current locale, falling back to English
// then to the raw key. Module-level export for use outside React.
export function t(key, vars) {
  const catalog = LOCALES[currentLocale] || LOCALES[FALLBACK]
  let value = lookup(catalog, key)
  if (value === undefined && currentLocale !== FALLBACK) {
    value = lookup(LOCALES[FALLBACK], key)
  }
  if (value === undefined) return key
  return interpolate(value, vars)
}

// useTranslation subscribes the component to locale changes and
// returns the translator + setter. No Provider required.
export function useTranslation() {
  const locale = useSyncExternalStore(subscribe, getSnapshot, getSnapshot)
  return {
    t: (key, vars) => t(key, vars),
    locale,
    setLocale,
    availableLocales: availableLocales(),
  }
}
