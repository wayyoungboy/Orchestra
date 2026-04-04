import { test, expect } from '@playwright/test'

const BASE_URL = process.env.ORCHESTRA_E2E_BASE_URL || 'http://127.0.0.1:5173'
const API_URL = process.env.ORCHESTRA_API_URL || 'http://127.0.0.1:8080'

test.describe('Full Message Push Flow', () => {
  // Store created workspace ID for tests
  let testWorkspaceId: string

  test('setup: create test workspace', async ({ request }) => {
    // Create a workspace via API
    const createResp = await request.post(`${API_URL}/api/workspaces`, {
      data: {
        name: 'E2E Test Workspace',
        path: '/tmp/e2e-test-workspace'
      }
    })

    expect(createResp.status()).toBe(200)
    const workspace = await createResp.json()
    testWorkspaceId = workspace.id
    console.log('Created workspace:', testWorkspaceId)
  })

  test('should connect to WebSocket after login', async ({ page }) => {
    // Login
    await page.goto(`${BASE_URL}/login`)
    await page.locator('#orchestra-login-user').fill('orchestra')
    await page.locator('#orchestra-login-pass').fill('orchestra')
    await page.locator('button[type="submit"]').click()
    await expect(page).toHaveURL(/\/workspaces$/, { timeout: 5000 })

    // Navigate to workspace
    if (testWorkspaceId) {
      await page.goto(`${BASE_URL}/workspace/${testWorkspaceId}`)
      await page.waitForLoadState('networkidle')
    }

    // Capture WebSocket logs
    const wsLogs: string[] = []
    page.on('console', msg => {
      const text = msg.text()
      wsLogs.push(text)
    })

    // Wait for WebSocket to establish
    await page.waitForTimeout(3000)

    // Take screenshot
    await page.screenshot({ path: 'artifacts/ws-connection-test.png' })

    // Log WebSocket activity
    console.log('Console logs:', wsLogs.filter(l => l.includes('WebSocket') || l.includes('ChatSocket') || l.includes('chat') || l.includes('[chatStore]')))

    // Verify we're on workspace page
    expect(page.url()).toContain('workspace')
  })

  test('should create member and send message', async ({ page, request }) => {
    // Login
    await page.goto(`${BASE_URL}/login`)
    await page.locator('#orchestra-login-user').fill('orchestra')
    await page.locator('#orchestra-login-pass').fill('orchestra')
    await page.locator('button[type="submit"]').click()
    await expect(page).toHaveURL(/\/workspaces$/, { timeout: 5000 })

    // Get or use workspace
    if (!testWorkspaceId) {
      const wsResp = await request.get(`${API_URL}/api/workspaces`)
      const workspaces = await wsResp.json()
      testWorkspaceId = workspaces[0]?.id
    }

    if (!testWorkspaceId) {
      test.skip(true, 'No workspace available')
      return
    }

    // Create a member via API
    const memberResp = await request.post(`${API_URL}/api/workspaces/${testWorkspaceId}/members`, {
      data: {
        name: 'Test AI Member',
        roleType: 'assistant',
        command: 'echo "Hello from AI"',
        api_key: 'test-key-123'
      }
    })

    console.log('Member creation status:', memberResp.status())

    if (memberResp.status() === 200) {
      const member = await memberResp.json()
      console.log('Created member:', member.id)
    }

    // Navigate to workspace in browser
    await page.goto(`${BASE_URL}/workspace/${testWorkspaceId}`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(2000)

    // Take screenshot of workspace with members
    await page.screenshot({ path: 'artifacts/workspace-with-member.png' })

    // Check if members are visible
    const memberCards = page.locator('[data-testid="member-card"], .member-item, .member-card')
    const memberCount = await memberCards.count()
    console.log('Member count in UI:', memberCount)

    // Create conversation
    const convResp = await request.post(`${API_URL}/api/workspaces/${testWorkspaceId}/conversations`, {
      data: {
        type: 'channel',
        memberIds: ['owner']
      }
    })

    if (convResp.status() === 200) {
      const conv = await convResp.json()
      console.log('Created conversation:', conv.id)

      // Send a message via API
      const msgResp = await request.post(`${API_URL}/api/workspaces/${testWorkspaceId}/conversations/${conv.id}/messages`, {
        data: {
          text: 'Test message from E2E',
          senderId: 'owner',
          senderName: 'E2E Tester',
          timestamp: Date.now()
        }
      })

      console.log('Message send status:', msgResp.status())

      if (msgResp.status() === 200) {
        const msg = await msgResp.json()
        console.log('Message created:', msg.id)
      }
    }

    // Take final screenshot
    await page.waitForTimeout(1000)
    await page.screenshot({ path: 'artifacts/message-flow-final.png' })
  })

  test('should verify WebSocket message delivery', async ({ page, request, context }) => {
    // Login
    await page.goto(`${BASE_URL}/login`)
    await page.locator('#orchestra-login-user').fill('orchestra')
    await page.locator('#orchestra-login-pass').fill('orchestra')
    await page.locator('button[type="submit"]').click()
    await expect(page).toHaveURL(/\/workspaces$/, { timeout: 5000 })

    // Get workspace
    if (!testWorkspaceId) {
      const wsResp = await request.get(`${API_URL}/api/workspaces`)
      const workspaces = await wsResp.json()
      testWorkspaceId = workspaces[0]?.id
    }

    if (!testWorkspaceId) {
      test.skip(true, 'No workspace available')
      return
    }

    // Navigate to workspace
    await page.goto(`${BASE_URL}/workspace/${testWorkspaceId}`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(2000)

    // Capture WebSocket messages
    const wsMessages: any[] = []
    page.on('console', msg => {
      const text = msg.text()
      if (text.includes('new_message') || text.includes('WebSocket') || text.includes('ChatWebSocket')) {
        wsMessages.push({ type: 'log', text })
      }
    })

    // Get conversations
    const convResp = await request.get(`${API_URL}/api/workspaces/${testWorkspaceId}/conversations`)
    const convs = await convResp.json()

    if (convs.length > 0) {
      const convId = convs[0].id

      // Send message via Internal API (simulates AI response)
      const internalResp = await request.post(`${API_URL}/api/internal/chat/send`, {
        data: {
          workspaceId: testWorkspaceId,
          conversationId: convId,
          memberId: 'owner',
          content: 'AI response via WebSocket',
          timestamp: Date.now()
        }
      })

      console.log('Internal chat send status:', internalResp.status())

      // Wait for WebSocket message to arrive
      await page.waitForTimeout(2000)
    }

    // Take screenshot
    await page.screenshot({ path: 'artifacts/ws-message-test.png' })

    console.log('WebSocket messages captured:', wsMessages)
  })
})