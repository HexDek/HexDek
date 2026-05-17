import { test, expect, Page, Locator } from '@playwright/test'

// Visual-polish round 5 — deck page deep-analysis section sweep on mobile.
// For each Panel inside the ANALYSIS tab we scroll to it, clip to its
// bounding box, and write a screenshot under audit-r5/. The goal is to
// surface overflowing tables, unreadable text, broken empty states, and
// dead-space layouts at Pixel 7 width (412px).
//
// Heavy on screenshots, light on assertions — this is an audit, not a
// regression gate. Failures here are exploratory.

import * as fs from 'fs'
import * as path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const AUDIT_DIR = path.resolve(__dirname, '../../audit-r5')

test.beforeAll(() => {
  fs.mkdirSync(AUDIT_DIR, { recursive: true })
})

const shot = (name: string) => {
  const project = test.info().project.name
  return path.join(AUDIT_DIR, `${name}-${project}.png`)
}

// Decks chosen to hit different analysis shapes:
//   - god_save_the_queen — owner content, long Marchesa win-lines table
//   - emet_selch — long commander, lots of Freya output
//   - a moxfield deck pulled from the archive
const DECKS = [
  { slug: '/decks/7174n1c/god_save_the_queen', tag: 'marchesa' },
  { slug: '/decks/moxfield/emet_selch_unsundered_hades_sorcerer_of_eld_b3_bisqck_NM3FWTJ1', tag: 'emet' },
]

async function gotoDeck(page: Page, slug: string) {
  await page.goto(slug)
  await expect(page.locator('h1')).toBeVisible({ timeout: 20_000 })
  // Wait for analysis to load: either the FREYA panel content or the
  // "no analysis on file" empty state. Either way, the page settles.
  await page.waitForTimeout(3500)
  // Make sure the ANALYSIS tab is active.
  const analysisTab = page.locator('button.deck-tab:has-text("ANALYSIS")').first()
  if (await analysisTab.count() > 0) {
    const isActive = await analysisTab.evaluate(el => el.classList.contains('active'))
    if (!isActive) await analysisTab.click()
    await page.waitForTimeout(500)
  }
}

async function clipShot(page: Page, locator: Locator, file: string, padBottom = 12) {
  await locator.scrollIntoViewIfNeeded()
  // Some panels have a sticky header above them; nudge up slightly.
  await page.evaluate(() => window.scrollBy(0, -60))
  await page.waitForTimeout(250)
  const box = await locator.boundingBox()
  if (!box) return false
  const vp = page.viewportSize() || { width: 412, height: 915 }
  const clip = {
    x: Math.max(0, Math.floor(box.x)),
    y: Math.max(0, Math.floor(box.y)),
    width: Math.min(vp.width, Math.ceil(box.width)),
    height: Math.min(vp.height, Math.ceil(box.height) + padBottom),
  }
  if (clip.height <= 4 || clip.width <= 4) return false
  await page.screenshot({ path: file, clip })
  return true
}

// Walk the .archive-main column and screenshot every direct child .panel
// plus a few specific known-tricky regions (matchup matrix, gauntlet
// placements, ELO history svg).
async function capturePanels(page: Page, tag: string) {
  // Full page first for context.
  await page.screenshot({ path: shot(`${tag}-full`), fullPage: true })

  // Each Panel — use its rendered code (04.x) prefix as the filename.
  const panels = page.locator('.archive-main .panel')
  const n = await panels.count()
  for (let i = 0; i < n; i++) {
    const p = panels.nth(i)
    // Pull the code header (e.g. "04.C") to name the screenshot.
    const code = await p.locator('.panel-hd .muted-2').first().textContent().catch(() => null)
    const title = await p.locator('.panel-hd > span').first().textContent().catch(() => null)
    const safe = (code || `panel${i}`).trim().replace(/[^A-Za-z0-9_.-]/g, '_').slice(0, 40)
    const titleSafe = (title || '').trim().replace(/[^A-Za-z0-9_.-]/g, '_').slice(0, 40)
    await clipShot(page, p, shot(`${tag}-${safe}-${titleSafe}`))
  }
}

for (const deck of DECKS) {
  test(`deep analysis sweep — ${deck.tag}`, async ({ page }) => {
    test.setTimeout(120_000)
    await gotoDeck(page, deck.slug)
    await capturePanels(page, deck.tag)
  })
}

// Targeted: no-analysis empty state on mobile (separate so it stands out
// in the audit dir).
test('deep analysis — no-analysis state mobile', async ({ page }) => {
  await page.route('**/api/decks/**/analysis', route => {
    route.fulfill({ status: 404, contentType: 'application/json', body: '{"error":"no analysis"}' })
  })
  await gotoDeck(page, '/decks/7174n1c/god_save_the_queen')
  await page.screenshot({ path: shot('no-analysis-full'), fullPage: true })
  const freya = page.locator('.archive-main .panel').first()
  await clipShot(page, freya, shot('no-analysis-freya-panel'))
})

// Targeted: gauntlet-running state. Stub the gauntlet endpoint so the
// "running" empty state always renders (we don't want to wait for a real
// run on a real deck during an audit).
test('deep analysis — gauntlet running state mobile', async ({ page }) => {
  await page.route('**/api/decks/**/gauntlet', route => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        status: 'running',
        games: 137,
        target: 500,
        win_rate: 24,
      }),
    })
  })
  await gotoDeck(page, '/decks/7174n1c/god_save_the_queen')
  const gauntletPanel = page.locator('.archive-main .panel:has-text("GAUNTLET REPORT")').first()
  if (await gauntletPanel.count() === 0) {
    test.info().annotations.push({ type: 'note', description: 'gauntlet panel not rendered — owner-only?' })
    return
  }
  await clipShot(page, gauntletPanel, shot('gauntlet-running'))
})
