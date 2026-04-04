package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// Claims represents the JWT claims
type Claims struct {
	jwt.RegisteredClaims
	UserID   string `json:"userId"`
	Username string `json:"username"`
}

// NewJWTConfig creates a new JWT config
func NewJWTConfig(secret string) *JWTConfig {
	if secret == "" {
		secret = "default-secret-change-in-production"
	}
	return &JWTConfig{
		Secret:     secret,
		Expiration: 24 * time.Hour,
	}
}

// GenerateToken generates a JWT token for a user
func (c *JWTConfig) GenerateToken(userID, username string) (string, int64, error) {
	expiresAt := time.Now().Add(c.Expiration)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "orchestra",
		},
		UserID:   userID,
		Username: username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(c.Secret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}

// ValidateToken validates a JWT token and returns the claims
func (c *JWTConfig) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(c.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}