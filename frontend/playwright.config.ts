import { defineConfig, devices } from '@playwright/test'

const isCI = !!process.env.CI

/**
 * End-to-end tests run against the real stack: the API, the web app and the
 * local services (PostgreSQL, Firebase Auth emulator) started by `make up`,
 * with the example data of `make db-seed`. `make e2e` prepares all of it.
 *
 * Tests share one database, so they run one at a time and create their own
 * data with unique emails.
 */
export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  workers: 1,
  retries: isCI ? 1 : 0,
  forbidOnly: isCI,
  reporter: isCI ? [['github'], ['html', { open: 'never' }]] : [['list']],
  use: {
    baseURL: 'http://localhost:5173',
    locale: 'es-ES',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: [
    {
      command: 'go run ./cmd/api',
      cwd: '../backend',
      url: 'http://localhost:8080/api/v1/health',
      reuseExistingServer: !isCI,
      timeout: 180_000,
    },
    {
      command: 'pnpm dev --strictPort',
      url: 'http://localhost:5173',
      reuseExistingServer: !isCI,
      timeout: 60_000,
    },
  ],
})
