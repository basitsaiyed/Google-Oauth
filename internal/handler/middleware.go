package handler

import (
	"context"
	"fmt"
	"google-calendar-api/utils"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

// Define a type-safe context key
type contextKey string

const clientKey contextKey = "client"

// AuthMiddleware checks for the token in Authorization header or cookie
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var accessToken string

		// 1️⃣ Check "Authorization" header first
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				accessToken = tokenParts[1]
			}
		}

		// 2️⃣ If not found in header, check cookies
		if accessToken == "" {
			cookie, err := r.Cookie("token")
			fmt.Println("Cookie:", cookie.Value)
			if err != nil {
				http.Error(w, "No authorization header or token cookie", http.StatusUnauthorized)
				return
			}
			accessToken = cookie.Value
		}

		// 3️⃣ Validate token and set context
		ctx, err := validateAndSetContext(r, accessToken, h)
		if err != nil {
			http.Error(w, "Invalid token in middleware", http.StatusUnauthorized)
			return
		}

		// 4️⃣ Call the next handler with the new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateAndSetContext validates the token and returns a new context with the OAuth client
func validateAndSetContext(r *http.Request, accessToken string, h *Handler) (context.Context, error) {
	_, err := utils.ValidateToken(accessToken) // Ensure ValidateToken is correctly implemented
	if err != nil {
		return r.Context(), err // Return original request context instead of nil
	}

	client := h.oauthConfig.Client(r.Context(), &oauth2.Token{
		AccessToken: accessToken,
	})

	ctx := context.WithValue(r.Context(), clientKey, client) // Use type-safe key
	return ctx, nil
}
