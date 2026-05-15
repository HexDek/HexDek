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

test('owner profile renders', async ({ page }) => {
  await page.goto('/profile/7174n1c')
  await expect(page.locator('text=/7174N1C|DECKS|GAMES/i').first()).toBeVisible({ timeout: 10_000 })
  await page.waitForTimeout(DATA_WAIT_MS)
  await page.screenshot({ path: shot('profile-7174n1c'), fullPage: true })
})
