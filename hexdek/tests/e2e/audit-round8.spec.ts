import { test, expect, Page, Locator } from '@playwright/test'

// Visual-polish round 8 — Hot Cards (04.HC) + Similar Decks tile (04.SD)
// + Similar Decks sidebar widget. Screenshots on desktop and Pixel 7;
// audit focus is alignment, overflow, hover states, click targets, and
// empty states.

import * as fs from 'fs'
import * as path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const AUDIT_DIR = path.resolve(__dirname, '../../audit-r8')

test.beforeAll(() => {
  fs.mkdirSync(AUDIT_DIR, { recursive: true })
})

const shot = (name: string) => {
  const project = test.info().project.name
  return path.join(AUDIT_DIR, `${name}-${project}.png`)
}

const DECKS = [
  { slug: '/decks/7174n1c/god_save_the_queen', tag: 'marchesa' },
  { slug: '/decks/moxfield/emet_selch_unsundered_hades_sorcerer_of_eld_b3_bisqck_NM3FWTJ1', tag: 'emet' },
]

async function gotoDeck(page: Page, slug: string) {
  await page.goto(slug)
  await expect(page.locator('h1')).toBeVisible({ timeout: 20_000 })
  await page.waitForTimeout(3500)
  const analysisTab = page.locator('button.deck-tab:has-text("ANALYSIS")').first()
  if (await analysisTab.count() > 0) {
    const isActive = await analysisTab.evaluate(el => el.classList.contains('active'))
    if (!isActive) await analysisTab.click()
    await page.waitForTimeout(800)
  }
  // Give similar-decks fetch + commanderCardStats time to land.
  await page.waitForTimeout(2500)
}

async function clipShot(page: Page, locator: Locator, file: string, padBottom = 12) {
  await locator.scrollIntoViewIfNeeded()
  await page.evaluate(() => window.scrollBy(0, -60))
  await page.waitForTimeout(250)
  const box = await locator.boundingBox()
  if (!box) return false
  const vp = page.viewportSize() || { width: 1440, height: 900 }
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

for (const deck of DECKS) {
  test(`hot-cards + similar-decks sweep — ${deck.tag}`, async ({ page }) => {
    test.setTimeout(120_000)
    await gotoDeck(page, deck.slug)

    // Sidebar SIMILAR DECKS widget (text list).
    const sidebarSD = page.locator('.archive-sidebar .panel:has-text("SIMILAR DECKS")').first()
    if (await sidebarSD.count() > 0) {
      await clipShot(page, sidebarSD, shot(`${deck.tag}-similar-decks-sidebar`))
    }

    // Main column HOT CARDS panel.
    const hotCards = page.locator('.archive-main .panel:has-text("HOT CARDS")').first()
    if (await hotCards.count() > 0) {
      await clipShot(page, hotCards, shot(`${deck.tag}-hot-cards`))
      // Hover state on first tile.
      const firstTile = hotCards.locator('div').filter({ has: page.locator('img') }).first()
      if (await firstTile.count() > 0) {
        await firstTile.hover()
        await page.waitForTimeout(400)
        await clipShot(page, hotCards, shot(`${deck.tag}-hot-cards-hover`))
      }
    } else {
      test.info().annotations.push({ type: 'note', description: `no HOT CARDS panel on ${deck.tag}` })
    }

    // Main column SIMILAR DECKS thumbnail tile panel.
    const tileSD = page.locator('.archive-main .panel:has-text("SIMILAR DECKS")').first()
    if (await tileSD.count() > 0) {
      await clipShot(page, tileSD, shot(`${deck.tag}-similar-decks-tiles`))
      const firstLink = tileSD.locator('a').first()
      if (await firstLink.count() > 0) {
        await firstLink.hover()
        await page.waitForTimeout(400)
        await clipShot(page, tileSD, shot(`${deck.tag}-similar-decks-tiles-hover`))
      }
    } else {
      test.info().annotations.push({ type: 'note', description: `no SIMILAR DECKS tile panel on ${deck.tag}` })
    }
  })
}

// Stubbed-data run for HOT CARDS — dev backend doesn't have ≥20 games
// for any commander yet, so the panel never renders against live data.
// Inject a deterministic dataset that lines up with cards Marchesa decks
// commonly run; uses PascalCase to match the live backend schema.
test('hot-cards panel — stubbed data', async ({ page }) => {
  await page.route('**/api/card-stats/**', route => {
    const cards = [
      { CardName: 'Sol Ring', Games: 412, Wins: 168, WinRate: 0.408 },
      { CardName: 'Command Tower', Games: 380, Wins: 144, WinRate: 0.379 },
      { CardName: 'Arcane Signet', Games: 350, Wins: 128, WinRate: 0.366 },
      { CardName: 'Cyclonic Rift', Games: 220, Wins: 92, WinRate: 0.418 },
      { CardName: 'Lightning Greaves', Games: 180, Wins: 70, WinRate: 0.389 },
      { CardName: 'Swords to Plowshares', Games: 140, Wins: 50, WinRate: 0.357 },
      { CardName: 'Counterspell', Games: 90, Wins: 28, WinRate: 0.311 },
    ]
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ cards }) })
  })
  await gotoDeck(page, '/decks/7174n1c/god_save_the_queen')
  const hotCards = page.locator('.archive-main .panel:has-text("HOT CARDS")').first()
  if (await hotCards.count() === 0) {
    test.info().annotations.push({ type: 'note', description: 'HOT CARDS panel did not render even with stubbed data' })
    await page.screenshot({ path: shot('hot-cards-stub-full'), fullPage: true })
    return
  }
  await clipShot(page, hotCards, shot('hot-cards-stubbed'))
  const firstTile = hotCards.locator('.hot-cards-tile').first()
  if (await firstTile.count() > 0) {
    await firstTile.hover()
    await page.waitForTimeout(400)
    await clipShot(page, hotCards, shot('hot-cards-stubbed-hover'))
  }
})

// Empty-state run: stub similar-decks + commander card-stats endpoints
// so both panels collapse to their respective empty states.
test('hot-cards + similar-decks — empty states', async ({ page }) => {
  await page.route('**/api/decks/**/similar*', route => {
    route.fulfill({ status: 200, contentType: 'application/json', body: '[]' })
  })
  await page.route('**/api/cards/commander/**', route => {
    route.fulfill({ status: 200, contentType: 'application/json', body: '[]' })
  })
  await gotoDeck(page, '/decks/7174n1c/god_save_the_queen')
  // Sidebar widget always renders even when empty — capture its empty branch.
  const sidebarSD = page.locator('.archive-sidebar .panel:has-text("SIMILAR DECKS")').first()
  if (await sidebarSD.count() > 0) {
    await clipShot(page, sidebarSD, shot('empty-similar-decks-sidebar'))
  }
  await page.screenshot({ path: shot('empty-full'), fullPage: true })
})
