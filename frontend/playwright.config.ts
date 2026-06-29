import { defineConfig, devices } from '@playwright/test';

// Full-stack E2E config. The suite runs against a real backend brought up with
// `npm run e2e:up` (Postgres + Go API on :8080) and the Angular dev server,
// which Playwright starts via `webServer` below.
//
// Workers are pinned to 1: the tests share one seeded database and each resets
// it to a known baseline before running (see e2e/fixtures.ts), so they must not
// run concurrently.
export default defineConfig({
  testDir: './e2e/specs',
  fullyParallel: false,
  workers: 1,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [['html', { open: 'never' }], ['list']] : 'list',
  globalSetup: './e2e/global-setup.ts',
  timeout: 30_000,
  expect: { timeout: 7_500 },
  use: {
    baseURL: 'http://localhost:4200',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
  // Start `ng serve` (which proxies /api + /auth/google to the backend on :8080
  // per proxy.conf.json). Reuse a running dev server locally; always fresh in CI.
  webServer: {
    command: 'npm start',
    url: 'http://localhost:4200',
    reuseExistingServer: !process.env.CI,
    timeout: 180_000, // cold `ng serve` build can be slow in CI
  },
});
