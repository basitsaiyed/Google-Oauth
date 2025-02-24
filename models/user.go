package models

import "time"

type User struct {
    ID           int       `gorm:"primaryKey"`
    GoogleID     string    `gorm:"unique"`
    Email        string    `gorm:"unique"`
    Name         string
    AccessToken  string
    RefreshToken string
    ExpiresAt    time.Time
}
