package models

import "time"

// APIKeyProvider represents supported AI providers
type APIKeyProvider string

const (
	ProviderAnthropic APIKeyProvider = "anthropic"
	ProviderOpenAI    APIKeyProvider = "openai"
	ProviderGoogle    APIKeyProvider = "google"
	ProviderCustom    APIKeyProvider = "custom"
)

// APIKey represents an encrypted API key stored in the database
type APIKey struct {
	ID           string         `json:"id"`
	Provider     APIKeyProvider `json:"provider"`
	EncryptedKey string         `json:"-"` // Never expose in JSON
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}

// APIKeyResponse is the response format for API keys (key is partially hidden)
type APIKeyResponse struct {
	ID           string         `json:"id"`
	Provider     APIKeyProvider `json:"provider"`
	KeyPreview   string         `json:"keyPreview"` // First 8 chars + "..." + last 4 chars
	IsValid      bool           `json:"isValid"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}

// APIKeyCreate is the request body for creating/updating an API key
type APIKeyCreate struct {
	Provider APIKeyProvider `json:"provider" binding:"required"`
	Key      string         `json:"key" binding:"required"`
}

// APIKeyTestRequest is the request body for testing an API key
type APIKeyTestRequest struct {
	Provider APIKeyProvider `json:"provider" binding:"required"`
	Key      string         `json:"key,omitempty"` // Optional: test a new key without saving
}

// APIKeyTestResult is the response for API key validation
type APIKeyTestResult struct {
	Valid    bool   `json:"valid"`
	Message  string `json:"message"`
	Provider string `json:"provider"`
}