const API_BASE = import.meta.env.VITE_API_URL ?? ''

async function request(path, opts = {}) {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...opts.headers },
    ...opts,
  })
  if (!res.ok) throw new Error(`API ${res.status}: ${path}`)
  return res.json()
}

export function cardArtUrl(name) {
  if (!name) return null
  const clean = name.split('//')[0].trim()
  return `${API_BASE}/api/card-art/${encodeURIComponent(clean)}`
}

export function cardImageUrl(name) {
  if (!name) return null
  const clean = name.split('//')[0].trim()
  return `${API_BASE}/api/card-art/${encodeURIComponent(clean)}?version=normal`
}

export { API_BASE }

export const api = {
  getDecks: (opts = {}) => {
    const params = new URLSearchParams()
    if (opts.owner) params.set('owner', opts.owner)
    if (opts.contains) params.set('contains', opts.contains)
    const qs = params.toString()
    return request(`/api/decks${qs ? `?${qs}` : ''}`)
  },
  getDeck: (id) => request(`/api/decks/${id}`),
  getDeckAnalysis: (id) => request(`/api/decks/${id}/analysis`),
  getProfile: () => request('/api/profile'),
  getGames: (limit = 20) => request(`/api/games?limit=${limit}`),
  getGame: (id) => request(`/api/games/${id}`),
  getGameReport: (id) => request(`/api/games/${id}/report`),
  getForgeStatus: () => request('/api/forge/status'),
  getForgeResults: (deckId) => request(`/api/forge/${deckId}/results`),
  startForge: (deckId, config) => request(`/api/forge/${deckId}/start`, { method: 'POST', body: JSON.stringify(config) }),
  getTournamentStats: () => request('/api/tournament/stats'),
  getLiveStats: () => request('/api/live/stats'),
  getLiveGame: () => request('/api/live/game'),
  getLiveELO: () => request('/api/live/elo'),
  importDeck: (name, owner, deckList) => request('/api/decks', {
    method: 'POST',
    body: JSON.stringify({ name, owner, deck_list: deckList }),
  }),
  // Full-page /import flow targets the dedicated alias route so the
  // backend can split metrics if we ever care to (same handler today).
  importDeckFull: ({ name, owner, deckList }) => request('/api/decks/import', {
    method: 'POST',
    body: JSON.stringify({ name, owner, deck_list: deckList }),
  }),
  importMoxfield: ({ url, owner }) => request('/api/import/moxfield', {
    method: 'POST',
    body: JSON.stringify({ url, owner }),
  }),
  searchCards: (q, limit = 6) => request(`/api/cards/search?q=${encodeURIComponent(q)}&limit=${limit}`),
  runAnalysis: (id) => request(`/api/decks/${id}/analyze`, { method: 'POST' }),
  updateDeck: (id, deckList) => request(`/api/decks/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ deck_list: deckList }),
  }),
  deleteDeck: (id) => request(`/api/decks/${id}`, { method: 'DELETE' }),
  patchDeck: (id, fields) => request(`/api/decks/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(fields),
  }),
  getDeckVersions: (id) => request(`/api/decks/${id}/versions`),
  getDeckCurse: (id) => request(`/api/decks/${id}/curse`),
  getSimilarDecks: (id, limit = 5) => request(`/api/decks/${id}/similar?limit=${limit}`),
  getAchievements: (owner) => request(`/api/achievements/${owner}`),
  setUserCountry: (owner) => request(`/api/user/profile/country`, {
    method: 'POST',
    body: JSON.stringify({ owner }),
  }),
  getOwnerProfile: (owner) => request(`/api/profile/${encodeURIComponent(owner)}`),
  getOwnerProfiles: (owners) => {
    const list = (owners || []).filter(Boolean).join(',')
    if (!list) return Promise.resolve({})
    return request(`/api/profiles?owners=${encodeURIComponent(list)}`)
  },
  getImports: (owner, limit = 10) => request(`/api/imports/${encodeURIComponent(owner)}?limit=${limit}`),
  startGauntlet: (id, games = 500) => request(`/api/gauntlet/${id}?games=${games}`, { method: 'POST' }),
  getGauntlet: (id) => request(`/api/gauntlet/${id}`),
  getDonationsSummary: () => request('/api/donations/summary'),
  search: (q, limit = 6) => request(`/api/search?q=${encodeURIComponent(q)}&limit=${limit}`),
  listFriends: (asSlug) => request(`/api/friends?as=${encodeURIComponent(asSlug)}`),
  addFriend: (target, asSlug) => request(`/api/friends/${encodeURIComponent(target)}?as=${encodeURIComponent(asSlug)}`, { method: 'POST' }),
  removeFriend: (target, asSlug) => request(`/api/friends/${encodeURIComponent(target)}?as=${encodeURIComponent(asSlug)}`, { method: 'DELETE' }),
}
