import { test, expect } from '@playwright/test'

test('login page loads', async ({ page }) => {
  await page.goto('/login')
  await expect(page.locator('body')).toBeVisible()
})

test('backend health', async ({ request }) => {
  const api = process.env.ORCHESTRA_API_URL ?? 'http://127.0.0.1:8080'
  const result = await request.get(`${api}/health`).catch((err: Error) => err)
  if (result instanceof Error) {
    test.skip(true, `no backend at ${api} (start server or set ORCHESTRA_API_URL)`)
    return
  }
  expect(result.ok()).toBeTruthy()
  const body = await result.json()
  expect(body).toHaveProperty('status', 'ok')
})
