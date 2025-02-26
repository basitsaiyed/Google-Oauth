package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents an authenticated user in the system.
// It stores authentication details for Google OAuth integration.
type User struct {
	gorm.Model
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"` // Unique User ID (UUID)
	GoogleID     string    `gorm:"unique" json:"google_id"`                                   // Google OAuth ID (Unique)
	Email        string    `gorm:"unique" json:"email"`                                       // User's email (Unique)
	Name         string    `json:"name"`                                                      // User's full name
	Picture      string    `json:"picture"`                                                   // Profile picture URL
	AccessToken  string    `json:"access_token"`                                             // OAuth access token for API calls
	RefreshToken string    `json:"refresh_token"`                                            // OAuth refresh token to renew access
	ExpiresAt    time.Time `json:"expires_at"`                                               // Token expiration timestamp
}
