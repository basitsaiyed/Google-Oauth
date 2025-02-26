package models

import "time"

// Meeting represents a scheduled meeting with details like title, description, time, and attendees.
// It includes an associated Google Calendar event ID for synchronization.
type Meeting struct {
	ID          uint      `gorm:"primaryKey" json:"id"`    // Unique meeting ID (Primary Key)
	Title       string    `json:"title"`                   // Meeting title
	Description string    `json:"description"`             // Meeting description or agenda
	StartTime   time.Time `json:"start_time"`              // Meeting start time
	EndTime     time.Time `json:"end_time"`                // Meeting end time
	EventID     string    `json:"event_id"`                // Google Calendar Event ID
	Attendees   string    `json:"attendees"`               // Comma-separated list of attendee emails
	CreatedBy   string    `json:"created_by"`              // Email of the user who created the meeting
	CreatedAt   time.Time `json:"created_at"`              // Timestamp of when the meeting was created
	UpdatedAt   time.Time `json:"updated_at"`              // Timestamp of the last update
}
