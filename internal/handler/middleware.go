package handler

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc"
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
			// fmt.Println("Cookie:", cookie.Value)
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
func validateAndSetContext(r *http.Request, idToken string, h *Handler) (context.Context, error) {
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

	var userInfo struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := idTokenObj.Claims(&userInfo); err != nil {
		log.Println("❌ Failed to parse ID Token claims:", err)
		return r.Context(), err
	}

	ctx := context.WithValue(r.Context(), "user", userInfo)
	return ctx, nil
}
