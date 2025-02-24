package handler

import (
	// "go/token"
	"context"
	"encoding/json"
	"google-calendar-api/utils"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"google-calendar-api/models"
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

// GoogleCallback: Verify state before proceeding
func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {

	state, err := r.Cookie("oauthstate")
	if err != nil || r.URL.Query().Get("state") != state.Value {
		http.Error(w, "Invalid OAuth state", http.StatusUnauthorized)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found in URL", http.StatusBadRequest)
		return
	}

	token, err := h.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Fetch user info from Google
	client := h.oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	// Save user details to DB
	user := models.User{
		GoogleID:     userInfo.ID,
		Email:        userInfo.Email,
		Name:         userInfo.Name,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(token.Expiry.Unix()) * time.Second),
	}

	// Check if user exists, update or insert
	var existingUser models.User
	result := h.DB.Where("google_id = ?", userInfo.ID).First(&existingUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			h.DB.Create(&user) // Insert new user
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	} else {
		existingUser.AccessToken = token.AccessToken
		existingUser.RefreshToken = token.RefreshToken
		existingUser.ExpiresAt = user.ExpiresAt
		h.DB.Save(&existingUser) // Update existing user
	}
	// Store token in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    utils.GenerateToken(token.AccessToken),
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Change to true in production
		SameSite: http.SameSiteLaxMode,
	})

	log.Println("Setting cookie with token:", token.AccessToken)
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
