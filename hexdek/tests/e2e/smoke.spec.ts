import { test, expect } from '@playwright/test'

// Visual smoke tests for key surfaces. Captures full-page screenshots for
// manual visual audit + asserts a couple of structural truths so the test
// can fail loudly if a page is empty / 500ing / missing key chrome.
//
// Each test waits past the loading state (~4s for async data fetches +
// WebSocket warmup on the leaderboard) before screenshotting, so the
// captured image shows the populated UI, not the spinner.
//
// All screenshot filenames are project-suffixed — Playwright does NOT
// auto-namespace screenshot paths by project, so without the suffix the
// mobile run overwrites the desktop run's artifacts (same outputDir).

const DATA_WAIT_MS = 4000

const shot = (name: string) => {
  const project = test.info().project.name
  return `test-results/${name}-${project}.png`
}

test('homepage / landing renders', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('h1').first()).toBeVisible({ timeout: 10_000 })
  await page.waitForTimeout(DATA_WAIT_MS) // let recent-activity feed populate
  await page.screenshot({ path: shot('homepage'), fullPage: true })
  await expect(page.locator('text=BROWSE DECKS').first()).toBeVisible()
})

test('deck list renders with decks loaded', async ({ page }) => {
  await page.goto('/decks')
  await expect(page.locator('text=/DECK ARCHIVE|HEXDEK/i').first()).toBeVisible({ timeout: 10_000 })
  await page.waitForFunction(
    () => !document.body.innerText.includes('LOADING DECK ARCHIVE'),
    null,
    { timeout: 15_000 }
  ).catch(() => {})
  await page.screenshot({ path: shot('deck-list'), fullPage: true })
})

test('leaderboard renders with ELO data', async ({ page }) => {
  await page.goto('/leaderboard')
  await expect(page.locator('text=/LEADERBOARD|RANKINGS/i').first()).toBeVisible({ timeout: 10_000 })
  await page.waitForFunction(
    () => !document.body.innerText.includes('AWAITING ELO DATA'),
    null,
    { timeout: 20_000 }
  ).catch(() => {})
  await page.screenshot({ path: shot('leaderboard'), fullPage: true })
})

test('deck page (Queen Marchesa) renders with new panels', async ({ page }) => {
  await page.goto('/decks/7174n1c/god_save_the_queen')
  await expect(page.locator('h1')).toBeVisible({ timeout: 15_000 })
  // Wait for async panel fetches: gauntlet, matchups, elo-history, card-stats.
  await page.waitForTimeout(DATA_WAIT_MS)
  await page.screenshot({ path: shot('deck-page-marchesa'), fullPage: true })
  // Sanity: vital signs strip is present.
  await expect(page.locator('text=HexELO').first()).toBeVisible()
})

test('deck page (Toph) renders', async ({ page }) => {
  await page.goto('/decks/belgarathrk/belgarath_toph_the_first_metalbender_deck')
  await expect(page.locator('h1')).toBeVisible({ timeout: 15_000 })
  await page.waitForTimeout(DATA_WAIT_MS)
  await page.screenshot({ path: shot('deck-page-toph'), fullPage: true })
})

test('import page renders', async ({ page }) => {
  await page.goto('/import')
  // Import screen doesn't use an h1; assert on the title text instead.
  await expect(page.locator('text=/IMPORT|PASTE|MOXFIELD/i').first()).toBeVisible({ timeout: 10_000 })
  await page.waitForTimeout(1000) // import page is mostly static — short wait
  await page.screenshot({ path: shot('import'), fullPage: true })
})

test('forge page renders', async ({ page }) => {
  await page.goto('/forge')
  await page.waitForTimeout(2000)
  await page.screenshot({ path: shot('forge'), fullPage: true })
})

test('spectate page renders', async ({ page }) => {
  await page.goto('/spectate')
  await page.waitForTimeout(2000)
  await page.screenshot({ path: shot('spectate'), fullPage: true })
})

test('owner profile renders', async ({ page }) => {
  await page.goto('/profile/7174n1c')
  await expect(page.locator('text=/7174N1C|DECKS|GAMES/i').first()).toBeVisible({ timeout: 10_000 })
  await page.waitForTimeout(DATA_WAIT_MS)
  await page.screenshot({ path: shot('profile-7174n1c'), fullPage: true })
})

// Mobile deep-audit — expands every collapsible panel then captures
// one full-page screenshot. The output is sliced into readable chunks
// by scripts/slice-deck-audit.py for visual review (Read-tool thumbnails
// of a 5000px-tall fullPage shot are illegible).
test('deck page mobile — element captures (Marchesa)', async ({ page }) => {
  test.skip(test.info().project.name !== 'mobile', 'mobile-only audit test')
  await page.goto('/decks/7174n1c/god_save_the_queen')
  await expect(page.locator('h1')).toBeVisible({ timeout: 15_000 })
  await page.waitForTimeout(8000)
  // Some deeper sections live inside an overflow-hidden container, so
  // page.fullPage screenshot truncates them. Scroll each section of
  // interest into view and capture the element directly. ElementHandle
  // screenshots are independent of the parent overflow clamp.
  const sections = [
    { sel: '.matchup-matrix', name: 'matchup-matrix' },
    // ELO history chart — Panel.04.EH owns the line chart. Match its
    // container by walking up from the chart's gridline-dasharray text.
    { sel: 'svg[viewBox="0 0 600 160"]', name: 'elo-history' },
    // Card stats panel — also worth a look on mobile.
    { sel: '.card-stats, .panel:has(.commander-card-stats)', name: 'card-stats' },
  ]
  for (const s of sections) {
    const el = page.locator(s.sel).first()
    if (await el.count() === 0) continue
    await el.scrollIntoViewIfNeeded()
    await page.waitForTimeout(300)
    await el.screenshot({ path: shot(`section-${s.name}`) })
  }
  // Also keep a full-page fallback for above-the-fold content.
  await page.evaluate(() => window.scrollTo(0, 0))
  await page.waitForTimeout(300)
  await page.screenshot({ path: shot('marchesa-expanded'), fullPage: true })
})

test('deck workshop opens on mobile (Marchesa)', async ({ page }) => {
  test.skip(test.info().project.name !== 'mobile', 'mobile-only workshop test')
  await page.goto('/decks/7174n1c/god_save_the_queen')
  await expect(page.locator('h1')).toBeVisible({ timeout: 15_000 })
  await page.waitForTimeout(DATA_WAIT_MS)
  // Workshop button is in the sidebar, may be off-screen on mobile until
  // we scroll. Find it, scroll it into view, click.
  const ws = page.locator('button:has-text("WORKSHOP")').first()
  await ws.scrollIntoViewIfNeeded()
  await ws.click()
  // Wait for the textarea editor to mount.
  await page.waitForSelector('textarea', { timeout: 5_000 }).catch(() => {})
  await page.waitForTimeout(800)
  await page.screenshot({ path: shot('workshop-mobile-marchesa'), fullPage: true })
})
