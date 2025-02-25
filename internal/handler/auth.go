package handler

import (
	// "go/token"

	"errors"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"google-calendar-api/models"

	"github.com/coreos/go-oidc"
	"gorm.io/gorm"
)

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join("templates", "login.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Message string
	}{
		Message: "Please log in to access your dashboard.",
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	_, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	tmplPath := filepath.Join("templates", "dashboard.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// GoogleLogin: Generate and store state
func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := "random_string" // Generate a proper random state
	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		HttpOnly: true,
		Secure:   false, // Use true in production with HTTPS
	})
	url := h.oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles the OAuth2 callback from Google
func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Get "code" from URL parameters
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Println("❌ No code found in request")
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	// Exchange auth code for tokens (access token + ID token)
	token, err := h.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Println("❌ Failed to exchange token:", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}
	log.Println("🔹 OAuth2 Token Response:", token)

	// Extract ID Token from the token response
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Println("❌ No ID Token found in OAuth response")
		http.Error(w, "No ID Token received", http.StatusInternalServerError)
		return
	}

	// Create OIDC provider using Google
	provider, err := oidc.NewProvider(r.Context(), "https://accounts.google.com")
	if err != nil {
		log.Println("❌ Failed to create OIDC provider:", err)
		http.Error(w, "Failed to verify ID token", http.StatusInternalServerError)
		return
	}

	// Verify and decode the ID Token
	verifier := provider.Verifier(&oidc.Config{ClientID: h.oauthConfig.ClientID})
	idTokenObj, err := verifier.Verify(r.Context(), idToken)
	if err != nil {
		log.Println("❌ Invalid ID Token:", err)
		http.Error(w, "Invalid ID Token", http.StatusUnauthorized)
		return
	}

	// Decode token claims to get user details
	var userInfo struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := idTokenObj.Claims(&userInfo); err != nil {
		log.Println("❌ Failed to parse ID Token claims:", err)
		http.Error(w, "Failed to parse ID Token", http.StatusInternalServerError)
		return
	}

	// Ensure required fields are present
	if userInfo.Sub == "" || userInfo.Email == "" {
		log.Println("❌ UserInfo missing required fields:", userInfo)
		http.Error(w, "Invalid user info received", http.StatusInternalServerError)
		return
	}

	// Check if the user exists in the database
	var existingUser models.User
	result := h.DB.Where("google_id = ?", userInfo.Sub).First(&existingUser)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// New user, create record
			newUser := models.User{
				GoogleID:     userInfo.Sub,
				Email:        userInfo.Email,
				Name:         userInfo.Name,
				Picture:      userInfo.Picture,
				AccessToken:  token.AccessToken,
				RefreshToken: token.RefreshToken,
				ExpiresAt:    token.Expiry,
			}

			log.Println("🔹 New user, inserting into DB...")
			if err := h.DB.Create(&newUser).Error; err != nil {
				log.Println("❌ Error inserting user:", err)
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
		} else {
			log.Println("❌ Error fetching user:", result.Error)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	} else {
		// Existing user, update access token
		log.Println("🔹 Existing user found, updating tokens...")
		existingUser.AccessToken = token.AccessToken
		existingUser.ExpiresAt = token.Expiry

		if token.RefreshToken != "" {
			existingUser.RefreshToken = token.RefreshToken
		}

		if err := h.DB.Save(&existingUser).Error; err != nil {
			log.Println("❌ Error updating user:", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	// Store token in a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    idToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	log.Println("✅ User authenticated successfully:", userInfo.Email)

	// Redirect to dashboard or another relevant page
	http.Redirect(w, r, "/api/dashboard", http.StatusTemporaryRedirect)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}
