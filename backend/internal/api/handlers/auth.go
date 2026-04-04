package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/security"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/pkg/utils"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userRepo    repository.UserRepository
	jwtConfig   *security.JWTConfig
	authEnabled bool
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo repository.UserRepository, jwtConfig *security.JWTConfig, authEnabled bool) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		jwtConfig:   jwtConfig,
		authEnabled: authEnabled,
	}
}

// GetAuthConfig returns auth configuration status
// @Summary Get auth configuration
// @Description Returns whether authentication is enabled
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]bool
// @Router /api/auth/config [get]
func (h *AuthHandler) GetAuthConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"enabled": h.authEnabled,
	})
}

// Login handles user login
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.UserLogin true "Login credentials"
// @Success 200 {object} models.AuthToken
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	// If auth is disabled, return a mock successful response
	if !h.authEnabled {
		c.JSON(http.StatusOK, models.AuthToken{
			Token:     "disabled-auth-mode",
			ExpiresAt: time.Now().Add(365 * 24 * time.Hour).Unix(),
			User: &models.User{
				ID:       "default",
				Username: "anonymous",
			},
		})
		return
	}

	var req models.UserLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !security.VerifyPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Update last login
	_ = h.userRepo.UpdateLastLogin(c.Request.Context(), user.ID)

	// Generate JWT token
	token, expiresAt, err := h.jwtConfig.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.AuthToken{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	})
}

// Register handles user registration
// @Summary User registration
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.UserCreate true "Registration data"
// @Success 201 {object} models.AuthToken
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	// If auth is disabled, return error
	if !h.authEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "registration not available when auth is disabled"})
		return
	}

	var req models.UserCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if username already exists
	existing, err := h.userRepo.GetByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}

	// Hash password
	hash, err := security.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &models.User{
		ID:           utils.GenerateID(),
		Username:     req.Username,
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}

	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Generate token for immediate login
	token, expiresAt, err := h.jwtConfig.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, models.AuthToken{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	})
}

// ValidateToken checks if a token is valid
// @Summary Validate JWT token
// @Description Check if the provided JWT token is valid
// @Tags auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /api/auth/validate [post]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	// If auth is disabled, return valid
	if !h.authEnabled {
		c.JSON(http.StatusOK, gin.H{
			"valid":    true,
			"userId":   "default",
			"username": "anonymous",
		})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no token provided"})
		return
	}

	// Strip "Bearer " prefix
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	claims, err := h.jwtConfig.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":    true,
		"userId":   claims.UserID,
		"username": claims.Username,
	})
}

// GetCurrentUser returns the current authenticated user
// @Summary Get current user
// @Description Get the currently authenticated user's information
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]*models.User
// @Failure 401 {object} map[string]string
// @Router /api/auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// If auth is disabled, return anonymous user
	if !h.authEnabled {
		c.JSON(http.StatusOK, gin.H{
			"user": &models.User{
				ID:       "default",
				Username: "anonymous",
			},
		})
		return
	}

	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID.(string))
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}