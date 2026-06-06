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

test.describe.serial('mvp member agent session flow', () => {
  let workspaceId = ''
  let assistantId = ''
  let sessionId = ''

  test.beforeAll(async ({ request }) => {
    if (!(await backendAvailable(request))) {
      test.skip(true, `no backend at ${API_URL} (start backend or set ORCHESTRA_API_URL)`)
    }

    const workspace = await request.post(`${API_URL}/api/workspaces`, {
      data: {
        name: `E2E Members ${RUN_ID}`,
        path: process.cwd(),
        ownerDisplayName: 'E2E Owner',
      },
    })
    expect(workspace.ok()).toBeTruthy()
    workspaceId = ((await workspace.json()) as { id: string }).id

    const assistant = await request.post(`${API_URL}/api/workspaces/${workspaceId}/members`, {
      data: {
        name: `Member Assistant ${RUN_ID}`,
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

  test('starts a configured assistant session from the member card', async ({ page, request }) => {
    await page.goto(`/workspace/${workspaceId}/members`)

    await expect(page.getByRole('heading', { name: '团队成员' })).toBeVisible({ timeout: 15_000 })
    await expect(page.locator('.page-subtitle')).toContainText(`E2E Members ${RUN_ID}`, { timeout: 15_000 })

    const memberCard = page.locator('.member-card', { hasText: `Member Assistant ${RUN_ID}` }).first()
    await expect(memberCard).toBeVisible({ timeout: 15_000 })
    await expect(memberCard.locator('.session-state')).toHaveText('未启动')

    await memberCard.getByRole('button', { name: '启动会话' }).click()
    await expect(memberCard.locator('.session-state')).toContainText('会话:', { timeout: 15_000 })

    await expect
      .poll(
        async () => {
          const session = await request.get(`${API_URL}/api/workspaces/${workspaceId}/members/${assistantId}/terminal-session`)
          if (!session.ok()) return ''
          return ((await session.json()) as { sessionId?: string }).sessionId ?? ''
        },
        { timeout: 15_000 },
      )
      .not.toBe('')

    const session = await request.get(`${API_URL}/api/workspaces/${workspaceId}/members/${assistantId}/terminal-session`)
    expect(session.ok()).toBeTruthy()
    sessionId = ((await session.json()) as { sessionId: string }).sessionId

    await page.goto(`/workspace/${workspaceId}/sessions`)
    await expect(page.getByText(`Member Assistant ${RUN_ID}`)).toBeVisible({ timeout: 15_000 })
  })
})
