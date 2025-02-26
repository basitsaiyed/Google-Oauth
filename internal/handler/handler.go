package handler

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

// Handler struct manages OAuth2 authentication and database interactions.
type Handler struct {
	oauthConfig *oauth2.Config // OAuth2 configuration for Google authentication
	DB          *gorm.DB      // Database connection instance
}

// NewHandler initializes a new Handler with OAuth2 configuration and database connection.
//
// It retrieves OAuth2 credentials from environment variables and sets up required scopes for authentication.
//
// Parameters:
//   - db: A pointer to a gorm.DB instance for database interactions.
//
// Returns:
//   - A pointer to a Handler instance with OAuth2 configuration and database connection.
func NewHandler(db *gorm.DB) *Handler {
	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),        // Google OAuth Client ID
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),    // Google OAuth Client Secret
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),     // OAuth Redirect URL

		Scopes: []string{
			"openid",  // Required for obtaining ID Token
			"email",   // Access to user's email
			"profile", // Access to user's profile information
			"https://www.googleapis.com/auth/calendar",         // Full access to Google Calendar
			"https://www.googleapis.com/auth/calendar.events",  // Manage calendar events
		},
		Endpoint: google.Endpoint,
	}

	return &Handler{
		oauthConfig: config,
		DB:          db,
	}
}
