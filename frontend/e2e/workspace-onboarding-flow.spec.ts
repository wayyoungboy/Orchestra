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

test.describe.serial('workspace onboarding flow', () => {
  let workspaceId = ''

  test.beforeAll(async ({ request }) => {
    if (!(await backendAvailable(request))) {
      test.skip(true, `no backend at ${API_URL} (start backend or set ORCHESTRA_API_URL)`)
    }
  })

  test.afterAll(async ({ request }) => {
    if (workspaceId) {
      await request.delete(`${API_URL}/api/workspaces/${workspaceId}`).catch(() => undefined)
    }
  })

  test('creates a workspace from the first-screen UI and opens chat', async ({ page, request }) => {
    const workspaceName = `Onboarding Workspace ${RUN_ID}`
    const workspacePath = process.cwd()

    await page.goto('/')
    await expect(page.getByRole('heading', { name: /ready to/i })).toBeVisible({ timeout: 15_000 })
    await expect(page.getByText('RECENT WORKSPACES', { exact: true })).toBeVisible()

    await page.getByRole('button', { name: /create new workspace/i }).click()
    await expect(page.getByRole('heading', { name: '选择工作目录' })).toBeVisible({ timeout: 15_000 })

    const pathInput = page.getByPlaceholder('输入或粘贴服务器路径...')
    await pathInput.fill(workspacePath)
    await pathInput.press('Enter')

    const nameInput = page.getByPlaceholder('工作区显示名称...')
    await nameInput.fill(workspaceName)

    const createResponse = page.waitForResponse(
      (response) => response.url().includes('/api/workspaces') && response.request().method() === 'POST',
    )
    await page.getByRole('button', { name: '创建工作区' }).click()
    const response = await createResponse
    expect(response.ok()).toBeTruthy()
    const body = (await response.json()) as { id?: string }
    expect(body.id).toBeTruthy()
    workspaceId = body.id!

    await expect(page).toHaveURL(new RegExp(`/workspace/${workspaceId}/chat`), { timeout: 15_000 })
    await expect(page.getByRole('heading', { name: workspaceName })).toBeVisible({ timeout: 15_000 })
    await expect(page.getByRole('button', { name: /general/i })).toBeVisible({ timeout: 15_000 })

    const workspace = await request.get(`${API_URL}/api/workspaces/${workspaceId}`)
    expect(workspace.ok()).toBeTruthy()
    const persisted = (await workspace.json()) as { name?: string; path?: string }
    expect(persisted.name).toBe(workspaceName)
    expect(persisted.path).toBe(workspacePath)
  })
})
