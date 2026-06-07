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

test.describe.serial('mvp dispatch diagnostics flow', () => {
  let workspaceId = ''
  let ownerId = ''
  let assistantId = ''
  let conversationId = ''
  const assistantName = `Broken Dispatch ${RUN_ID}`

  test.beforeAll(async ({ request }) => {
    if (!(await backendAvailable(request))) {
      test.skip(true, `no backend at ${API_URL} (start backend or set ORCHESTRA_API_URL)`)
    }

    const workspace = await request.post(`${API_URL}/api/workspaces`, {
      data: {
        name: `E2E Dispatch Diagnostics ${RUN_ID}`,
        path: process.cwd(),
        ownerDisplayName: 'E2E Owner',
      },
    })
    expect(workspace.ok()).toBeTruthy()
    workspaceId = ((await workspace.json()) as { id: string }).id
    ownerId = await getOwnerMemberId(request, workspaceId)

    const assistant = await request.post(`${API_URL}/api/workspaces/${workspaceId}/members`, {
      data: {
        name: assistantName,
        roleType: 'assistant',
        terminalType: 'native',
        terminalCommand: '/bin/definitely-missing-orchestra-agent',
        acpEnabled: true,
        acpCommand: '/bin/definitely-missing-orchestra-agent',
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

  test('surfaces failed agent dispatch diagnostics in chat without a page refresh', async ({ page, request }) => {
    await page.goto(`/workspace/${workspaceId}/chat`)

    await expect(page.locator('.workspace-title')).toHaveText(`E2E Dispatch Diagnostics ${RUN_ID}`, { timeout: 15_000 })
    await expect(page.getByRole('heading', { name: /general/i })).toBeVisible({ timeout: 15_000 })
    await expect(page.getByText(assistantName)).toBeVisible({ timeout: 15_000 })
    await expect(page.locator('.connection-status.connected')).toBeVisible({ timeout: 15_000 })

    const messageText = `Trigger dispatch diagnostics ${RUN_ID}`
    const input = page.locator('textarea.chat-textarea')
    await input.fill(`@${assistantName} ${messageText}`)

    const mentionItem = page.locator('.mention-item', { hasText: assistantName }).first()
    if (await mentionItem.isVisible({ timeout: 2_000 }).catch(() => false)) {
      await mentionItem.click()
      await input.fill(`@${assistantName} ${messageText}`)
    }

    await page.locator('button.send-btn').click()
    await expect(input).toHaveValue('', { timeout: 10_000 })
    await expect(page.locator('.message-text', { hasText: messageText })).toBeVisible({ timeout: 15_000 })

    await expect
      .poll(
        async () => {
          const outbox = await request.get(`${API_URL}/api/workspaces/${workspaceId}/outbox?conversationId=${conversationId}`)
          if (!outbox.ok()) return ''
          const body = (await outbox.json()) as { items?: Array<{ status?: string; target_member_id?: string }> }
          return (body.items ?? [])
            .filter((item) => item.target_member_id === assistantId)
            .map((item) => item.status ?? '')
            .join(',')
        },
        { timeout: 20_000 },
      )
      .toMatch(/failed|dead/)

    await expect(page.locator('.dispatch-warning', { hasText: 'Delivery issue' })).toBeVisible({ timeout: 15_000 })
  })
})
