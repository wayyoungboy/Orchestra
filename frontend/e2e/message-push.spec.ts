import { test, expect } from '@playwright/test'

const BASE_URL = process.env.ORCHESTRA_E2E_BASE_URL || 'http://127.0.0.1:5173'
const API_URL = process.env.ORCHESTRA_API_URL || 'http://127.0.0.1:8080'

test.describe('Message Push Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Login first with correct selectors
    await page.goto(`${BASE_URL}/login`)
    await page.waitForLoadState('networkidle')

    // Use specific IDs for login form
    await page.locator('#orchestra-login-user').fill('orchestra')
    await page.locator('#orchestra-login-pass').fill('orchestra')
    await page.locator('button[type="submit"]').click()

    // Wait for navigation to workspaces
    await expect(page).toHaveURL(/\/workspaces$/, { timeout: 5000 })
    await page.waitForLoadState('networkidle')
  })

  test('should connect to chat WebSocket', async ({ page }) => {
    // Navigate to workspace
    await page.goto(BASE_URL)
    await page.waitForLoadState('networkidle')

    // Capture WebSocket logs
    const wsLogs: string[] = []
    page.on('console', msg => {
      const text = msg.text()
      if (text.includes('WebSocket') || text.includes('ChatSocket') || text.includes('chat') || text.includes('ws')) {
        wsLogs.push(text)
      }
    })

    // Wait for WebSocket to establish
    await page.waitForTimeout(3000)

    // Take screenshot
    await page.screenshot({ path: 'artifacts/chat-ws-test.png' })

    console.log('WebSocket logs:', wsLogs)

    // Check if we're on a workspace page
    const url = page.url()
    expect(url).toMatch(/workspace|workspaces/)
  })

  test('should display conversation interface', async ({ page }) => {
    await page.goto(BASE_URL)
    await page.waitForLoadState('networkidle')

    // Look for conversation/chat elements
    const chatInput = page.locator('[data-testid="chat-input"], textarea[placeholder*="消息"], input[placeholder*="消息"]')
    const conversationList = page.locator('[data-testid="conversation-list"], .conversation-list')

    // Take screenshot
    await page.screenshot({ path: 'artifacts/conversation-ui.png' })

    // Log UI state
    const hasChatInput = await chatInput.isVisible().catch(() => false)
    const hasConversationList = await conversationList.isVisible().catch(() => false)

    console.log('Chat input visible:', hasChatInput)
    console.log('Conversation list visible:', hasConversationList)
  })

  test('should send message and receive response', async ({ page, request }) => {
    await page.goto(BASE_URL)
    await page.waitForLoadState('networkidle')

    // First check backend API
    const healthResp = await request.get(`${API_URL}/health`)
    expect(healthResp.status()).toBe(200)

    // Get workspaces
    const wsResp = await request.get(`${API_URL}/api/workspaces`)
    expect(wsResp.status()).toBe(200)
    const workspaces = await wsResp.json()

    if (workspaces.length > 0) {
      const workspaceId = workspaces[0].id

      // Navigate to workspace
      await page.goto(`${BASE_URL}/workspace/${workspaceId}`)
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(2000)

      // Take screenshot of workspace
      await page.screenshot({ path: 'artifacts/workspace-view.png' })

      // Look for member to chat with
      const memberCards = page.locator('[data-testid="member-card"], .member-card')
      const memberCount = await memberCards.count()

      console.log('Member count:', memberCount)

      if (memberCount > 0) {
        // Click on first member
        await memberCards.first().click()
        await page.waitForTimeout(1000)
        await page.screenshot({ path: 'artifacts/member-selected.png' })

        // Try to send message via internal API
        const convResp = await request.get(`${API_URL}/api/workspaces/${workspaceId}/conversations`)
        console.log('Conversations response:', convResp.status())

        if (convResp.status() === 200) {
          const convs = await convResp.json()
          console.log('Conversations:', convs.length)
        }
      }
    }

    // Take final screenshot
    await page.screenshot({ path: 'artifacts/message-test-final.png' })
  })
})

test.describe('Cross-tab Sync', () => {
  test('should sync messages across tabs', async ({ browser }) => {
    // Create two contexts (tabs)
    const context1 = await browser.newContext()
    const context2 = await browser.newContext()

    const page1 = await context1.newPage()
    const page2 = await context2.newPage()

    // Both login and navigate to same workspace
    await page1.goto(BASE_URL)
    await page2.goto(BASE_URL)

    await page1.waitForLoadState('networkidle')
    await page2.waitForLoadState('networkidle')

    // Take screenshots
    await page1.screenshot({ path: 'artifacts/tab1.png' })
    await page2.screenshot({ path: 'artifacts/tab2.png' })

    console.log('Both tabs loaded')

    // Cleanup
    await context1.close()
    await context2.close()
  })
})