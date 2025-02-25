package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	GoogleID     string    `gorm:"unique"`
	Email        string    `gorm:"unique"`
	Name         string
	Picture      string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}
