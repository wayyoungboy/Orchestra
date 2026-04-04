import { test, expect } from '@playwright/test'

/**
 * 占位认证流（development-roadmap-parity-01 R01-3）。
 * 不依赖后端；用固定 id 选择器，避免随 i18n 语言变化。
 */
test('login with default credentials navigates to workspaces', async ({ page }) => {
  await page.goto('/login')
  await page.locator('#orchestra-login-user').fill('orchestra')
  await page.locator('#orchestra-login-pass').fill('orchestra')
  await page.locator('button[type="submit"]').click()
  await expect(page).toHaveURL(/\/workspaces$/)
})
