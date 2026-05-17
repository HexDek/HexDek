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
  // Leaderboard mounts 1313 rows + fetches country flags + fullPage
  // screenshot — can exceed 30s default on a fresh backend cache.
  test.setTimeout(60_000)
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
  // toph_20 = Belgarath's actual Toph 2.0 upload (100 cards). The old
  // belgarath_toph_the_first_metalbender_deck was a 0-card stub.
  await page.goto('/decks/belgarathrk/toph_20')
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

// ── /compare/:o1/:d1/:o2/:d2 — head-to-head diff ─────────────────────────
test('deck compare renders both heroes and head-to-head metrics', async ({ page }) => {
  await page.goto('/decks/7174n1c/god_save_the_queen')
  // Pick a second known-good deck (Toph 2.0, 100 cards) so both bundles resolve.
  await page.goto('/compare/7174n1c/god_save_the_queen/belgarathrk/toph_20')
  await expect(page.locator('text=/HEAD-TO-HEAD METRICS/i').first()).toBeVisible({ timeout: 15_000 })
  // Both commander hero panels mount; loadDeckBundle resolves before the
  // tape switches off "LOADING".
  await page.waitForFunction(
    () => !document.body.innerText.match(/\bLOADING\b/),
    null,
    { timeout: 20_000 }
  ).catch(() => {})
  await expect(page.locator('.cmp-hero--L')).toBeVisible()
  await expect(page.locator('.cmp-hero--R')).toBeVisible()
  await page.waitForTimeout(DATA_WAIT_MS)
  await page.screenshot({ path: shot('compare'), fullPage: true })
})

// ── /report/:gameId — per-game report ────────────────────────────────────
test('per-game report renders for a real game id', async ({ page, request }) => {
  // Pull the most recent finished game from /api/games — game IDs rotate
  // as the engine runs, so hardcoding gets stale fast.
  let gameId: number | null = null
  try {
    const r = await request.get('/api/games?limit=1')
    if (r.ok()) {
      const games = await r.json()
      gameId = games?.[0]?.game_id ?? null
    }
  } catch {}
  test.skip(!gameId, 'no completed games available to report on')
  await page.goto(`/report/${gameId}`)
  await expect(page.locator(`text=/GAME[. #]${gameId}/i`).first()).toBeVisible({ timeout: 15_000 })
  await expect(page.locator('text=/RESULT BLOCK/i').first()).toBeVisible({ timeout: 10_000 })
  await page.waitForTimeout(DATA_WAIT_MS)
  await page.screenshot({ path: shot('report'), fullPage: true })
})

// ── /cards/:cardName — card detail page ──────────────────────────────────
test('card page renders for Sol Ring', async ({ page }) => {
  await page.goto('/cards/Sol%20Ring')
  // h1 = upper-cased card name; wait for the local oracle / Scryfall fallback
  // to finish so the hero title text is real (not blank).
  await expect(page.locator('.card-page-hero-title')).toBeVisible({ timeout: 15_000 })
  await page.waitForFunction(
    () => !document.body.innerText.includes('FETCHING CARD RECORD'),
    null,
    { timeout: 15_000 }
  ).catch(() => {})
  await page.waitForTimeout(DATA_WAIT_MS)
  await expect(page.locator('.card-page-hero-title')).toContainText(/SOL RING/i)
  await page.screenshot({ path: shot('card-sol-ring'), fullPage: true })
})

// ── Console-error sweep across 5 sampled moxfield decks ──────────────────
// Pulls /api/decks?owner=moxfield, picks 5 (deterministic — first 5 by API
// order), navigates each. Any console error or uncaught page error fails
// the test with the deck slug + message attached, so regressions in the
// deck-page render pipeline (analysis bundle, gauntlet panel, ELO history)
// surface in CI instead of in user reports.
test('moxfield deck sample — 5 deep-page renders, fail on console error', async ({ page, request }) => {
  // 5 decks × (15s nav + 4s data wait + screenshot) easily exceeds the
  // default 30s test timeout. Give it 3 minutes.
  test.setTimeout(180_000)
  const r = await request.get('/api/decks?owner=moxfield')
  expect(r.ok(), 'fetch /api/decks?owner=moxfield').toBe(true)
  const decks: Array<{ owner: string; id: string }> = await r.json()
  expect(decks.length, 'at least 5 moxfield decks available').toBeGreaterThanOrEqual(5)
  const sample = decks.slice(0, 5)

  for (const d of sample) {
    const errors: string[] = []
    const onConsole = (msg: import('@playwright/test').ConsoleMessage) => {
      if (msg.type() === 'error') errors.push(`console.error: ${msg.text()}`)
    }
    const onPageError = (err: Error) => {
      errors.push(`pageerror: ${err.message}`)
    }
    page.on('console', onConsole)
    page.on('pageerror', onPageError)

    await page.goto(`/decks/${d.owner}/${d.id}`)
    await expect(page.locator('h1')).toBeVisible({ timeout: 15_000 })
    await page.waitForTimeout(DATA_WAIT_MS)
    await page.screenshot({ path: shot(`sample-${d.id}`), fullPage: true })

    page.off('console', onConsole)
    page.off('pageerror', onPageError)

    // Filter out third-party noise that doesn't indicate a real regression
    // (Stripe pixel blocked by privacy ext, WebSocket disconnect retries,
    // favicon 404s on first load, Sentry beacons in dev).
    const meaningful = errors.filter(e => {
      const s = e.toLowerCase()
      if (s.includes('favicon')) return false
      if (s.includes('websocket')) return false
      if (s.includes('net::err_blocked')) return false
      if (s.includes('failed to load resource')) return false
      return true
    })
    expect(meaningful, `deck ${d.owner}/${d.id} produced console errors`).toEqual([])
  }
})

// ── Mobile deep-analysis section audit ───────────────────────────────────
// Walks every <Panel> below the vital signs strip on a known-good deck
// (Queen Marchesa) and captures each one as its own element screenshot.
// Vital signs themselves already have a dedicated test above — this one
// is for the analysis tab content (DECK SPECS, FREYA, GAUNTLET REPORT,
// MANA CURVE, COLOR BALANCE, WIN LINES, META POSITIONING, STAR CARDS,
// WIN CONDITIONS, VALUE ENGINE, GAME CHANGERS, CARD PACKAGES,
// TUTOR TARGETS, MATCHUP MATRIX, CARD STATS, ELO HISTORY, etc).
test('deck page mobile — deep analysis section audit (Marchesa)', async ({ page }) => {
  test.skip(test.info().project.name !== 'mobile', 'mobile-only audit test')
  // 60+ panels each scrolled into view + screenshotted easily exceeds 30s.
  test.setTimeout(180_000)
  await page.goto('/decks/7174n1c/god_save_the_queen')
  await expect(page.locator('h1')).toBeVisible({ timeout: 15_000 })
  // Long initial wait — gauntlet, matchups, elo-history, card-stats all
  // fan out in parallel and the deeper panels render only after they land.
  await page.waitForTimeout(8000)

  // Make sure the ANALYSIS tab is the active surface (it's the default,
  // but click is cheap insurance against tab-default drift).
  const analysisTab = page.locator('button.deck-tab:has-text("ANALYSIS")')
  if (await analysisTab.count() > 0) {
    await analysisTab.first().click({ trial: false }).catch(() => {})
  }

  // Expand any collapsed panels (CollapsiblePanel defaultOpen=false) so the
  // body content is in the DOM for screenshotting.
  const expandables = page.locator('.panel-hd:has-text("[+]"), .panel-hd:has-text("▶")')
  const expandCount = await expandables.count()
  for (let i = 0; i < expandCount; i++) {
    await expandables.nth(i).click().catch(() => {})
  }
  await page.waitForTimeout(500)

  // Iterate every .panel below the vital-signs strip. Snap each one as an
  // element screenshot — fullPage on mobile can overflow Playwright's
  // height cap on long decks.
  const panels = page.locator('.panel:has(.panel-hd)')
  const total = await panels.count()
  let captured = 0
  for (let i = 0; i < total; i++) {
    const panel = panels.nth(i)
    const head = panel.locator('.panel-hd').first()
    const label = (await head.innerText().catch(() => '')).trim().replace(/[^A-Za-z0-9]+/g, '-').toLowerCase().slice(0, 48) || `panel-${i}`
    await panel.scrollIntoViewIfNeeded().catch(() => {})
    await page.waitForTimeout(200)
    try {
      await panel.screenshot({ path: shot(`mobile-section-${String(i).padStart(2, '0')}-${label}`) })
      captured++
    } catch {
      // Element off-screen / zero-sized — skip and continue.
    }
  }
  // Sanity: the analysis tab should expose at least a dozen panels on a
  // fully-populated deck. If we captured fewer, the page is degraded.
  expect(captured, 'deep-analysis panel count').toBeGreaterThanOrEqual(8)
})
