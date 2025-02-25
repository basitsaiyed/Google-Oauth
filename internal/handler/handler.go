package handler

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

type Handler struct {
	oauthConfig *oauth2.Config
	DB          *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),

		Scopes: []string{
			"openid",  // Required for ID Token
			"email",   // Get user email
			"profile", // Get user name and profile pic
			"https://www.googleapis.com/auth/calendar",
			"https://www.googleapis.com/auth/calendar.events",
		},
		Endpoint: google.Endpoint,
	}

	return &Handler{
		oauthConfig: config,
		DB: db,
	}
}
