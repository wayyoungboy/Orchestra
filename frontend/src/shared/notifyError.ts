import { getApiErrorMessage } from '@/shared/api/errors'
import { useDiagnosticsStore } from '@/stores/diagnosticsStore'
import { useToastStore } from '@/stores/toastStore'

/**
 * User-visible + diagnostics + console. Use for any failure that must not be silent.
 * Axios errors are usually already reported via the client interceptor; callers may skip
 * duplicate reporting with shouldNotifyAxios:false when chaining the same error.
 */
export function notifyUserError(source: string, err: unknown, options?: { skipToast?: boolean }): void {
  const detail = getApiErrorMessage(err)
  const line = `[${source}] ${detail}`
  console.error(line, err)
  try {
    useDiagnosticsStore().log(line)
  } catch {
    /* Pinia not ready (tests / early boot) */
  }
  if (options?.skipToast) return
  try {
    useToastStore().pushToast(line, { tone: 'error', duration: 8000 })
  } catch {
    /* Pinia not ready */
  }
}
