import axios from 'axios'
import { notifyUserError } from '@/shared/notifyError'

const client = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

/** GET .../members/:id/terminal-session returns 404 when no server PTY yet; client then POST /terminals. Not an error. */
function isExpectedTerminalAttachProbe404(err: unknown): boolean {
  if (!axios.isAxiosError(err)) return false
  if (err.response?.status !== 404) return false
  if ((err.config?.method ?? '').toLowerCase() !== 'get') return false
  const u = err.config?.url ?? ''
  return /\/members\/[^/]+\/terminal-session\/?(\?|$)/.test(u) && !u.includes('terminal-sessions')
}

client.interceptors.response.use(
  (response) => response,
  (error) => {
    const ax = axios.isAxiosError(error) ? error : null
    const url = ax?.config?.url ?? ''
    const method = (ax?.config?.method ?? '').toUpperCase()
    const skipNoise = isExpectedTerminalAttachProbe404(error)

    const skipByConfig = !!(ax?.config as { skipErrorToast?: boolean } | undefined)?.skipErrorToast
    if (!skipNoise && !skipByConfig) {
      const label = `API ${method} ${url || '(no url)'}`
      notifyUserError(label, error)
    }
    return Promise.reject(error)
  }
)

export default client