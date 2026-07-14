import { expect, test, type APIRequestContext } from '@playwright/test'

const apiURL = process.env.ORCHESTRA_API_URL ?? 'http://127.0.0.1:18080'

interface WorkspaceFixture {
  id: string
  ownerId: string
  secretaryId: string
  conversationId: string
}

async function createWorkspace(request: APIRequestContext): Promise<WorkspaceFixture> {
  const name = `E2E workspace ${Date.now()}-${Math.random().toString(16).slice(2)}`
  const workspaceResponse = await request.post(`${apiURL}/api/workspaces`, {
    data: { name, path: '/tmp', ownerDisplayName: 'E2E Owner' }
  })
  expect(workspaceResponse.status()).toBe(201)
  const workspace = await workspaceResponse.json()

  const membersResponse = await request.get(`${apiURL}/api/workspaces/${workspace.id}/members`)
  expect(membersResponse.ok()).toBeTruthy()
  const members = await membersResponse.json()
  const owner = members.find((member: { roleType: string }) => member.roleType === 'owner')
  expect(owner).toBeTruthy()

  const secretaryResponse = await request.post(`${apiURL}/api/workspaces/${workspace.id}/members`, {
    data: { name: 'E2E Secretary', roleType: 'secretary' }
  })
  expect(secretaryResponse.status()).toBe(201)
  const secretary = await secretaryResponse.json()

  const conversationsResponse = await request.get(`${apiURL}/api/workspaces/${workspace.id}/conversations`)
  expect(conversationsResponse.ok()).toBeTruthy()
  const conversations = await conversationsResponse.json()
  expect(conversations.defaultChannelId).toBeTruthy()

  return {
    id: workspace.id,
    ownerId: owner.id,
    secretaryId: secretary.id,
    conversationId: conversations.defaultChannelId
  }
}

test('runs a workspace through chat and task updates over the real WebSocket', async ({ page, request }) => {
  const workspace = await createWorkspace(request)
  const chatSocket = page.waitForEvent('websocket', socket => socket.url().includes(`/ws/chat/${workspace.id}`))

  await page.goto(`/workspace/${workspace.id}/chat`)
  await chatSocket
  await expect(page.getByRole('heading', { name: /E2E workspace/ })).toBeVisible()

  const messageText = `WebSocket message ${Date.now()}`
  const messageResponse = await request.post(
    `${apiURL}/api/workspaces/${workspace.id}/conversations/${workspace.conversationId}/messages`,
    { data: { text: messageText, senderId: workspace.ownerId, senderName: 'E2E Owner' } }
  )
  expect(messageResponse.status()).toBe(201)
  await expect(page.getByText(messageText, { exact: true })).toBeVisible()

  const taskSocket = page.waitForEvent('websocket', socket => socket.url().includes(`/ws/chat/${workspace.id}`))
  await page.goto(`/workspace/${workspace.id}/tasks`)
  await taskSocket
  await expect(page.getByText('任务管理', { exact: true })).toBeVisible()
  await expect(page.getByText('暂无任务', { exact: true })).toBeVisible()

  const taskTitle = `Task event ${Date.now()}`
  const taskResponse = await request.post(`${apiURL}/api/internal/tasks/create`, {
    data: {
      workspaceId: workspace.id,
      conversationId: workspace.conversationId,
      secretaryId: workspace.secretaryId,
      title: taskTitle,
      description: 'created by the end-to-end test'
    }
  })
  expect(taskResponse.status()).toBe(201)
  await expect(page.getByText(taskTitle, { exact: true })).toBeVisible()
})
