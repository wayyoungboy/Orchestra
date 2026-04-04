import { test, expect } from '@playwright/test'

/**
 * Requires Orchestra backend (default http://127.0.0.1:8080).
 * Set ORCHESTRA_API_URL if the API is on another host/port.
 */
test('GET /api/workspaces returns JSON array when backend up', async ({ request }) => {
  const base = process.env.ORCHESTRA_API_URL ?? 'http://127.0.0.1:8080'
  const result = await request.get(`${base}/api/workspaces`).catch((err: Error) => err)
  if (result instanceof Error) {
    test.skip(true, `no backend at ${base} (start server or set ORCHESTRA_API_URL)`)
    return
  }
  expect(result.ok()).toBeTruthy()
  const body = await result.json()
  expect(Array.isArray(body)).toBeTruthy()
})
