import client from '@/shared/api/client'

export interface OutboxDiagnosticItem {
  id: string
  conversation_id: string
  sender_id: string
  content: string
  status: 'pending' | 'sending' | 'sent' | 'failed' | 'dead'
  attempt_count: number
  last_error: string
  workspace_id: string
  target_member_id: string
}

export async function fetchConversationDispatchDiagnostics(workspaceId: string, conversationId: string) {
  const response = await client.get(`/workspaces/${workspaceId}/outbox`, {
    params: {
      conversationId,
      limit: 20
    },
    skipErrorToast: true
  })
  return (response.data?.items || []) as OutboxDiagnosticItem[]
}
