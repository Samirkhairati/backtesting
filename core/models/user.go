package models

import "time"

// User represents the users table in the database
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Never expose the password hash in JSON responses
	CreatedAt    time.Time `json:"created_at"`
}

// RegisterRequest represents the JSON payload for registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
}

// LoginRequest represents the JSON payload for logging in
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents the JSON response containing the JWT
type AuthResponse struct {
	Token string `json:"token"`
}
