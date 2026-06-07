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

async function getOwnerMemberId(request: APIRequestContext, workspaceId: string) {
  const members = await request.get(`${API_URL}/api/workspaces/${workspaceId}/members`)
  expect(members.ok()).toBeTruthy()
  const body = (await members.json()) as Array<{ id: string; roleType: string }>
  const owner = body.find((member) => member.roleType === 'owner')
  expect(owner?.id).toBeTruthy()
  return owner!.id
}

test.describe.serial('mvp chat flow', () => {
  let workspaceId = ''
  let ownerId = ''
  let assistantId = ''
  let conversationId = ''

  test.beforeAll(async ({ request }) => {
    if (!(await backendAvailable(request))) {
      test.skip(true, `no backend at ${API_URL} (start backend or set ORCHESTRA_API_URL)`)
    }

    const workspace = await request.post(`${API_URL}/api/workspaces`, {
      data: {
        name: `E2E Chat ${RUN_ID}`,
        path: process.cwd(),
        ownerDisplayName: 'E2E Owner',
      },
    })
    expect(workspace.ok()).toBeTruthy()
    const workspaceBody = (await workspace.json()) as { id: string }
    workspaceId = workspaceBody.id
    ownerId = await getOwnerMemberId(request, workspaceId)

    const assistant = await request.post(`${API_URL}/api/workspaces/${workspaceId}/members`, {
      data: {
        name: `Chat Assistant ${RUN_ID}`,
        roleType: 'assistant',
        terminalType: 'bash',
        terminalCommand: '/bin/bash',
        acpEnabled: true,
        acpCommand: '/bin/bash',
      },
    })
    expect(assistant.ok()).toBeTruthy()
    assistantId = ((await assistant.json()) as { id: string }).id

    const conversations = await request.get(`${API_URL}/api/workspaces/${workspaceId}/conversations?userId=${ownerId}`)
    expect(conversations.ok()).toBeTruthy()
    const conversationsBody = (await conversations.json()) as { defaultChannelId?: string }
    expect(conversationsBody.defaultChannelId).toBeTruthy()
    conversationId = conversationsBody.defaultChannelId!
  })

  test.afterAll(async ({ request }) => {
    if (workspaceId) {
      await request.delete(`${API_URL}/api/workspaces/${workspaceId}`).catch(() => undefined)
    }
  })

  test('creates workspace members, enters chat, sends a page message, and persists it', async ({ page, request }) => {
    await page.goto(`/workspace/${workspaceId}/chat`)

    await expect(page.locator('.workspace-title')).toHaveText(`E2E Chat ${RUN_ID}`, { timeout: 15_000 })
    await expect(page.getByRole('heading', { name: /general/i })).toBeVisible({ timeout: 15_000 })
    await expect(page.getByRole('button', { name: /general/i })).toBeVisible({ timeout: 15_000 })
    await expect(page.getByText(`Chat Assistant ${RUN_ID}`)).toBeVisible({ timeout: 15_000 })
    await expect(page.locator('.connection-status.connected')).toBeVisible({ timeout: 15_000 })

    const messageText = `MVP chat browser flow ${RUN_ID}`
    const mentionText = `@Chat Assistant ${RUN_ID}`
    const input = page.locator('textarea.chat-textarea')
    await input.fill(`${mentionText} ${messageText}`)

    const mentionItem = page.locator('.mention-item', { hasText: `Chat Assistant ${RUN_ID}` }).first()
    if (await mentionItem.isVisible({ timeout: 2_000 }).catch(() => false)) {
      await mentionItem.click()
      await input.fill(`${mentionText} ${messageText}`)
    }

    await page.locator('button.send-btn').click()
    await expect(input).toHaveValue('', { timeout: 10_000 })
    await expect(page.locator('.message-text', { hasText: messageText })).toBeVisible({ timeout: 15_000 })

    await expect
      .poll(
        async () => {
          const messages = await request.get(`${API_URL}/api/workspaces/${workspaceId}/conversations/${conversationId}/messages?limit=20`)
          if (!messages.ok()) return ''
          const body = (await messages.json()) as Array<{ content?: { text?: string } }>
          return body.map((message) => message.content?.text ?? '').join('\n')
        },
        { timeout: 15_000 },
      )
      .toContain(messageText)

    const outbox = await request.get(`${API_URL}/api/workspaces/${workspaceId}/outbox?status=failed`)
    expect(outbox.ok()).toBeTruthy()
    const outboxBody = (await outbox.json()) as { items?: unknown[] }
    expect(outboxBody.items ?? []).toHaveLength(0)
    expect(assistantId).toBeTruthy()
  })
})
