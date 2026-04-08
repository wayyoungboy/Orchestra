package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/pkg/utils"
)

// AttachmentHandler handles file attachment operations
type AttachmentHandler struct {
	msgRepo     *repository.MessageRepository
	convRepo    *repository.ConversationRepository
	attachRepo  *repository.AttachmentRepository
	uploadDir   string
	maxFileSize int64
}

// NewAttachmentHandler creates a new attachment handler
func NewAttachmentHandler(msgRepo *repository.MessageRepository, convRepo *repository.ConversationRepository, attachRepo *repository.AttachmentRepository, uploadDir string) *AttachmentHandler {
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	// Ensure upload directory exists
	os.MkdirAll(uploadDir, 0755)

	return &AttachmentHandler{
		msgRepo:     msgRepo,
		convRepo:    convRepo,
		attachRepo:  attachRepo,
		uploadDir:   uploadDir,
		maxFileSize: 50 * 1024 * 1024, // 50MB default
	}
}

// UploadAttachment handles file upload for a conversation
// @Summary Upload attachment
// @Description Upload a file attachment to a conversation
// @Tags attachments
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param convId path string true "Conversation ID"
// @Param file formData file true "File to upload"
// @Param senderId formData string true "Sender member ID"
// @Success 201 {object} models.AttachmentUpload
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/conversations/{convId}/attachments [post]
func (h *AttachmentHandler) UploadAttachment(c *gin.Context) {
	workspaceID := c.Param("id")
	convID := c.Param("convId")
	senderID := c.PostForm("senderId")

	if senderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "senderId is required"})
		return
	}

	// Verify conversation exists
	conv, err := h.convRepo.GetByID(convID)
	if err != nil || conv == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	// Check file size
	if header.Size > h.maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file size exceeds maximum allowed (%d bytes)", h.maxFileSize)})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	fileID := utils.GenerateID()
	fileName := fileID + ext

	// Create workspace-specific directory
	uploadPath := filepath.Join(h.uploadDir, workspaceID)
	os.MkdirAll(uploadPath, 0755)

	// Create file on disk
	dstPath := filepath.Join(uploadPath, fileName)
	dst, err := os.Create(dstPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	// Detect MIME type
	mimeType := detectMimeType(header.Filename, dstPath)
	isImage := strings.HasPrefix(mimeType, "image/")

	// Create message with attachment
	msg, err := h.msgRepo.Create(repository.MessageCreate{
		ConversationID: convID,
		SenderID:       senderID,
		Content: repository.MessageContent{
			Type: "attachment",
			Text: header.Filename,
		},
		IsAI: false,
	})
	if err != nil {
		// Clean up file if message creation fails
		os.Remove(dstPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create message"})
		return
	}

	// Create attachment record in database
	attachment := &models.Attachment{
		ID:          fileID,
		FileName:    header.Filename,
		FilePath:    dstPath,
		FileSize:    header.Size,
		MimeType:    mimeType,
		MessageID:   msg.ID,
		WorkspaceID: workspaceID,
		UploadedBy:  senderID,
		CreatedAt:   time.Now(),
	}

	if h.attachRepo != nil {
		if err := h.attachRepo.Create(c.Request.Context(), attachment); err != nil {
			// Log error but don't fail the upload - file is already saved
			fmt.Printf("Warning: failed to save attachment record: %v\n", err)
		}
	}

	// Build response
	url := fmt.Sprintf("/api/workspaces/%s/attachments/%s", workspaceID, fileID)

	response := models.AttachmentUpload{
		ID:       attachment.ID,
		FileName: attachment.FileName,
		FileSize: attachment.FileSize,
		MimeType: attachment.MimeType,
		URL:      url,
		IsImage:  isImage,
	}

	c.JSON(http.StatusCreated, gin.H{
		"attachment": response,
		"message": gin.H{
			"id":        msg.ID,
			"senderId":  msg.SenderID,
			"content":   msg.Content,
			"createdAt": msg.CreatedAt,
		},
	})
}

// DownloadAttachment downloads an attachment
// @Summary Download attachment
// @Description Download a file attachment
// @Tags attachments
// @Produce octet-stream
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param attachmentId path string true "Attachment ID"
// @Success 200 {file} binary
// @Failure 404 {object} map[string]string
// @Router /api/workspaces/{id}/attachments/{attachmentId} [get]
func (h *AttachmentHandler) DownloadAttachment(c *gin.Context) {
	workspaceID := c.Param("id")
	attachmentID := c.Param("attachmentId")

	// Try to get from database first
	if h.attachRepo != nil {
		attachment, err := h.attachRepo.GetByID(c.Request.Context(), attachmentID)
		if err == nil && attachment != nil {
			c.FileAttachment(attachment.FilePath, attachment.FileName)
			return
		}
	}

	// Fallback: Find file in upload directory
	uploadPath := filepath.Join(h.uploadDir, workspaceID)

	files, err := os.ReadDir(uploadPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "attachment not found"})
		return
	}

	var filePath string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), attachmentID) {
			filePath = filepath.Join(uploadPath, f.Name())
			break
		}
	}

	if filePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "attachment not found"})
		return
	}

	c.FileAttachment(filePath, filepath.Base(filePath))
}

// ListAttachments lists all attachments in a workspace
// @Summary List attachments
// @Description List all attachments in a workspace
// @Tags attachments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param conversationId query string false "Filter by conversation ID"
// @Success 200 {array} models.Attachment
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/attachments [get]
func (h *AttachmentHandler) ListAttachments(c *gin.Context) {
	workspaceID := c.Param("id")
	conversationID := c.Query("conversationId")

	if h.attachRepo == nil {
		c.JSON(http.StatusOK, []*models.Attachment{})
		return
	}

	var attachments []*models.Attachment
	var err error

	if conversationID != "" {
		attachments, err = h.attachRepo.ListByConversation(c.Request.Context(), conversationID)
	} else {
		attachments, err = h.attachRepo.ListByWorkspace(c.Request.Context(), workspaceID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if attachments == nil {
		attachments = []*models.Attachment{}
	}

	c.JSON(http.StatusOK, attachments)
}

// DeleteAttachment deletes an attachment
// @Summary Delete attachment
// @Description Delete a file attachment
// @Tags attachments
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param attachmentId path string true "Attachment ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/attachments/{attachmentId} [delete]
func (h *AttachmentHandler) DeleteAttachment(c *gin.Context) {
	attachmentID := c.Param("attachmentId")

	if h.attachRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "attachment repository not available"})
		return
	}

	// Get attachment to find file path
	attachment, err := h.attachRepo.GetByID(c.Request.Context(), attachmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "attachment not found"})
		return
	}

	// Delete from database
	if err := h.attachRepo.Delete(c.Request.Context(), attachmentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Delete file from disk
	if attachment.FilePath != "" {
		os.Remove(attachment.FilePath)
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetAttachmentInfo returns information about an attachment
// @Summary Get attachment info
// @Description Get attachment metadata
// @Tags attachments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param attachmentId path string true "Attachment ID"
// @Success 200 {object} models.Attachment
// @Failure 404 {object} map[string]string
// @Router /api/workspaces/{id}/attachments/{attachmentId}/info [get]
func (h *AttachmentHandler) GetAttachmentInfo(c *gin.Context) {
	attachmentID := c.Param("attachmentId")

	// Try to get from database first
	if h.attachRepo != nil {
		attachment, err := h.attachRepo.GetByID(c.Request.Context(), attachmentID)
		if err == nil && attachment != nil {
			c.JSON(http.StatusOK, attachment)
			return
		}
	}

	// Fallback: Get from file system
	workspaceID := c.Param("id")
	uploadPath := filepath.Join(h.uploadDir, workspaceID)

	files, err := os.ReadDir(uploadPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "attachment not found"})
		return
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name(), attachmentID) {
			filePath := filepath.Join(uploadPath, f.Name())
			info, err := f.Info()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get file info"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"id":        attachmentID,
				"fileName":  f.Name(),
				"fileSize":  info.Size(),
				"createdAt": info.ModTime().UnixMilli(),
				"mimeType":  detectMimeType(f.Name(), filePath),
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "attachment not found"})
}

// detectMimeType detects MIME type based on file extension
func detectMimeType(filename, filePath string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".zip":
		return "application/zip"
	case ".tar":
		return "application/x-tar"
	case ".gz":
		return "application/gzip"
	default:
		return "application/octet-stream"
	}
}