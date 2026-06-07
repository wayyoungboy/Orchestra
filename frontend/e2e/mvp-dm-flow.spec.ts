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

test.describe.serial('mvp direct message flow', () => {
  let workspaceId = ''
  let assistantId = ''
  let sessionId = ''

  test.beforeAll(async ({ request }) => {
    if (!(await backendAvailable(request))) {
      test.skip(true, `no backend at ${API_URL} (start backend or set ORCHESTRA_API_URL)`)
    }

    const workspace = await request.post(`${API_URL}/api/workspaces`, {
      data: {
        name: `E2E DM ${RUN_ID}`,
        path: process.cwd(),
        ownerDisplayName: 'E2E Owner',
      },
    })
    expect(workspace.ok()).toBeTruthy()
    workspaceId = ((await workspace.json()) as { id: string }).id

    const assistant = await request.post(`${API_URL}/api/workspaces/${workspaceId}/members`, {
      data: {
        name: `DM Assistant ${RUN_ID}`,
        roleType: 'assistant',
        terminalType: 'native',
        terminalCommand: '/bin/cat',
        acpEnabled: true,
        acpCommand: '/bin/cat',
      },
    })
    expect(assistant.ok()).toBeTruthy()
    assistantId = ((await assistant.json()) as { id: string }).id
  })

  test.afterAll(async ({ request }) => {
    if (sessionId) {
      await request.delete(`${API_URL}/api/terminals/${sessionId}`).catch(() => undefined)
    }
    if (workspaceId) {
      await request.delete(`${API_URL}/api/workspaces/${workspaceId}`).catch(() => undefined)
    }
  })

  test('creates a reusable DM, dispatches to the assistant, and shows the reply in the same DM', async ({ page, request }) => {
    await page.goto(`/workspace/${workspaceId}/chat`)
    await expect(page.locator('.workspace-title')).toHaveText(`E2E DM ${RUN_ID}`, { timeout: 15_000 })
    await expect(page.getByText(`DM Assistant ${RUN_ID}`)).toBeVisible({ timeout: 15_000 })

    await page.locator('button.add-btn').click()
    await page.getByRole('button', { name: /direct message/i }).click()
    await page.getByRole('button', { name: new RegExp(`DM Assistant ${RUN_ID}`) }).click()

    const directResponse = page.waitForResponse(
      (response) => response.url().includes('/api/workspaces/') &&
        response.url().includes('/conversations/direct') &&
        response.request().method() === 'POST',
    )
    await page.getByRole('button', { name: 'Create' }).click()
    const response = await directResponse
    expect([200, 201]).toContain(response.status())
    const dm = (await response.json()) as { id?: string; type?: string; memberIds?: string[]; targetId?: string }
    expect(dm.id).toBeTruthy()
    expect(dm.type).toBe('dm')
    expect(dm.memberIds ?? []).toContain(assistantId)
    const conversationId = dm.id!

    await expect(page.getByRole('heading', { name: `DM Assistant ${RUN_ID}` })).toBeVisible({ timeout: 15_000 })
    await expect(page.getByRole('button', { name: new RegExp(`DM Assistant ${RUN_ID}`) })).toBeVisible({ timeout: 15_000 })

    const userText = `DM dispatch browser flow ${RUN_ID}`
    const input = page.locator('textarea.chat-textarea')
    await input.fill(userText)
    await page.locator('button.send-btn').click()
    await expect(input).toHaveValue('', { timeout: 10_000 })
    await expect(page.locator('.message-text', { hasText: userText })).toBeVisible({ timeout: 15_000 })

    await expect
      .poll(
        async () => {
          const session = await request.get(`${API_URL}/api/workspaces/${workspaceId}/members/${assistantId}/terminal-session`)
          if (!session.ok()) return ''
          const body = (await session.json()) as { sessionId?: string }
          sessionId = body.sessionId ?? sessionId
          return sessionId
        },
        { timeout: 15_000 },
      )
      .not.toBe('')

    await expect
      .poll(
        async () => {
          const snapshot = await request.get(`${API_URL}/api/terminals/${sessionId}/snapshot?lines=120`)
          if (!snapshot.ok()) return ''
          return ((await snapshot.json()) as { content?: string }).content ?? ''
        },
        { timeout: 15_000 },
      )
      .toContain(userText)

    await expect
      .poll(
        async () => {
          const snapshot = await request.get(`${API_URL}/api/terminals/${sessionId}/snapshot?lines=120`)
          if (!snapshot.ok()) return ''
          return ((await snapshot.json()) as { content?: string }).content ?? ''
        },
        { timeout: 15_000 },
      )
      .toContain(`#conversationId{${conversationId}}`)

    const replyText = `DM assistant reply ${RUN_ID}`
    const reply = await request.post(`${API_URL}/api/internal/chat/send`, {
      data: {
        workspaceId,
        conversationId,
        senderId: assistantId,
        senderName: `DM Assistant ${RUN_ID}`,
        text: replyText,
      },
    })
    expect(reply.ok()).toBeTruthy()

    await expect(page.locator('.message-text', { hasText: replyText })).toBeVisible({ timeout: 15_000 })

    const messages = await request.get(`${API_URL}/api/workspaces/${workspaceId}/conversations/${conversationId}/messages?limit=20`)
    expect(messages.ok()).toBeTruthy()
    const messageTexts = ((await messages.json()) as Array<{ content?: { text?: string } }>)
      .map((message) => message.content?.text ?? '')
    expect(messageTexts).toContain(userText)
    expect(messageTexts).toContain(replyText)

    const outbox = await request.get(`${API_URL}/api/workspaces/${workspaceId}/outbox?status=failed`)
    expect(outbox.ok()).toBeTruthy()
    const outboxBody = (await outbox.json()) as { items?: unknown[] }
    expect(outboxBody.items ?? []).toHaveLength(0)
  })
})
