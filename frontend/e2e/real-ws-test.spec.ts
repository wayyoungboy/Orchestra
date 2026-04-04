import { test, expect } from '@playwright/test'

const BASE_URL = process.env.ORCHESTRA_E2E_BASE_URL || 'http://127.0.0.1:5173'
const WORKSPACE_ID = '01KNAA9P1KCXH6SXG555S3NABY'

test.describe('Real WebSocket Connection Test', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(`${BASE_URL}/login`)
    await page.locator('#orchestra-login-user').fill('orchestra')
    await page.locator('#orchestra-login-pass').fill('orchestra')
    await page.locator('button[type="submit"]').click()
    await expect(page).toHaveURL(/\/workspaces$/, { timeout: 5000 })
  })

  test('should verify WebSocket connection state in browser', async ({ page }) => {
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_ID}/chat`)
    await page.waitForLoadState('networkidle')

    // Wait for WebSocket to potentially connect
    await page.waitForTimeout(8000)

    // Check actual WebSocket instances in browser
    const wsCheck = await page.evaluate(() => {
      // Override WebSocket constructor to track connections
      const originalWS = window.WebSocket
      const connections: any[] = []

      // Check Performance API for WebSocket connections
      const entries = performance.getEntriesByType('resource')
      const wsEntries = entries.filter(e => e.name.includes('ws://') || e.name.includes('wss://'))

      return {
        wsResourceEntries: wsEntries.map(e => e.name),
        activeWebSocketCount: (window as any).__activeWebSockets || 'not tracked'
      }
    })

    console.log('WebSocket resource entries:', wsCheck.wsResourceEntries)
    await page.screenshot({ path: 'artifacts/ws-resource-check.png' })
  })

  test('should receive InternalChatSend message via WebSocket', async ({ page }) => {
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_ID}/chat`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(5000)

    // Get initial message count
    const beforeState = await page.evaluate(() => {
      const app = document.querySelector('#app')?.__vue_app__
      if (!app) return { error: 'No app' }
      const pinia = app.config.globalProperties.$pinia
      const stores = pinia._s || new Map()
      const chatStore = stores.get('chat')
      if (!chatStore) return { error: 'No chatStore' }
      const conv = chatStore.conversations?.find((c: any) => c.id === chatStore.activeConversationId)
      return {
        messageCount: conv?.messages?.length || 0,
        conversationId: chatStore.activeConversationId
      }
    })

    console.log('Before InternalChatSend:', beforeState)

    const conversationId = beforeState.conversationId || 'conv_97049025'

    // Send message via InternalChatSend API (simulates AI response)
    const response = await page.evaluate(async (convId) => {
      const resp = await fetch('/api/internal/chat/send', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          workspaceId: '01KNAA9P1KCXH6SXG555S3NABY',
          conversationId: convId,
          senderId: '01KNAA9P1KCXH6SXG559R94TAR',
          senderName: 'AI Assistant',
          text: 'Real AI response test via InternalChatSend ' + Date.now()
        })
      })
      return { status: resp.status, body: await resp.json() }
    }, conversationId)

    console.log('InternalChatSend response:', response)

    // Wait for WebSocket broadcast
    await page.waitForTimeout(5000)

    // Get message count after
    const afterState = await page.evaluate(() => {
      const app = document.querySelector('#app')?.__vue_app__
      if (!app) return { error: 'No app' }
      const pinia = app.config.globalProperties.$pinia
      const stores = pinia._s || new Map()
      const chatStore = stores.get('chat')
      if (!chatStore) return { error: 'No chatStore' }
      const conv = chatStore.conversations?.find((c: any) => c.id === chatStore.activeConversationId)
      const lastMsg = conv?.messages?.[conv.messages.length - 1]
      return {
        messageCount: conv?.messages?.length || 0,
        lastMessageText: lastMsg?.content?.text,
        lastMessageSender: lastMsg?.senderName,
        lastMessageIsAi: lastMsg?.isAi
      }
    })

    console.log('After InternalChatSend:', afterState)
    await page.screenshot({ path: 'artifacts/internal-send-received.png' })

    // Verify message was received
    expect(afterState.messageCount).toBeGreaterThan(beforeState.messageCount)
    expect(afterState.lastMessageIsAi).toBe(true)
    expect(afterState.lastMessageText).toContain('Real AI response test')
  })

  test('should test with manual WebSocket listener', async ({ page }) => {
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_ID}/chat`)
    await page.waitForLoadState('networkidle')

    // Create a manual WebSocket listener and then trigger InternalChatSend
    const result = await page.evaluate(async () => {
      const messages: string[] = []
      const convId = 'conv_97049025'

      return new Promise<any>(async (resolve) => {
        // Create WebSocket
        const ws = new WebSocket(`ws://${window.location.host}/ws/chat/01KNAA9P1KCXH6SXG555S3NABY`)

        ws.onopen = async () => {
          messages.push('WS_CONNECTED')

          // Call InternalChatSend
          const resp = await fetch('/api/internal/chat/send', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              workspaceId: '01KNAA9P1KCXH6SXG555S3NABY',
              conversationId: convId,
              senderId: '01KNAA9P1KCXH6SXG559R94TAR',
              senderName: 'Manual WS Test AI',
              text: 'Manual WebSocket test ' + Date.now()
            })
          })
          messages.push(`API_RESPONSE: ${resp.status}`)
        }

        ws.onmessage = (event) => {
          messages.push(`WS_MESSAGE: ${event.data}`)
        }

        ws.onerror = (e) => {
          messages.push(`WS_ERROR: ${e}`)
        }

        // Timeout after 10 seconds
        setTimeout(() => {
          ws.close()
          resolve({ messages })
        }, 10000)
      })
    })

    console.log('Manual WebSocket test result:', result)
    await page.screenshot({ path: 'artifacts/manual-ws-internal-send.png' })

    // Check if we received the broadcast
    const hasNewMessage = result.messages.some((m: string) => m.includes('new_message'))
    expect(hasNewMessage).toBe(true)
  })
})