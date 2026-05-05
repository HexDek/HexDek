// Temporal Pincer — get-or-create the anonymous browser ID used to
// stitch pre-auth activity to the authenticated owner on login.
const STORAGE_KEY = 'hexdek_anon_id'

function uuidv4() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  // Fallback — RFC 4122 §4.4 random UUID built on getRandomValues.
  const bytes = new Uint8Array(16)
  if (typeof crypto !== 'undefined' && crypto.getRandomValues) {
    crypto.getRandomValues(bytes)
  } else {
    for (let i = 0; i < bytes.length; i++) bytes[i] = Math.floor(Math.random() * 256)
  }
  bytes[6] = (bytes[6] & 0x0f) | 0x40
  bytes[8] = (bytes[8] & 0x3f) | 0x80
  const hex = [...bytes].map(b => b.toString(16).padStart(2, '0'))
  return `${hex.slice(0, 4).join('')}-${hex.slice(4, 6).join('')}-${hex.slice(6, 8).join('')}-${hex.slice(8, 10).join('')}-${hex.slice(10, 16).join('')}`
}

export function getAnonId() {
  if (typeof window === 'undefined' || !window.localStorage) return null
  let id = null
  try { id = window.localStorage.getItem(STORAGE_KEY) } catch { /* private mode */ }
  if (id && /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(id)) {
    return id
  }
  id = uuidv4()
  try { window.localStorage.setItem(STORAGE_KEY, id) } catch { /* ignore */ }
  return id
}
