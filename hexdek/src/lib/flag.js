// countryFlagEmoji turns a 2-letter ISO 3166-1 alpha-2 country code into
// the matching flag emoji using the Regional Indicator code points. Each
// uppercase ASCII letter A-Z maps to U+1F1E6 + (letter - 'A'), and a
// pair of regional indicators is rendered as a single flag glyph by the
// platform's emoji font.
//
// Returns "" for missing/invalid input so callers can use it inline:
//   <span>{countryFlagEmoji(profile?.country)} {owner}</span>
//
// We deliberately don't try to translate language codes ("en") into a
// country — the backend already extracted the region segment from
// Accept-Language. If the backend returns "", we render nothing.
export function countryFlagEmoji(code) {
  if (typeof code !== 'string') return ''
  const cc = code.trim().toUpperCase()
  if (cc.length !== 2) return ''
  const A = 0x1F1E6
  const c0 = cc.charCodeAt(0)
  const c1 = cc.charCodeAt(1)
  if (c0 < 65 || c0 > 90 || c1 < 65 || c1 > 90) return ''
  return String.fromCodePoint(A + (c0 - 65)) + String.fromCodePoint(A + (c1 - 65))
}
