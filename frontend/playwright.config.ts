import { defineConfig, devices } from '@playwright/test'

/**
 * E2E starts an isolated, auth-disabled Orchestra backend and proxies Vite
 * preview traffic to it. Override the URLs only when testing an external stack.
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
  webServer: [
    {
      command: 'go run ./cmd/server',
      cwd: '../backend',
      env: {
        ORCHESTRA_CONFIG: 'configs/config.e2e.yaml'
      },
      url: process.env.ORCHESTRA_API_URL ?? 'http://127.0.0.1:18080/health',
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
    },
    {
      command: 'pnpm exec vite preview --host 127.0.0.1 --port 5173 --strictPort',
      env: {
        ORCHESTRA_API_URL: process.env.ORCHESTRA_API_URL ?? 'http://127.0.0.1:18080'
      },
      url: process.env.ORCHESTRA_E2E_BASE_URL ?? 'http://127.0.0.1:5173',
      reuseExistingServer: !process.env.CI,
      timeout: 60_000,
    }
  ],
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
})
