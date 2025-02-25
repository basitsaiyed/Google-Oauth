package models

import (
	"time"
)

type Meeting struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	EventID     string    `json:"event_id"` // Google Calendar Event ID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
