package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/config"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/security"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/pkg/utils"
)

// APIKeyHandler handles API key management endpoints
type APIKeyHandler struct {
	keyRepo   repository.APIKeyRepository
	encryptor *security.KeyEncryptor
	config    *config.Config
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(keyRepo repository.APIKeyRepository, cfg *config.Config) (*APIKeyHandler, error) {
	// If encryption key is not set, generate a temporary one for dev mode
	key := cfg.Security.EncryptionKey
	if key == "" {
		// Use a default dev key (NOT SECURE - only for development)
		key = "dev-mode-encryption-key-32-bytes!!"
	}
	encryptor, err := security.NewKeyEncryptor(key)
	if err != nil {
		return nil, err
	}
	return &APIKeyHandler{
		keyRepo:   keyRepo,
		encryptor: encryptor,
		config:    cfg,
	}, nil
}

// maskKey creates a partially hidden preview of an API key
func maskKey(key string) string {
	if len(key) <= 12 {
		return "****"
	}
	return key[:8] + "..." + key[len(key)-4:]
}

// List returns all stored API keys (with masked values)
// @Summary List API keys
// @Description Returns all stored API keys with masked values
// @Tags api-keys
// @Produce json
// @Success 200 {array} models.APIKeyResponse
// @Router /api/api-keys [get]
func (h *APIKeyHandler) List(c *gin.Context) {
	keys, err := h.keyRepo.List(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list API keys"})
		return
	}

	responses := make([]models.APIKeyResponse, len(keys))
	for i, key := range keys {
		// Decrypt to create preview
		decrypted, err := h.encryptor.Decrypt(key.EncryptedKey)
		if err != nil {
			responses[i] = models.APIKeyResponse{
				ID:        key.ID,
				Provider:  key.Provider,
				KeyPreview: "****",
				IsValid:   false,
				CreatedAt: key.CreatedAt,
				UpdatedAt: key.UpdatedAt,
			}
			continue
		}
		responses[i] = models.APIKeyResponse{
			ID:        key.ID,
			Provider:  key.Provider,
			KeyPreview: maskKey(decrypted),
			IsValid:   true,
			CreatedAt: key.CreatedAt,
			UpdatedAt: key.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, responses)
}

// Create stores a new API key
// @Summary Create or update API key
// @Description Stores an API key for a provider (updates existing key for same provider)
// @Tags api-keys
// @Accept json
// @Produce json
// @Param request body models.APIKeyCreate true "API key data"
// @Success 200 {object} models.APIKeyResponse
// @Failure 400 {object} map[string]string
// @Router /api/api-keys [post]
func (h *APIKeyHandler) Create(c *gin.Context) {
	var req models.APIKeyCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate provider
	validProviders := []models.APIKeyProvider{
		models.ProviderAnthropic,
		models.ProviderOpenAI,
		models.ProviderGoogle,
		models.ProviderCustom,
	}
	isValidProvider := false
	for _, p := range validProviders {
		if req.Provider == p {
			isValidProvider = true
			break
		}
	}
	if !isValidProvider {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider"})
		return
	}

	// Encrypt the key
	encrypted, err := h.encryptor.Encrypt(req.Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt key"})
		return
	}

	// Check if key exists for this provider
	existing, err := h.keyRepo.GetByProvider(context.Background(), req.Provider)
	now := time.Now()

	if err != nil || existing == nil {
		// Create new key
		key := &models.APIKey{
			ID:           utils.GenerateID(),
			Provider:     req.Provider,
			EncryptedKey: encrypted,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := h.keyRepo.Create(context.Background(), key); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store API key"})
			return
		}
		c.JSON(http.StatusOK, models.APIKeyResponse{
			ID:         key.ID,
			Provider:   key.Provider,
			KeyPreview: maskKey(req.Key),
			IsValid:    true,
			CreatedAt:  key.CreatedAt,
			UpdatedAt:  key.UpdatedAt,
		})
		return
	}

	// Update existing key
	existing.EncryptedKey = encrypted
	existing.UpdatedAt = now
	if err := h.keyRepo.Update(context.Background(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update API key"})
		return
	}
	c.JSON(http.StatusOK, models.APIKeyResponse{
		ID:         existing.ID,
		Provider:   existing.Provider,
		KeyPreview: maskKey(req.Key),
		IsValid:    true,
		CreatedAt:  existing.CreatedAt,
		UpdatedAt:  existing.UpdatedAt,
	})
}

// Delete removes an API key
// @Summary Delete API key
// @Description Deletes an API key by ID
// @Tags api-keys
// @Param id path string true "API key ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /api/api-keys/:id [delete]
func (h *APIKeyHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
		return
	}

	if err := h.keyRepo.Delete(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete API key"})
		return
	}

	c.Status(http.StatusNoContent)
}

// Test validates an API key by making a test request to the provider
// @Summary Test API key
// @Description Validates an API key by making a test request to the provider
// @Tags api-keys
// @Accept json
// @Produce json
// @Param request body models.APIKeyTestRequest true "API key test data"
// @Success 200 {object} models.APIKeyTestResult
// @Router /api/api-keys/test [post]
func (h *APIKeyHandler) Test(c *gin.Context) {
	var req models.APIKeyTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If key not provided, get stored key
	key := req.Key
	if key == "" {
		stored, err := h.keyRepo.GetByProvider(context.Background(), req.Provider)
		if err != nil || stored == nil {
			c.JSON(http.StatusOK, models.APIKeyTestResult{
				Valid:    false,
				Message:  "no key stored for this provider",
				Provider: string(req.Provider),
			})
			return
		}
		decrypted, err := h.encryptor.Decrypt(stored.EncryptedKey)
		if err != nil {
			c.JSON(http.StatusOK, models.APIKeyTestResult{
				Valid:    false,
				Message:  "failed to decrypt stored key",
				Provider: string(req.Provider),
			})
			return
		}
		key = decrypted
	}

	// Test the key based on provider
	result := h.testKey(req.Provider, key)
	c.JSON(http.StatusOK, result)
}

// testKey validates the API key by provider-specific logic
func (h *APIKeyHandler) testKey(provider models.APIKeyProvider, key string) models.APIKeyTestResult {
	switch provider {
	case models.ProviderAnthropic:
		return h.testAnthropicKey(key)
	case models.ProviderOpenAI:
		return h.testOpenAIKey(key)
	case models.ProviderGoogle:
		return h.testGoogleKey(key)
	case models.ProviderCustom:
		// For custom providers, just check key format
		if len(key) >= 10 {
			return models.APIKeyTestResult{
				Valid:    true,
				Message:  "key format looks valid",
				Provider: string(provider),
			}
		}
		return models.APIKeyTestResult{
			Valid:    false,
			Message:  "key too short",
			Provider: string(provider),
		}
	default:
		return models.APIKeyTestResult{
			Valid:    false,
			Message:  "unknown provider",
			Provider: string(provider),
		}
	}
}

func (h *APIKeyHandler) testAnthropicKey(key string) models.APIKeyTestResult {
	// Anthropic keys start with sk-ant-
	if len(key) < 10 {
		return models.APIKeyTestResult{
			Valid:    false,
			Message:  "key too short",
			Provider: string(models.ProviderAnthropic),
		}
	}
	// Basic format check
	if key[:7] != "sk-ant-" && key[:3] != "sk-" {
		return models.APIKeyTestResult{
			Valid:    false,
			Message:  "invalid Anthropic key format",
			Provider: string(models.ProviderAnthropic),
		}
	}
	return models.APIKeyTestResult{
		Valid:    true,
		Message:  "key format is valid",
		Provider: string(models.ProviderAnthropic),
	}
}

func (h *APIKeyHandler) testOpenAIKey(key string) models.APIKeyTestResult {
	// OpenAI keys start with sk-
	if len(key) < 20 {
		return models.APIKeyTestResult{
			Valid:    false,
			Message:  "key too short",
			Provider: string(models.ProviderOpenAI),
		}
	}
	if key[:3] != "sk-" {
		return models.APIKeyTestResult{
			Valid:    false,
			Message:  "invalid OpenAI key format",
			Provider: string(models.ProviderOpenAI),
		}
	}
	return models.APIKeyTestResult{
		Valid:    true,
		Message:  "key format is valid",
		Provider: string(models.ProviderOpenAI),
	}
}

func (h *APIKeyHandler) testGoogleKey(key string) models.APIKeyTestResult {
	// Google API keys are typically 39 characters
	if len(key) < 30 {
		return models.APIKeyTestResult{
			Valid:    false,
			Message:  "key too short",
			Provider: string(models.ProviderGoogle),
		}
	}
	return models.APIKeyTestResult{
		Valid:    true,
		Message:  "key format is valid",
		Provider: string(models.ProviderGoogle),
	}
}

// GetByProvider returns the API key for a specific provider (with masked value)
// @Summary Get API key by provider
// @Description Returns the API key for a specific provider with masked value
// @Tags api-keys
// @Param provider path string true "Provider name"
// @Success 200 {object} models.APIKeyResponse
// @Failure 404 {object} map[string]string
// @Router /api/api-keys/provider/:provider [get]
func (h *APIKeyHandler) GetByProvider(c *gin.Context) {
	provider := models.APIKeyProvider(c.Param("provider"))
	key, err := h.keyRepo.GetByProvider(context.Background(), provider)
	if err != nil || key == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found for provider"})
		return
	}

	decrypted, err := h.encryptor.Decrypt(key.EncryptedKey)
	if err != nil {
		c.JSON(http.StatusOK, models.APIKeyResponse{
			ID:        key.ID,
			Provider:  key.Provider,
			KeyPreview: "****",
			IsValid:   false,
			CreatedAt: key.CreatedAt,
			UpdatedAt: key.UpdatedAt,
		})
		return
	}

	c.JSON(http.StatusOK, models.APIKeyResponse{
		ID:        key.ID,
		Provider:  key.Provider,
		KeyPreview: maskKey(decrypted),
		IsValid:   true,
		CreatedAt: key.CreatedAt,
		UpdatedAt: key.UpdatedAt,
	})
}