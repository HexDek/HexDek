import { defineConfig, devices } from '@playwright/test'

// Playwright config — runs visual smoke tests against dev.hexdek.dev
// (the staging site populated by ./scripts/deploy.sh frontend-dev).
//
// Usage:
//   cd hexdek && npx playwright test                  # all tests
//   npx playwright test --project=desktop             # desktop only
//   npx playwright test --project=mobile              # mobile only
//   npx playwright test --update-snapshots            # rebaseline
//
// Screenshots land in hexdek/test-results/ and hexdek/playwright-report/.
// CI mode (PLAYWRIGHT_CI=1) enables retries + headless-only.
export default defineConfig({
  testDir: './tests/e2e',
  outputDir: './test-results',
  reporter: process.env.PLAYWRIGHT_CI ? 'github' : 'list',
  fullyParallel: true,
  forbidOnly: !!process.env.PLAYWRIGHT_CI,
  retries: process.env.PLAYWRIGHT_CI ? 2 : 0,
  workers: process.env.PLAYWRIGHT_CI ? 1 : undefined,
  use: {
    baseURL: process.env.HEXDEK_E2E_URL || 'https://dev.hexdek.dev',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    ignoreHTTPSErrors: true,
  },
  projects: [
    {
      name: 'desktop',
      use: { ...devices['Desktop Chrome'], viewport: { width: 1440, height: 900 } },
    },
    {
      name: 'mobile',
      use: { ...devices['Pixel 7'] },
    },
  ],
})
