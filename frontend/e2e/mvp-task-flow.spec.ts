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

test.describe.serial('mvp task flow', () => {
  let workspaceId = ''
  let secretaryId = ''
  let assistantId = ''
  let conversationId = ''
  let taskId = ''

  test.beforeAll(async ({ request }) => {
    if (!(await backendAvailable(request))) {
      test.skip(true, `no backend at ${API_URL} (start backend or set ORCHESTRA_API_URL)`)
    }

    const workspace = await request.post(`${API_URL}/api/workspaces`, {
      data: {
        name: `E2E Tasks ${RUN_ID}`,
        path: process.cwd(),
        ownerDisplayName: 'E2E Owner',
      },
    })
    expect(workspace.ok()).toBeTruthy()
    workspaceId = ((await workspace.json()) as { id: string }).id

    const secretary = await request.post(`${API_URL}/api/workspaces/${workspaceId}/members`, {
      data: {
        name: `Task Secretary ${RUN_ID}`,
        roleType: 'secretary',
        terminalType: 'bash',
        terminalCommand: '/bin/bash',
        acpEnabled: true,
        acpCommand: '/bin/bash',
      },
    })
    expect(secretary.ok()).toBeTruthy()
    secretaryId = ((await secretary.json()) as { id: string }).id

    const assistant = await request.post(`${API_URL}/api/workspaces/${workspaceId}/members`, {
      data: {
        name: `Task Assistant ${RUN_ID}`,
        roleType: 'assistant',
        terminalType: 'bash',
        terminalCommand: '/bin/bash',
        acpEnabled: true,
        acpCommand: '/bin/bash',
      },
    })
    expect(assistant.ok()).toBeTruthy()
    assistantId = ((await assistant.json()) as { id: string }).id

    const conversations = await request.get(`${API_URL}/api/workspaces/${workspaceId}/conversations`)
    expect(conversations.ok()).toBeTruthy()
    const conversationsBody = (await conversations.json()) as { defaultChannelId?: string }
    expect(conversationsBody.defaultChannelId).toBeTruthy()
    conversationId = conversationsBody.defaultChannelId!

    const task = await request.post(`${API_URL}/api/internal/tasks/create`, {
      data: {
        workspaceId,
        conversationId,
        secretaryId,
        title: `Browser task ${RUN_ID}`,
        description: `Created for the MVP task browser flow ${RUN_ID}`,
        assigneeId: assistantId,
      },
    })
    expect(task.ok()).toBeTruthy()
    taskId = ((await task.json()) as { taskId: string }).taskId
  })

  test.afterAll(async ({ request }) => {
    if (workspaceId) {
      await request.delete(`${API_URL}/api/workspaces/${workspaceId}`).catch(() => undefined)
    }
  })

  test('loads assigned tasks and completes one through the browser UI', async ({ page, request }) => {
    await page.goto(`/workspace/${workspaceId}/tasks`)

    await expect(page.getByRole('heading', { name: '任务管理' })).toBeVisible({ timeout: 15_000 })
    await expect(page.locator('.page-subtitle')).toContainText(`E2E Tasks ${RUN_ID}`, { timeout: 15_000 })

    const taskCard = page.locator('.task-card', { hasText: `Browser task ${RUN_ID}` }).first()
    await expect(taskCard).toBeVisible({ timeout: 15_000 })
    await expect(taskCard.locator('.status-badge')).toHaveText('已分配')

    await taskCard.getByRole('button', { name: '开始' }).click()
    await expect(taskCard.locator('.status-badge')).toHaveText('进行中', { timeout: 10_000 })

    const resultSummary = `Browser completed ${RUN_ID}`
    await taskCard.getByRole('button', { name: '完成' }).click()
    await expect(page.getByRole('heading', { name: '完成任务' })).toBeVisible({ timeout: 10_000 })
    await page.locator('textarea.input-textarea').fill(resultSummary)
    await page.getByRole('button', { name: '标记完成' }).click()

    await expect(taskCard.locator('.status-badge')).toHaveText('已完成', { timeout: 10_000 })
    await expect(taskCard.locator('.task-result')).toContainText(resultSummary)

    const taskResp = await request.get(`${API_URL}/api/workspaces/${workspaceId}/tasks/${taskId}`)
    expect(taskResp.ok()).toBeTruthy()
    const taskBody = (await taskResp.json()) as { task?: { status?: string; resultSummary?: string } }
    expect(taskBody.task?.status).toBe('completed')
    expect(taskBody.task?.resultSummary).toBe(resultSummary)
  })
})
