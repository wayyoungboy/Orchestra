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

test.describe.serial('mvp notification flow', () => {
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
        name: `E2E Notify ${RUN_ID}`,
        path: process.cwd(),
        ownerDisplayName: 'E2E Owner',
      },
    })
    expect(workspace.ok()).toBeTruthy()
    workspaceId = ((await workspace.json()) as { id: string }).id
    ownerId = await getOwnerMemberId(request, workspaceId)

    const assistant = await request.post(`${API_URL}/api/workspaces/${workspaceId}/members`, {
      data: {
        name: `Notify Assistant ${RUN_ID}`,
        roleType: 'assistant',
        terminalType: 'native',
        terminalCommand: '/bin/cat',
        acpEnabled: true,
        acpCommand: '/bin/cat',
      },
    })
    expect(assistant.ok()).toBeTruthy()
    assistantId = ((await assistant.json()) as { id: string }).id

    const dm = await request.post(`${API_URL}/api/workspaces/${workspaceId}/conversations/direct`, {
      data: {
        userId: ownerId,
        targetId: assistantId,
      },
    })
    expect([200, 201]).toContain(dm.status())
    conversationId = ((await dm.json()) as { id: string }).id
  })

  test.afterAll(async ({ request }) => {
    if (workspaceId) {
      await request.delete(`${API_URL}/api/workspaces/${workspaceId}`).catch(() => undefined)
    }
  })

  test('shows an agent completion toast and persists notification counts without refresh', async ({ page, request }) => {
    await page.goto(`/workspace/${workspaceId}/chat`)
    await expect(page.locator('.workspace-title')).toHaveText(`E2E Notify ${RUN_ID}`, { timeout: 15_000 })
    await expect(page.locator('.connection-status.connected')).toBeVisible({ timeout: 15_000 })

    const reportText = `Notify assistant finished ${RUN_ID}`
    const reply = await request.post(`${API_URL}/api/internal/chat/send`, {
      data: {
        workspaceId,
        conversationId,
        senderId: assistantId,
        senderName: `Notify Assistant ${RUN_ID}`,
        text: reportText,
      },
    })
    expect(reply.ok()).toBeTruthy()

    await expect(page.locator('.toast-item', { hasText: reportText })).toBeVisible({ timeout: 15_000 })

    const badge = await request.get(`${API_URL}/api/workspaces/${workspaceId}/notifications/badge?userId=${ownerId}`)
    expect(badge.ok()).toBeTruthy()
    const badgeBody = (await badge.json()) as { unread?: number; total?: number }
    expect(badgeBody.unread).toBe(1)
    expect(badgeBody.total).toBe(1)

    const notifications = await request.get(`${API_URL}/api/workspaces/${workspaceId}/notifications?userId=${ownerId}`)
    expect(notifications.ok()).toBeTruthy()
    const list = (await notifications.json()) as Array<{ type?: string; body?: string; conversationId?: string; isRead?: boolean }>
    expect(list).toHaveLength(1)
    expect(list[0]).toMatchObject({
      type: 'agent_completion',
      body: reportText,
      conversationId,
      isRead: false,
    })
  })
})
