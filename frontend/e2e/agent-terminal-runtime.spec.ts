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

test.describe.serial('agent terminal runtime', () => {
  let workspaceId = ''
  let memberId = ''
  let sessionId = ''

  test.beforeAll(async ({ request }) => {
    if (!(await backendAvailable(request))) {
      test.skip(true, `no backend at ${API_URL} (start backend or set ORCHESTRA_API_URL)`)
    }

    const workspace = await request.post(`${API_URL}/api/workspaces`, {
      data: {
        name: `E2E Terminal ${RUN_ID}`,
        path: process.cwd(),
        ownerDisplayName: 'E2E Owner',
      },
    })
    expect(workspace.ok()).toBeTruthy()
    workspaceId = (await workspace.json()).id

    const member = await request.post(`${API_URL}/api/workspaces/${workspaceId}/members`, {
      data: {
        name: `Shell Agent ${RUN_ID}`,
        roleType: 'assistant',
        terminalType: 'bash',
        terminalCommand: '/bin/bash',
        acpEnabled: true,
        acpCommand: '/bin/bash',
      },
    })
    expect(member.ok()).toBeTruthy()
    memberId = (await member.json()).id

    const session = await request.post(`${API_URL}/api/workspaces/${workspaceId}/members/${memberId}/terminal-session`, {
      data: {},
    })
    if (!session.ok()) {
      test.skip(true, `terminal session could not start: ${session.status()} ${await session.text()}`)
    }
    sessionId = (await session.json()).sessionId
  })

  test.afterAll(async ({ request }) => {
    if (sessionId) {
      await request.delete(`${API_URL}/api/terminals/${sessionId}`).catch(() => undefined)
    }
    if (workspaceId) {
      await request.delete(`${API_URL}/api/workspaces/${workspaceId}`).catch(() => undefined)
    }
  })

  test('renders tmux-backed session and accepts raw xterm input', async ({ page, request }) => {
    await page.goto(`/workspace/${workspaceId}/sessions`)
    await expect(page.getByText(`Shell Agent ${RUN_ID}`)).toBeVisible({ timeout: 10_000 })

    await page.getByRole('button', { name: '查看输出' }).click()
    await expect(page.locator('[data-test="terminal-surface"]')).toBeVisible({ timeout: 10_000 })

    const wsUrl = `${API_URL.replace(/^http/, 'ws')}/ws/terminal/${sessionId}?token=disabled-auth-mode`
    const marker = `ORCH_E2E_RAW_${RUN_ID}`
    const wsResult = await page.evaluate(
      async ({ url, command }) => {
        const ws = new WebSocket(url)
        await new Promise<void>((resolve, reject) => {
          ws.onopen = () => resolve()
          ws.onerror = () => reject(new Error('terminal websocket failed to open'))
          setTimeout(() => reject(new Error('terminal websocket open timeout')), 5000)
        })
        ws.send(JSON.stringify({ type: 'resize', cols: 100, rows: 30 }))
        ws.send(JSON.stringify({ type: 'raw_input', content: command }))
        ws.send(JSON.stringify({ type: 'raw_input', content: '\r' }))
        await new Promise((resolve) => setTimeout(resolve, 500))
        ws.close()
        return true
      },
      { url: wsUrl, command: `printf '${marker}\\n'` },
    )
    expect(wsResult).toBeTruthy()

    await expect
      .poll(
        async () => {
          const snapshot = await request.get(`${API_URL}/api/terminals/${sessionId}/snapshot?lines=80`)
          if (!snapshot.ok()) return ''
          return ((await snapshot.json()) as { content?: string }).content ?? ''
        },
        { timeout: 10_000 },
      )
      .toContain(marker)

    await page.getByRole('button', { name: '终止' }).click()
    await expect.poll(async () => {
      const sessions = await request.get(`${API_URL}/api/workspaces/${workspaceId}/terminal-sessions`)
      if (!sessions.ok()) return []
      return ((await sessions.json()) as { sessions?: unknown[] }).sessions ?? []
    }).toHaveLength(0)
    sessionId = ''
  })
})
