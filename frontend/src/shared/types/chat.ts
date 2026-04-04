export type ConversationType = 'channel' | 'dm'
export type MessageStatus = 'sending' | 'sent' | 'failed'

export interface MessageMentionsPayload {
  mentionIds: string[]
  mentionAll: boolean
}

export interface MessageAttachment {
  type: 'image' | 'roadmap'
  filePath?: string
  fileName?: string
  fileSize?: number
  mimeType?: string
  width?: number
  height?: number
  thumbnailPath?: string
  title?: string
}

export interface MessageContent {
  type: 'text' | 'system'
  text?: string
  key?: string
  args?: Record<string, string>
}

export interface Message {
  id: string
  senderId?: string
  senderName: string
  senderAvatar?: string
  content: MessageContent
  createdAt: number
  isAi: boolean
  attachment?: MessageAttachment
  status?: MessageStatus
}

export interface Conversation {
  id: string
  type: ConversationType
  memberIds: string[]
  targetId?: string
  nameKey?: string
  customName?: string
  descriptionKey?: string
  pinned: boolean
  muted: boolean
  lastMessageAt?: number
  lastMessagePreview?: string
  lastMessageSenderId?: string
  lastMessageSenderName?: string
  lastMessageSenderAvatar?: string
  lastMessageAttachment?: MessageAttachment
  isDefault?: boolean
  unreadCount?: number
  messages: Message[]
}

export interface ConversationDto {
  id: string
  type: ConversationType
  memberIds?: string[]
  targetId?: string
  customName?: string
  pinned?: boolean
  muted?: boolean
  lastMessageAt?: number
  lastMessagePreview?: string
  isDefault?: boolean
  unreadCount?: number
}

export interface MessageDto {
  id: string
  senderId?: string
  content: MessageContent
  createdAt: number
  isAi: boolean
  attachment?: MessageAttachment
  status?: MessageStatus
}

export interface ChatMessageCreatedPayload {
  workspaceId?: string
  conversationId: string
  spanId?: string
  message: MessageDto
}

export interface ChatMessageStatusPayload {
  workspaceId?: string
  conversationId: string
  messageId: string
  status: MessageStatus
}

export interface ChatUnreadSyncPayload {
  workspaceId?: string
  conversationId?: string
  conversationUnreadCount?: number
  totalUnreadCount?: number
  resetAll?: boolean
}

export interface ChatDispatchMentions {
  mentionIds: string[]
  mentionAll: boolean
}

export interface ChatListResponse {
  pinned: ConversationDto[]
  timeline: ConversationDto[]
  defaultChannelId?: string
  totalUnreadCount?: number
}