import { test, expect } from '@playwright/test'

const BASE_URL = process.env.ORCHESTRA_E2E_BASE_URL || 'http://127.0.0.1:5173'
const WORKSPACE_ID = '01KNAA9P1KCXH6SXG555S3NABY'

test.describe('Vue ChatStore WebSocket Integration', () => {
  test.beforeEach(async ({ page }) => {
    // Login
    await page.goto(`${BASE_URL}/login`)
    await page.locator('#orchestra-login-user').fill('orchestra')
    await page.locator('#orchestra-login-pass').fill('orchestra')
    await page.locator('button[type="submit"]').click()
    await expect(page).toHaveURL(/\/workspaces$/, { timeout: 5000 })
  })

  test('should initialize chatStore on workspace navigation', async ({ page }) => {
    // Navigate to workspace
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_ID}`)
    await page.waitForLoadState('networkidle')

    // Wait for Vue to initialize
    await page.waitForTimeout(5000)

    // Check the Pinia stores
    const storeState = await page.evaluate(() => {
      // Try to access Pinia stores via Vue app
      const app = document.querySelector('#app')?.__vue_app__
      if (!app) {
        return { error: 'Vue app not found' }
      }

      // Get Pinia instance
      const pinia = app.config.globalProperties.$pinia
      if (!pinia) {
        return { error: 'Pinia not found' }
      }

      // Check if chatStore exists
      const stores = pinia._s || new Map()
      const storeNames = Array.from(stores.keys())

      return {
        storeNames,
        hasChatStore: storeNames.includes('chat'),
        hasWorkspaceStore: storeNames.includes('workspace')
      }
    })

    console.log('Store state:', storeState)
    await page.screenshot({ path: 'artifacts/store-state.png' })

    expect(storeState.hasChatStore).toBe(true)
    expect(storeState.hasWorkspaceStore).toBe(true)
  })

  test('should have WebSocket connected after chatStore init', async ({ page }) => {
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_ID}/chat`)
    await page.waitForLoadState('networkidle')

    // Wait for WebSocket to connect
    await page.waitForTimeout(8000)

    // Check WebSocket state via page.evaluate
    const wsState = await page.evaluate(() => {
      // Check active WebSocket connections
      const wsInstances: any[] = []

      // Can't directly access WebSocket instances, but we can check if the app state is ready
      const app = document.querySelector('#app')?.__vue_app__
      if (!app) {
        return { error: 'Vue app not found' }
      }

      // Get Pinia state
      const pinia = app.config.globalProperties.$pinia
      if (!pinia) {
        return { error: 'Pinia not found' }
      }

      // Get chatStore state
      const stores = pinia._s || new Map()
      const chatStore = stores.get('chat')
      if (!chatStore) {
        return { error: 'chatStore not found' }
      }

      return {
        chatStoreReady: chatStore.isReady,
        conversationsCount: chatStore.conversations?.length || 0,
        activeConversationId: chatStore.activeConversationId,
        hasPolling: chatStore.pollTimer !== null
      }
    })

    console.log('WebSocket state:', wsState)
    await page.screenshot({ path: 'artifacts/ws-state-after-init.png' })
  })

  test('should receive message broadcast via Vue app', async ({ page, request }) => {
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_ID}/chat`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(5000)

    // Get conversation
    const convResp = await request.get(`http://localhost:8080/api/workspaces/${WORKSPACE_ID}/conversations`)
    const convs = await convResp.json()
    const convId = convs.timeline?.[0]?.id || convs.defaultChannelId

    if (!convId) {
      test.skip(true, 'No conversation')
      return
    }

    // Track message count before
    const beforeState = await page.evaluate(() => {
      const app = document.querySelector('#app')?.__vue_app__
      if (!app) return { error: 'No app' }

      const pinia = app.config.globalProperties.$pinia
      const stores = pinia._s || new Map()
      const chatStore = stores.get('chat')
      if (!chatStore) return { error: 'No chatStore' }

      // Get message count for the active conversation
      const conv = chatStore.conversations?.find(c => c.id === chatStore.activeConversationId)
      return {
        messageCount: conv?.messages?.length || 0,
        conversationId: chatStore.activeConversationId
      }
    })

    console.log('Before state:', beforeState)

    // Send a message via API
    await request.post(`http://localhost:8080/api/workspaces/${WORKSPACE_ID}/conversations/${convId}/messages`, {
      data: {
        text: 'Vue WebSocket test ' + Date.now(),
        senderId: 'owner',
        senderName: 'Vue Tester',
        timestamp: Date.now()
      }
    })

    // Wait for WebSocket broadcast
    await page.waitForTimeout(5000)

    // Check message count after
    const afterState = await page.evaluate((conversationId) => {
      const app = document.querySelector('#app')?.__vue_app__
      if (!app) return { error: 'No app' }

      const pinia = app.config.globalProperties.$pinia
      const stores = pinia._s || new Map()
      const chatStore = stores.get('chat')
      if (!chatStore) return { error: 'No chatStore' }

      const conv = chatStore.conversations?.find(c => c.id === conversationId)
      return {
        messageCount: conv?.messages?.length || 0,
        lastMessageText: conv?.messages?.[conv.messages.length - 1]?.content?.text
      }
    }, convId)

    console.log('After state:', afterState)
    await page.screenshot({ path: 'artifacts/vue-message-received.png' })

    // Check if message count increased
    expect(afterState.messageCount).toBeGreaterThan(beforeState.messageCount)
  })
})