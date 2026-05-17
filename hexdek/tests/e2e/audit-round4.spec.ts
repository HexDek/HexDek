import { test, expect } from '@playwright/test'

// Focused visual-polish round 4 audit. NOT part of the regular suite — pulls
// specific edge cases: long commander names on the deck shelf, hero name
// wrapping, no-analysis state, glossary popover overflow. Captures every
// screenshot under a dedicated audit-r4/ directory so it survives a
// subsequent test run wiping outputDir.

import * as fs from 'fs'
import * as path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const AUDIT_DIR = path.resolve(__dirname, '../../audit-r4')

test.beforeAll(() => {
  fs.mkdirSync(AUDIT_DIR, { recursive: true })
})

const shot = (name: string) => {
  const project = test.info().project.name
  return path.join(AUDIT_DIR, `${name}-${project}.png`)
}

test('deck shelf — long commander tiles', async ({ page }) => {
  await page.goto('/decks')
  await page.waitForFunction(
    () => !document.body.innerText.includes('LOADING DECK ARCHIVE'),
    null,
    { timeout: 15_000 }
  ).catch(() => {})
  await page.waitForTimeout(2000)
  // Zoom in on the moxfield tiles row — has Emet-Selch (very long), Clavileño,
  // Herigast, etc.
  await page.screenshot({ path: shot('deck-shelf'), fullPage: true })
})

test('deck hero — long commander (Emet-Selch)', async ({ page }) => {
  await page.goto('/decks/moxfield/emet_selch_unsundered_hades_sorcerer_of_eld_b3_bisqck_NM3FWTJ1')
  await expect(page.locator('h1')).toBeVisible({ timeout: 15_000 })
  await page.waitForTimeout(3000)
  // Hero-only crop: 0,0 → 1440×600 (desktop) / 412×500 (mobile)
  const isMobile = test.info().project.name === 'mobile'
  await page.screenshot({
    path: shot('hero-emet-selch'),
    clip: isMobile ? { x: 0, y: 0, width: 412, height: 600 } : { x: 0, y: 0, width: 1440, height: 700 },
  })
})

test('deck hero — extreme synthetic via param echo', async ({ page }) => {
  // Pick a deck guaranteed to load + override the deck-hero title in the DOM
  // post-render so we can stress-test layout on a name well past anything
  // real users have produced. Direct DOM injection — only valid post-mount.
  await page.goto('/decks/7174n1c/god_save_the_queen')
  await expect(page.locator('.deck-hero__title')).toBeVisible({ timeout: 15_000 })
  await page.waitForTimeout(1500)
  await page.evaluate(() => {
    const h1 = document.querySelector('.deck-hero__title') as HTMLElement | null
    if (h1) h1.textContent = 'NAEL DEUS DARNUS, DOOM OF THE TWELFTH ASTRAL ERA OF ETERNAL DOOM'
    const sub = document.querySelector('.deck-hero__sub') as HTMLElement | null
    if (sub) sub.textContent = 'Nael deus Darnus, Doom of the Twelfth Astral Era'
  })
  await page.waitForTimeout(400)
  const isMobile = test.info().project.name === 'mobile'
  await page.screenshot({
    path: shot('hero-synthetic-extreme'),
    clip: isMobile ? { x: 0, y: 0, width: 412, height: 700 } : { x: 0, y: 0, width: 1440, height: 700 },
  })
})

test('deck — no analysis state', async ({ page, request }) => {
  // Find a deck without analysis: try a few candidates by querying /api/decks
  // and probing /analysis. We mutate the page state by stubbing the analysis
  // endpoint to return 404 so we deterministically render the empty state.
  await page.route('**/api/decks/**/analysis', route => {
    route.fulfill({ status: 404, contentType: 'application/json', body: '{"error":"no analysis"}' })
  })
  await page.goto('/decks/7174n1c/god_save_the_queen')
  await expect(page.locator('h1')).toBeVisible({ timeout: 15_000 })
  await page.waitForTimeout(3000)
  // Switch to analysis tab if not already
  const analysisTab = page.locator('button.deck-tab:has-text("ANALYSIS")')
  if (await analysisTab.count() > 0) {
    await analysisTab.first().click().catch(() => {})
  }
  await page.waitForTimeout(800)
  await page.screenshot({ path: shot('no-analysis-state'), fullPage: true })
})

test('glossary popover — tooltip overflow', async ({ page }) => {
  await page.goto('/decks/7174n1c/god_save_the_queen')
  await expect(page.locator('h1')).toBeVisible({ timeout: 15_000 })
  await page.waitForTimeout(3000)
  // Find any GlossaryTerm trigger (button.gloss-trigger). Open the rightmost
  // one — most likely to cause right-edge overflow on tablet widths.
  const triggers = page.locator('button.gloss-trigger')
  const n = await triggers.count()
  if (n === 0) {
    test.info().annotations.push({ type: 'note', description: 'no gloss-trigger found on Marchesa page' })
    return
  }
  // Open trigger at the right edge by picking the one with max x.
  const boxes = await Promise.all(
    Array.from({ length: n }, (_, i) => triggers.nth(i).boundingBox())
  )
  let maxIdx = 0
  let maxX = -1
  boxes.forEach((b, i) => {
    if (b && b.x > maxX) { maxX = b.x; maxIdx = i }
  })
  await triggers.nth(maxIdx).scrollIntoViewIfNeeded()
  await page.waitForTimeout(300)
  await triggers.nth(maxIdx).click()
  await page.waitForTimeout(400)
  await page.screenshot({ path: shot('glossary-popover-right-edge'), fullPage: false })
})
