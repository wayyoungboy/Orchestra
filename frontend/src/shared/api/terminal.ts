import client from './client'

export interface CreateSessionRequest {
  command?: string
  args?: string[]
  workspaceId?: string
  terminalType?: string
  memberId?: string
  memberName?: string // For introduction prompt to AI assistant
}

export interface CreateSessionResponse {
  sessionId: string
  pid: number
}

export interface WorkspaceTerminalSessionDto {
  memberId: string
  sessionId: string
  pid: number
}

export const terminalApi = {
  /** 若服务端池中仍有该成员的 PTY，返回其 sessionId（用于刷新页后重连）。 */
  async getSessionForMember(
    workspaceId: string,
    memberId: string
  ): Promise<CreateSessionResponse | null> {
    const response = await client.get<CreateSessionResponse>(
      `/workspaces/${workspaceId}/members/${memberId}/terminal-session`,
      {
        // 404 = no pooled PTY for this member yet; normal before first POST /terminals.
        // Treat as success so axios does not reject (avoids red console + error interceptor).
        validateStatus: (status) => status === 200 || status === 404
      }
    )
    if (response.status === 404) return null
    return response.data
  },

  async createSession(
    data: CreateSessionRequest,
    options?: { skipErrorToast?: boolean }
  ): Promise<CreateSessionResponse> {
    const response = await client.post<CreateSessionResponse>('/terminals', data, {
      skipErrorToast: options?.skipErrorToast
    })
    return response.data
  },

  async deleteSession(sessionId: string): Promise<void> {
    await client.delete(`/terminals/${sessionId}`)
  },

  /** 服务端池中当前工作区全部 PTY（用于轮询成员终端状态 REQ-303）。 */
  async listWorkspaceTerminalSessions(workspaceId: string): Promise<WorkspaceTerminalSessionDto[]> {
    const response = await client.get<{ sessions: WorkspaceTerminalSessionDto[] }>(
      `/workspaces/${workspaceId}/terminal-sessions`
    )
    return response.data.sessions ?? []
  }
}