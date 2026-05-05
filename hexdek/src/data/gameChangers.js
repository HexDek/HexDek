// Mirror of internal/cmd/hexdek-freya/archetype.go gameChangersList.
// 53 cards on the WotC Commander Game Changers list (Feb 2026).
// Keep in sync with the Go source of truth.
export const GAME_CHANGERS = new Set([
  // White
  'drannith magistrate', 'enlightened tutor', 'farewell',
  'humility', "teferi's protection", 'smothering tithe',
  // Blue
  'consecrated sphinx', 'cyclonic rift', 'force of will',
  'fierce guardianship', 'gifts ungiven', 'intuition',
  'mystical tutor', 'narset, parter of veils', 'rhystic study',
  "thassa's oracle",
  // Black
  'ad nauseam', "bolas's citadel", 'braids, cabal minion',
  'demonic tutor', 'imperial seal', 'necropotence',
  'opposition agent', 'orcish bowmasters',
  'tergrid, god of fright', 'vampiric tutor',
  // Red
  'gamble', "jeska's will", 'underworld breach',
  // Green
  'biorhythm', 'crop rotation', 'natural order',
  'seedborn muse', 'survival of the fittest', 'worldly tutor',
  // Multicolor
  'aura shards', 'coalition victory',
  'grand arbiter augustin iv', 'notion thief',
  // Colorless
  'ancient tomb', 'chrome mox', 'field of the dead',
  "gaea's cradle", 'glacial chasm', 'grim monolith',
  "lion's eye diamond", 'mana vault', "mishra's workshop",
  'mox diamond', 'panoptic mirror', "serra's sanctum",
  'the one ring', 'the tabernacle at pendrell vale',
])

export function isGameChanger(name) {
  if (!name) return false
  return GAME_CHANGERS.has(String(name).toLowerCase().trim())
}

// Returns the matched GC name (lowercase) if any appears as a token in the
// uppercase action text emitted by the log builder, else null. Order is
// longest-first to avoid "gamble" matching inside a longer card name.
const SORTED_GCS = [...GAME_CHANGERS].sort((a, b) => b.length - a.length)
export function findGameChangerInText(text) {
  if (!text) return null
  const haystack = String(text).toLowerCase()
  for (const name of SORTED_GCS) {
    if (haystack.includes(name)) return name
  }
  return null
}
