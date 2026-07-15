export interface TerminalEvent {
  type: string
  sessionId?: string
  content?: string
  tool_name?: string
  tool_input?: unknown
  tool_use_id?: string
  message?: string
  status?: string
  error?: string
  code?: number
  cost_usd?: number
  duration_ms?: number
}

export function terminalWebSocketURL(sessionId: string): string {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const url = new URL(`/ws/terminal/${encodeURIComponent(sessionId)}`, `${protocol}//${window.location.host}`)
  const token = localStorage.getItem('orchestra.auth.token')
  if (token && token !== 'disabled-auth-mode') {
    url.searchParams.set('token', token)
  }
  return url.toString()
}
