import axios from 'axios'

/** Prefer server JSON `{ error: string }` over generic axios message. */
export function getApiErrorMessage(e: unknown): string {
  if (axios.isAxiosError(e)) {
    const data = e.response?.data as { error?: string } | undefined
    if (data?.error && typeof data.error === 'string') return data.error
    if (e.response?.status) {
      return `Request failed (${e.response.status})`
    }
  }
  if (e instanceof Error) return e.message
  return String(e)
}
