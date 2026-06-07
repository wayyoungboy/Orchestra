import { expect, test, type APIRequestContext } from '@playwright/test'

declare const process: {
  cwd(): string
  env: Record<string, string | undefined>
}

const API_URL = process.env.ORCHESTRA_API_URL ?? 'http://127.0.0.1:8080'
const RUN_ID = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`

async function backendAvailable(request: APIRequestContext) {
  const result = await request.get(`${API_URL}/health`, { timeout: 2_000 }).catch((err: Error) => err)
  if (result instanceof Error) return false
  return result.ok()
}

test.describe.serial('mvp unread sync flow', () => {
  let workspaceId = ''
  let ownerId = ''
  let assistantId = ''
  let defaultChannelId = ''
  let alertChannelId = ''
  const alertChannelName = `alerts-${RUN_ID}`

  test.beforeAll(async ({ request }) => {
    if (!(await backendAvailable(request))) {
      test.skip(true, `no backend at ${API_URL} (start backend or set ORCHESTRA_API_URL)`)
    }

    const workspace = await request.post(`${API_URL}/api/workspaces`, {
      data: {
        name: `E2E Unread ${RUN_ID}`,
        path: process.cwd(),
        ownerDisplayName: 'E2E Owner',
      },
    })
    expect(workspace.ok()).toBeTruthy()
    const workspaceBody = (await workspace.json()) as { id: string; ownerMemberId?: string }
    workspaceId = workspaceBody.id
    ownerId = workspaceBody.ownerMemberId ?? 'default'

    const assistant = await request.post(`${API_URL}/api/workspaces/${workspaceId}/members`, {
      data: {
        name: `Unread Assistant ${RUN_ID}`,
        roleType: 'assistant',
        terminalType: 'native',
        terminalCommand: '/bin/cat',
        acpEnabled: true,
        acpCommand: '/bin/cat',
      },
    })
    expect(assistant.ok()).toBeTruthy()
    assistantId = ((await assistant.json()) as { id: string }).id

    const conversations = await request.get(`${API_URL}/api/workspaces/${workspaceId}/conversations?userId=${ownerId}`)
    expect(conversations.ok()).toBeTruthy()
    const conversationsBody = (await conversations.json()) as { defaultChannelId?: string }
    expect(conversationsBody.defaultChannelId).toBeTruthy()
    defaultChannelId = conversationsBody.defaultChannelId!

    const alertChannel = await request.post(`${API_URL}/api/workspaces/${workspaceId}/conversations`, {
      data: {
        type: 'channel',
        name: alertChannelName,
        memberIDs: [ownerId, assistantId],
      },
    })
    expect(alertChannel.ok()).toBeTruthy()
    alertChannelId = ((await alertChannel.json()) as { id: string }).id
  })

  test.afterAll(async ({ request }) => {
    if (workspaceId) {
      await request.delete(`${API_URL}/api/workspaces/${workspaceId}`).catch(() => undefined)
    }
  })

  test('updates an inactive channel unread badge over WebSocket and clears it when opened', async ({ page, request }) => {
    await page.goto(`/workspace/${workspaceId}/chat`)
    await expect(page.locator('.workspace-title')).toHaveText(`E2E Unread ${RUN_ID}`, { timeout: 15_000 })
    await expect(page.getByRole('heading', { name: /general/i })).toBeVisible({ timeout: 15_000 })
    await expect(page.locator('.connection-status.connected')).toBeVisible({ timeout: 15_000 })

    const alertItem = page.locator('.channel-item', { hasText: alertChannelName })
    await expect(alertItem).toBeVisible({ timeout: 15_000 })
    await expect(alertItem.locator('.unread-badge')).toBeHidden()

    const replyText = `Unread assistant update ${RUN_ID}`
    const reply = await request.post(`${API_URL}/api/internal/chat/send`, {
      data: {
        workspaceId,
        conversationId: alertChannelId,
        senderId: assistantId,
        senderName: `Unread Assistant ${RUN_ID}`,
        text: replyText,
      },
    })
    expect(reply.ok()).toBeTruthy()

    await expect(alertItem.locator('.unread-badge')).toHaveText('1', { timeout: 15_000 })

    await expect
      .poll(
        async () => {
          const conversations = await request.get(`${API_URL}/api/workspaces/${workspaceId}/conversations?userId=${ownerId}`)
          if (!conversations.ok()) return -1
          const body = (await conversations.json()) as { timeline?: Array<{ id: string; unreadCount?: number }> }
          return body.timeline?.find((conversation) => conversation.id === alertChannelId)?.unreadCount ?? -1
        },
        { timeout: 15_000 },
      )
      .toBe(1)

    await alertItem.click()
    await expect(page.getByRole('heading', { name: alertChannelName })).toBeVisible({ timeout: 15_000 })
    await expect(page.locator('.message-text', { hasText: replyText })).toBeVisible({ timeout: 15_000 })
    await expect(alertItem.locator('.unread-badge')).toBeHidden({ timeout: 15_000 })

    await expect
      .poll(
        async () => {
          const conversations = await request.get(`${API_URL}/api/workspaces/${workspaceId}/conversations?userId=${ownerId}`)
          if (!conversations.ok()) return -1
          const body = (await conversations.json()) as { timeline?: Array<{ id: string; unreadCount?: number }> }
          return body.timeline?.find((conversation) => conversation.id === alertChannelId)?.unreadCount ?? -1
        },
        { timeout: 15_000 },
      )
      .toBe(0)

    expect(defaultChannelId).toBeTruthy()
  })
})
