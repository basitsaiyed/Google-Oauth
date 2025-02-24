package utils

import (
	"errors"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var secretKey = []byte(os.Getenv("SECRET_KEY"))

func GenerateToken(accessToken string) string {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "access_token": accessToken,
        "exp":         time.Now().Add(time.Hour * 24).Unix(),
    })
    
    tokenString, _ := token.SignedString(secretKey)
    return tokenString
}

func ValidateToken(tokenString string) (string, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return secretKey, nil
    })

    if err != nil {
        return "", err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return claims["access_token"].(string), nil
    }

    return "", errors.New("invalid token here in token.go")
}
