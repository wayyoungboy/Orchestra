import { test, expect } from '@playwright/test'

const BASE_URL = process.env.ORCHESTRA_E2E_BASE_URL || 'http://127.0.0.1:5173'
const API_URL = process.env.ORCHESTRA_API_URL || 'http://127.0.0.1:8080'
const WORKSPACE_ID = '01KNAA9P1KCXH6SXG555S3NABY'

test.describe('Direct WebSocket Test', () => {
  test.beforeEach(async ({ page }) => {
    // Login
    await page.goto(`${BASE_URL}/login`)
    await page.locator('#orchestra-login-user').fill('orchestra')
    await page.locator('#orchestra-login-pass').fill('orchestra')
    await page.locator('button[type="submit"]').click()
    await expect(page).toHaveURL(/\/workspaces$/, { timeout: 5000 })
  })

  test('should manually connect WebSocket in browser', async ({ page, context }) => {
    // Navigate to workspace
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_ID}`)
    await page.waitForLoadState('networkidle')

    // Create a manual WebSocket connection and monitor it
    const wsResult = await page.evaluate(() => {
      return new Promise<{ connected: boolean; messages: string[]; error?: string }>((resolve) => {
        const messages: string[] = []
        const wsUrl = `ws://${window.location.host}/ws/chat/${'01KNAA9P1KCXH6SXG555S3NABY'}`

        try {
          const ws = new WebSocket(wsUrl)

          ws.onopen = () => {
            messages.push('WebSocket connected!')
            // Send a ping
            ws.send(JSON.stringify({ type: 'ping' }))
          }

          ws.onmessage = (event) => {
            messages.push(`Message received: ${event.data}`)
          }

          ws.onerror = (error) => {
            messages.push(`WebSocket error: ${error}`)
          }

          ws.onclose = (event) => {
            messages.push(`WebSocket closed: code=${event.code}, reason=${event.reason}`)
            resolve({ connected: false, messages })
          }

          // Timeout after 10 seconds
          setTimeout(() => {
            if (ws.readyState === WebSocket.OPEN) {
              ws.close()
              resolve({ connected: true, messages })
            } else {
              resolve({ connected: false, messages, error: 'Connection timeout' })
            }
          }, 10000)
        } catch (e) {
          resolve({ connected: false, messages, error: String(e) })
        }
      })
    })

    console.log('WebSocket test result:', wsResult)
    await page.screenshot({ path: 'artifacts/manual-ws-test.png' })

    // Verify WebSocket connected
    expect(wsResult.connected).toBe(true)
    expect(wsResult.messages).toContain('WebSocket connected!')
  })

  test('should check if ChatSocket singleton exists', async ({ page }) => {
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_ID}`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(3000)

    // Check if the ChatSocket is available
    const chatSocketState = await page.evaluate(() => {
      // Try to access global ChatSocket
      // @ts-ignore
      const socket = window.__chatSocket || null
      return {
        hasSocket: socket !== null,
        isConnected: socket?.isConnected || false,
        workspaceId: socket?.currentWorkspaceId || null
      }
    })

    console.log('ChatSocket state:', chatSocketState)
    await page.screenshot({ path: 'artifacts/chatsocket-state.png' })
  })

  test('should send message and wait for WebSocket notification', async ({ page, request }) => {
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_ID}`)
    await page.waitForLoadState('networkidle')

    // Get conversation
    const convResp = await request.get(`${API_URL}/api/workspaces/${WORKSPACE_ID}/conversations`)
    const convs = await convResp.json()
    const convId = convs.timeline?.[0]?.id || convs.defaultChannelId

    if (!convId) {
      test.skip(true, 'No conversation')
      return
    }

    // Use page.evaluate to create WebSocket and listen
    const result = await page.evaluate((conversationId) => {
      return new Promise<{ messages: string[]; receivedBroadcast: boolean }>((resolve) => {
        const messages: string[] = []
        const wsUrl = `ws://${window.location.host}/ws/chat/01KNAA9P1KCXH6SXG555S3NABY`

        const ws = new WebSocket(wsUrl)

        ws.onopen = () => {
          messages.push('Connected')
        }

        ws.onmessage = (event) => {
          messages.push(`Received: ${event.data}`)
          // Try to parse
          try {
            const data = JSON.parse(event.data)
            if (data.type === 'new_message') {
              messages.push('Got new_message broadcast!')
            }
          } catch {}
        }

        // After 3 seconds, send a message via fetch
        setTimeout(async () => {
          try {
            const resp = await fetch(`/api/workspaces/01KNAA9P1KCXH6SXG555S3NABY/conversations/${conversationId}/messages`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                text: 'Direct WebSocket test ' + Date.now(),
                senderId: 'owner',
                senderName: 'Tester',
                timestamp: Date.now()
              })
            })
            messages.push(`Fetch response: ${resp.status}`)
          } catch (e) {
            messages.push(`Fetch error: ${e}`)
          }
        }, 3000)

        // Resolve after 10 seconds
        setTimeout(() => {
          ws.close()
          const receivedBroadcast = messages.some(m => m.includes('new_message'))
          resolve({ messages, receivedBroadcast })
        }, 10000)
      })
    }, convId)

    console.log('Test result:', result)
    await page.screenshot({ path: 'artifacts/ws-broadcast-test.png' })

    // Check if we received the broadcast
    console.log('Messages received:', result.messages)
    console.log('Received broadcast:', result.receivedBroadcast)
  })

  test('should verify internal API triggers broadcast', async ({ page, request }) => {
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_ID}`)
    await page.waitForLoadState('networkidle')

    // Get conversation
    const convResp = await request.get(`${API_URL}/api/workspaces/${WORKSPACE_ID}/conversations`)
    const convs = await convResp.json()
    const convId = convs.timeline?.[0]?.id || convs.defaultChannelId

    if (!convId) {
      test.skip(true, 'No conversation')
      return
    }

    // Create WebSocket in browser context
    const result = await page.evaluate((conversationId) => {
      return new Promise<{ messages: string[] }>((resolve) => {
        const messages: string[] = []
        const wsUrl = `ws://${window.location.host}/ws/chat/01KNAA9P1KCXH6SXG555S3NABY`

        const ws = new WebSocket(wsUrl)

        ws.onopen = () => {
          messages.push('WS Connected')
          // After connecting, trigger internal API
          setTimeout(async () => {
            // Use fetch to call internal API
            const resp = await fetch(`/api/internal/chat/send`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                workspaceId: '01KNAA9P1KCXH6SXG555S3NABY',
                conversationId: conversationId,
                memberId: '01KNAA9P1KCXH6SXG559R94TAR',
                content: 'Internal API broadcast test',
                timestamp: Date.now()
              })
            })
            messages.push(`Internal API: ${resp.status}`)
          }, 2000)
        }

        ws.onmessage = (event) => {
          messages.push(`WS Message: ${event.data}`)
        }

        ws.onerror = (e) => {
          messages.push(`WS Error: ${e}`)
        }

        setTimeout(() => {
          ws.close()
          resolve({ messages })
        }, 10000)
      })
    }, convId)

    console.log('Internal API test result:', result)
    await page.screenshot({ path: 'artifacts/internal-api-ws.png' })
  })
})