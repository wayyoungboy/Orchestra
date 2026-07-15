import client from './client'

export interface TerminalSession {
  sessionId: string
  pid: number
}

export const terminalApi = {
  getOrCreate: (workspaceId: string, memberId: string) =>
    client.post<TerminalSession>(`/workspaces/${workspaceId}/members/${memberId}/terminal-session`),

  close: (sessionId: string) => client.delete(`/terminals/${sessionId}`),
}
