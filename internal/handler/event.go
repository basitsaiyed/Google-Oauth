package handler

import (
	"encoding/json"
	"fmt"
	"google-calendar-api/models"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
)

/*
Step-by-Step Process:
1. Decode the incoming JSON request payload to extract event details.
2. Retrieve the OAuth token from the database (user authentication).
3. Create a Google Calendar service using the OAuth token.
4. Convert the received date-time format into RFC3339 format.
5. Create a new event structure and set necessary details.
6. Insert the event into the Google Calendar using the API.
7. Store the event details in the PostgreSQL database.
8. Return a success or failure response.
*/

// CreateEvent handles the creation of a new Google Calendar event
func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	fmt.Println("📌 In CreateEvent handler")

	// Step 1: Decode JSON request body
	var request struct {
		Title       string `json:"summary"`
		Description string `json:"description"`
		Start       struct {
			DateTime string `json:"dateTime"`
			TimeZone string `json:"timeZone"`
		} `json:"start"`
		End struct {
			DateTime string `json:"dateTime"`
			TimeZone string `json:"timeZone"`
		} `json:"end"`
		Attendees []string `json:"attendees"` // List of attendee emails
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Println("[ERROR] Failed to decode request body:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Step 2: Retrieve OAuth token from session or database
	token, userEmail, err := h.getUserTokenFromDB(r) // Implement this function to get the token
	if err != nil {
		log.Println("[ERROR] Failed to retrieve user token:", err)
		http.Error(w, "Failed to retrieve token", http.StatusUnauthorized)
		return
	}
	fmt.Println("📌 Token Retrieved Successfully")

	// Step 3: Create Google Calendar service client
	client := h.oauthConfig.Client(oauth2.NoContext, token)
	service, err := calendar.New(client)
	if err != nil {
		log.Println("[ERROR] Failed to create calendar service:", err)
		http.Error(w, "Failed to create calendar service", http.StatusInternalServerError)
		return
	}
	fmt.Println("📌 Google Calendar Service Created")

	// Step 4: Convert attendee emails into Google Calendar Attendee objects
	var eventAttendees []*calendar.EventAttendee
	for _, email := range request.Attendees {
		eventAttendees = append(eventAttendees, &calendar.EventAttendee{Email: email})
	}

	// Step 5: Create the event object
	event := &calendar.Event{
		Summary:     request.Title,
		Description: request.Description,
		Start: &calendar.EventDateTime{
			DateTime: request.Start.DateTime,
			TimeZone: request.Start.TimeZone,
		},
		End: &calendar.EventDateTime{
			DateTime: request.End.DateTime,
			TimeZone: request.End.TimeZone,
		},
		Attendees: eventAttendees, // Add attendees
	}

	// Step 6: Log the event details
	log.Println("📌 Creating Event:")
	log.Printf("    - Title: %s", event.Summary)
	log.Printf("    - Start: %s (%s)", event.Start.DateTime, event.Start.TimeZone)
	log.Printf("    - End: %s (%s)", event.End.DateTime, event.End.TimeZone)
	log.Printf("    - Attendees: %v", request.Attendees)

	// Step 7: Insert event into Google Calendar
	createdEvent, err := service.Events.Insert("primary", event).Do()
	if err != nil {
		log.Println("[ERROR] Failed to create event in Google Calendar:", err)
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	// Step 8: Store event in the database
	startTime, _ := time.Parse(time.RFC3339, request.Start.DateTime)
	endTime, _ := time.Parse(time.RFC3339, request.End.DateTime)

	meeting := models.Meeting{
		Title:       request.Title,
		Description: request.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		EventID:     createdEvent.Id,
		Attendees:   strings.Join(request.Attendees, ","), // Store emails as comma-separated values
        CreatedBy:   userEmail,
	}

	if err := h.DB.Create(&meeting).Error; err != nil {
		log.Println("[ERROR] Failed to save event to database:", err)
		http.Error(w, "Failed to save event to database", http.StatusInternalServerError)
		return
	}

	// Step 9: Respond with success message
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Event created successfully", "event_id": createdEvent.Id})

	fmt.Println("✅ Event Created Successfully!")
}

func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	// Fetch all events from database
	var meetings []models.Meeting
	if err := h.DB.Find(&meetings).Error; err != nil {
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}

	// Return events as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"events": meetings})
}
