package utils

import (
	"errors"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// secretKey stores the JWT signing secret, retrieved from environment variables.
var secretKey = []byte(os.Getenv("SECRET_KEY"))

// GenerateToken creates a JWT containing the given OAuth access token.
// The token expires in 24 hours.
//
// Parameters:
//   - accessToken: The OAuth access token to embed in the JWT.
//
// Returns:
//   - A signed JWT as a string.
func GenerateToken(accessToken string) (string, error) {
	if len(secretKey) == 0 {
		return "", errors.New("SECRET_KEY is not set")
	}

	// Define JWT claims
	claims := jwt.MapClaims{
		"access_token": accessToken,
		"exp":          time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token using the secret key
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken verifies a JWT and extracts the access token.
//
// Parameters:
//   - tokenString: The JWT to validate.
//
// Returns:
//   - The extracted access token if valid.
//   - An error if the token is invalid or expired.
func ValidateToken(tokenString string) (string, error) {
	if len(secretKey) == 0 {
		return "", errors.New("SECRET_KEY is not set")
	}

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return "", err
	}

	// Extract and return the access token if valid
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		accessToken, ok := claims["access_token"].(string)
		if !ok {
			return "", errors.New("access_token not found in token claims")
		}
		return accessToken, nil
	}

	return "", errors.New("invalid token")
}
