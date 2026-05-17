import { test, expect, type APIRequestContext, type Locator } from '@playwright/test'

// Coverage for the post-game replay viewer (Report.jsx → ReplayScrubber),
// which landed via dev/replay-viewer-v2. Finds a recent CompletedGame
// that carries a per-turn timeline (games played after replay-viewer-v2
// landed), opens /report/{id}, scrubs the slider three positions, and
// verifies the turn header, per-seat board area, and event log all
// respond.
//
// Older games rehydrated from SQLite have no timeline — the component
// renders a "NO PER-TURN TIMELINE" fallback panel with no slider, so
// those games are skipped, not failed.

const MIN_TURNS = 4 // need ≥ 4 snapshots to scrub 3 positions ahead
const SCAN_LIMIT = 30

async function findGameWithTimeline(request: APIRequestContext): Promise<number | null> {
  const list = await request.get(`/api/games?limit=${SCAN_LIMIT}`)
  if (!list.ok()) return null
  const games: Array<{ game_id: number }> = await list.json()
  for (const g of games) {
    const detail = await request.get(`/api/games/${g.game_id}/report`)
    if (!detail.ok()) continue
    const body: { timeline?: unknown[] } = await detail.json()
    if (Array.isArray(body.timeline) && body.timeline.length >= MIN_TURNS) {
      return g.game_id
    }
  }
  return null
}

test('replay scrubber advances turn, board, and event log', async ({ page, request }) => {
  test.setTimeout(120_000)

  const gameId = await findGameWithTimeline(request)
  test.skip(
    !gameId,
    `no completed game with timeline length ≥ ${MIN_TURNS} found in last ${SCAN_LIMIT} games`,
  )

  await page.goto(`/report/${gameId}`)
  await expect(page.locator('text=/REPLAY VIEWER/i').first()).toBeVisible({ timeout: 15_000 })

  const slider = page.locator('input[aria-label="Turn slider"]')
  await expect(slider).toBeVisible({ timeout: 10_000 })

  // The replay panel is the .panel ancestor that owns the slider — scope
  // every other locator to it so unrelated panels (RESULT BLOCK, etc.)
  // can't leak into our frame snapshots.
  const replayPanel = page.locator('div.panel', { has: slider }).first()
  const turnHeader = replayPanel.locator('text=/TURN \\d+ OF \\d+/').first()
  const eventsHeader = replayPanel.locator('text=/EVENTS — TURN \\d+/').first()
  const boardGrid: Locator = replayPanel.locator('.grid.col-4').first()
  // The events list (or "— NO RECORDED EVENTS —" fallback) sits as the
  // sibling immediately after the EVENTS header.
  const eventsBody = eventsHeader.locator('xpath=following-sibling::*[1]')

  await expect(turnHeader).toBeVisible()
  await expect(eventsHeader).toBeVisible()
  await expect(boardGrid).toBeVisible()

  const headerSnaps: string[] = []
  const boardSnaps: string[] = []
  const eventSnaps: string[] = []

  const capture = async () => {
    headerSnaps.push(((await turnHeader.textContent()) || '').trim())
    boardSnaps.push(((await boardGrid.textContent()) || '').trim())
    eventSnaps.push(((await eventsBody.textContent().catch(() => '')) || '').trim())
  }

  await capture()

  // Scrub three positions forward. ArrowRight on a focused range input
  // increments by `step` (default 1), so this lands on turnIdx 1, 2, 3.
  await slider.focus()
  for (let i = 0; i < 3; i++) {
    await page.keyboard.press('ArrowRight')
    // Small wait so React commits the new snapshot before we read.
    await page.waitForTimeout(150)
    await capture()
  }

  // The turn header includes the per-snapshot turn number, and each
  // timeline snapshot is one game turn, so all four frames must differ.
  expect(
    new Set(headerSnaps).size,
    `turn header should advance every scrub — saw ${JSON.stringify(headerSnaps)}`,
  ).toBe(4)

  // Board + event log are turn-derived. They aren't strictly required to
  // shift on every single tick (e.g. two adjacent quiet turns with the
  // same battlefield + zero events could repeat verbatim), but across
  // four frames we should see at least two distinct values for each.
  expect(
    new Set(boardSnaps).size,
    `per-seat board area should update as the slider moves — saw ${boardSnaps.length} frames, ${new Set(boardSnaps).size} unique`,
  ).toBeGreaterThanOrEqual(2)
  expect(
    new Set(eventSnaps).size,
    `event log should update as the slider moves — saw ${eventSnaps.length} frames, ${new Set(eventSnaps).size} unique`,
  ).toBeGreaterThanOrEqual(2)

  // Final sanity: the slider's own value should have advanced to 3.
  await expect(slider).toHaveValue('3')
})
