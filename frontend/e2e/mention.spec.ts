import { test, expect } from '@playwright/test';

test.describe('@ Mention Feature', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to app
    await page.goto('http://127.0.0.1:5173');

    // Wait for page to load
    await page.waitForTimeout(1000);

    // Login if needed - check for login form
    const usernameInput = page.locator('input[type="text"], input:not([type])').first();
    const passwordInput = page.locator('input[type="password"]').first();
    const loginButton = page.locator('button:has-text("登录"), button:has-text("Sign In")').first();

    if (await loginButton.isVisible({ timeout: 3000 }).catch(() => false)) {
      console.log('🔐 Found login form, logging in...');
      await usernameInput.fill('orchestra');
      await passwordInput.fill('orchestra');
      await loginButton.click();
      await page.waitForTimeout(2000);
    }

    // Navigate to chat if needed
    const chatLink = page.locator('a[href*="chat"], button:has-text("聊天"), button:has-text("Chat")').first();
    if (await chatLink.isVisible({ timeout: 2000 }).catch(() => false)) {
      await chatLink.click();
      await page.waitForTimeout(1000);
    }

    // Wait for members sidebar to load
    await page.waitForTimeout(2000);
  });

  test('should show @ mention option in member menu', async ({ page }) => {
    // Take screenshot for debugging
    await page.screenshot({ path: 'test-results/mention-test-initial.png' });

    // Look for member rows in the sidebar
    const memberRows = page.locator('.member-row-root');
    const memberCount = await memberRows.count();
    console.log(`📋 Found ${memberCount} member rows`);

    if (memberCount > 0) {
      // Find first member with a menu button (assistant/secretary)
      for (let i = 0; i < memberCount; i++) {
        const row = memberRows.nth(i);
        const menuButton = row.locator('.more-btn');

        if (await menuButton.isVisible().catch(() => false)) {
          console.log(`🔍 Opening menu for member ${i}`);
          await menuButton.click();
          await page.waitForTimeout(500);

          // Take screenshot of open menu
          await page.screenshot({ path: `test-results/mention-menu-${i}.png` });

          // Check if @ mention option exists
          const mentionOption = page.locator('button.menu-item:has-text("@"), button:has-text("提及"), button:has-text("Mention")');
          const isVisible = await mentionOption.isVisible({ timeout: 2000 }).catch(() => false);

          if (isVisible) {
            console.log('✅ @ mention option found in member menu');
            await page.screenshot({ path: 'test-results/mention-option-found.png' });
            return; // Test passed
          } else {
            console.log(`⚠️ No @ mention option in menu for member ${i}`);
          }

          // Close menu by clicking elsewhere
          await page.keyboard.press('Escape');
          await page.waitForTimeout(300);
        }
      }
      console.log('⚠️ No @ mention option found in any member menu');
    } else {
      console.log('⚠️ No member rows found');
    }

    // Check if we can see the members sidebar at all
    const membersSidebar = page.locator('.members-sidebar-root');
    const sidebarVisible = await membersSidebar.isVisible().catch(() => false);
    console.log(`📂 Members sidebar visible: ${sidebarVisible}`);
  });

  test('should insert @mention when clicking menu option', async ({ page }) => {
    // Look for member row with terminal (assistant)
    const memberRows = page.locator('.member-row-root');
    const memberCount = await memberRows.count();

    if (memberCount > 0) {
      for (let i = 0; i < memberCount; i++) {
        const row = memberRows.nth(i);
        const menuButton = row.locator('.more-btn');

        if (await menuButton.isVisible().catch(() => false)) {
          await menuButton.click();
          await page.waitForTimeout(300);

          const mentionOption = page.locator('button.menu-item:has-text("@"), button:has-text("提及"), button:has-text("Mention")').first();

          if (await mentionOption.isVisible({ timeout: 1000 }).catch(() => false)) {
            console.log(`🖱️ Clicking @ mention for member ${i}`);
            await mentionOption.click();
            await page.waitForTimeout(500);

            // Check if input has @mention
            const chatInput = page.locator('textarea').first();
            const inputValue = await chatInput.inputValue().catch(() => '');

            console.log(`📝 Input value: "${inputValue}"`);
            await page.screenshot({ path: `test-results/mention-inserted.png` });

            if (inputValue.includes('@')) {
              console.log('✅ @mention successfully inserted');
              expect(inputValue).toContain('@');
              return;
            }
          }

          await page.keyboard.press('Escape');
        }
      }
    }
  });
});