package models

import "time"

// User represents an authenticated user
type User struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"` // Never expose in JSON
	CreatedAt    time.Time  `json:"createdAt"`
	LastLoginAt  *time.Time `json:"lastLoginAt,omitempty"`
}

// UserLogin is the request body for login
type UserLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserCreate is the request body for registration
type UserCreate struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

// AuthToken is the response for successful login
type AuthToken struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
	User      *User  `json:"user"`
}