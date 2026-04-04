import { defineConfig, devices } from '@playwright/test'

/**
 * E2E: Vite preview starts via webServer (reuseExistingServer: true if already running).
 * Backend-dependent tests use request to ORCHESTRA_API_URL (default http://127.0.0.1:8080): health, /api/workspaces.
 * Run: `make run` in backend, then `pnpm test:e2e`
 */
export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  use: {
    baseURL: process.env.ORCHESTRA_E2E_BASE_URL ?? 'http://127.0.0.1:5173',
    trace: 'on-first-retry',
  },
  webServer: {
    command: 'pnpm exec vite preview --host 127.0.0.1 --port 5173 --strictPort',
    url: 'http://127.0.0.1:5173',
    reuseExistingServer: true,
    timeout: 60_000,
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
})
