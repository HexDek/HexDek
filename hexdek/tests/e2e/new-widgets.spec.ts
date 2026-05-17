import { test, expect } from '@playwright/test'

// Coverage for the new deck-page widgets that landed on the analysis tab:
//   - 04.HC HOT CARDS         (top WR-contribution cards from card-stats)
//   - 04.SD SIMILAR DECKS     (thumbnail tiles for peer builds)
//
// Both panels render conditionally — HOT CARDS only mounts when commander
// card-stats include ≥1 card with positive lift, and SIMILAR DECKS only
// mounts when /api/decks/:id/similar resolves with at least one match.
// We target Marchesa (god_save_the_queen) because it has the data depth
// (200+ gauntlets, lots of shared-card peers) to surface both reliably.
//
// Test runs against the same baseURL as smoke.spec.ts (dev.hexdek.dev by
// default; override with HEXDEK_E2E_URL).

const DECK_URL = '/decks/7174n1c/god_save_the_queen'
const DATA_WAIT_MS = 6000

const shot = (name: string) => {
  const project = test.info().project.name
  return `test-results/${name}-${project}.png`
}

// Locate a Panel by its title text in the .panel-hd span. Code prefixes
// like "04.SD" can substring-match neighbor panels (e.g. "04.S DECK STATS"
// rendered without a space → "04.SDECK STATS"), so we match on title.
const panelByTitle = (page: import('@playwright/test').Page, title: RegExp) =>
  page.locator('.panel', { has: page.locator('.panel-hd', { hasText: title }) }).first()

test.describe('new deck-page widgets (Marchesa)', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(DECK_URL)
    await expect(page.locator('h1')).toBeVisible({ timeout: 15_000 })
    // The hot-cards + similar-decks fetches fan out alongside gauntlet,
    // matchups, elo-history, card-stats — give them all room to land.
    await page.waitForTimeout(DATA_WAIT_MS)
  })

  test('HOT CARDS panel renders with tile content', async ({ page }) => {
    const hot = panelByTitle(page, /HOT CARDS/i)
    await expect(hot, 'HOT CARDS panel must be present').toBeVisible({ timeout: 10_000 })

    // Each tile is a CardThumb — the body contains at least one card image.
    // We assert ≥1 img so an empty grid (commander-stats fetch failed silently)
    // hard-fails the test.
    const tiles = hot.locator('.panel-bd img')
    await expect(tiles.first()).toBeVisible({ timeout: 10_000 })
    const count = await tiles.count()
    expect(count, 'HOT CARDS should have at least one tile').toBeGreaterThan(0)

    await hot.scrollIntoViewIfNeeded()
    await page.waitForTimeout(300)
    await hot.screenshot({ path: shot('widget-hot-cards') })
  })

  test('SIMILAR DECKS panel renders with peer tiles', async ({ page }) => {
    const sim = panelByTitle(page, /SIMILAR DECKS/i)
    await expect(sim, 'SIMILAR DECKS panel must be present').toBeVisible({ timeout: 10_000 })

    // Each peer is a Link with class "panel" pointing at /decks/:owner/:id.
    const tiles = sim.locator('.panel-bd a[href^="/decks/"]')
    await expect(tiles.first()).toBeVisible({ timeout: 10_000 })
    const count = await tiles.count()
    expect(count, 'SIMILAR DECKS should have at least one peer tile').toBeGreaterThan(0)

    await sim.scrollIntoViewIfNeeded()
    await page.waitForTimeout(300)
    await sim.screenshot({ path: shot('widget-similar-decks') })
  })
})
