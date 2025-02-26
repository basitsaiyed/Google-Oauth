package handler

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc"
)

// contextKey is a type-safe key for storing user information in the request context.
type contextKey string

const userKey contextKey = "user"

// AuthMiddleware validates authentication tokens from request headers or cookies.
// It verifies the token against Google's OIDC provider and injects user details into the request context.
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var accessToken string

		// Check "Authorization" header for a Bearer token
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				accessToken = tokenParts[1]
			}
		}

		// If no token found in header, check cookies
		if accessToken == "" {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Unauthorized: No valid authentication token", http.StatusUnauthorized)
				return
			}
			accessToken = cookie.Value
		}

		// Validate token and set user info in request context
		ctx, err := h.validateAndSetContext(r, accessToken)
		if err != nil {
			http.Error(w, "Unauthorized: Invalid authentication token", http.StatusUnauthorized)
			return
		}

		// Proceed to the next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateAndSetContext verifies the given token using Google's OIDC provider
// and returns a new request context containing user details.
func (h *Handler) validateAndSetContext(r *http.Request, idToken string) (context.Context, error) {
	provider, err := oidc.NewProvider(r.Context(), "https://accounts.google.com")
	if err != nil {
		log.Println("❌ Failed to create OIDC provider:", err)
		return r.Context(), err
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: h.oauthConfig.ClientID})
	idTokenObj, err := verifier.Verify(r.Context(), idToken)
	if err != nil {
		log.Println("❌ Invalid ID Token:", err)
		return r.Context(), err
	}

	// Extract user claims from ID token
	var userInfo struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idTokenObj.Claims(&userInfo); err != nil {
		log.Println("❌ Failed to parse ID Token claims:", err)
		return r.Context(), err
	}

	// Store user info in request context
	ctx := context.WithValue(r.Context(), userKey, userInfo)
	return ctx, nil
}
