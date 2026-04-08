package models

import "time"

// Attachment represents a file attached to a message
type Attachment struct {
	ID          string    `json:"id"`
	FileName    string    `json:"fileName"`
	FilePath    string    `json:"filePath"`
	FileSize    int64     `json:"fileSize"`
	MimeType    string    `json:"mimeType"`
	MessageID   string    `json:"messageId,omitempty"`
	WorkspaceID string    `json:"workspaceId"`
	UploadedBy  string    `json:"uploadedBy"`
	CreatedAt   time.Time `json:"createdAt"`
}

// AttachmentUpload represents the response after uploading a file
type AttachmentUpload struct {
	ID        string `json:"id"`
	FileName  string `json:"fileName"`
	FileSize  int64  `json:"fileSize"`
	MimeType  string `json:"mimeType"`
	URL       string `json:"url"`
	Preview   string `json:"preview,omitempty"` // Base64 preview for images
	IsImage   bool   `json:"isImage"`
}