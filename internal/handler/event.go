package handler

import (
	"encoding/json"
	"fmt"
	"google-calendar-api/models"
	"log"
	"net/http"
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

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	fmt.Println("📌 In Create Event")

	// Decode JSON request
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
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Println("❌ [ERROR] Failed to decode request body:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Retrieve OAuth token
	token, err := h.getUserTokenFromDB(r)
	if err != nil {
		log.Println("❌ [ERROR] Failed to retrieve user token:", err)
		http.Error(w, "Failed to retrieve token", http.StatusUnauthorized)
		return
	}

	// Create Google Calendar service
	client := h.oauthConfig.Client(oauth2.NoContext, token)
	service, err := calendar.New(client)
	if err != nil {
		http.Error(w, "Failed to create calendar service", http.StatusInternalServerError)
		return
	}

	// Parse Start and End DateTime to RFC3339
	parsedStartTime, err := time.Parse(time.RFC3339, request.Start.DateTime)
	if err != nil {
		log.Println("❌ [ERROR] Invalid Start DateTime:", err)
		http.Error(w, "Invalid Start DateTime format", http.StatusBadRequest)
		return
	}
	parsedEndTime, err := time.Parse(time.RFC3339, request.End.DateTime)
	if err != nil {
		log.Println("❌ [ERROR] Invalid End DateTime:", err)
		http.Error(w, "Invalid End DateTime format", http.StatusBadRequest)
		return
	}

	// Create event
	event := &calendar.Event{
		Summary:     request.Title,
		Description: request.Description,
		Start: &calendar.EventDateTime{
			DateTime: parsedStartTime.Format(time.RFC3339),
			TimeZone: request.Start.TimeZone,
		},
		End: &calendar.EventDateTime{
			DateTime: parsedEndTime.Format(time.RFC3339),
			TimeZone: request.End.TimeZone,
		},
	}

	// Insert event into Google Calendar
	createdEvent, err := service.Events.Insert("primary", event).Do()
	if err != nil {
		log.Println("❌ [ERROR] Failed to create event:", err)
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	// Store event in database
	meeting := models.Meeting{
		Title:       request.Title,
		Description: request.Description,
		StartTime:   parsedStartTime,
		EndTime:     parsedEndTime,
		EventID:     createdEvent.Id,
	}

	if err := h.DB.Create(&meeting).Error; err != nil {
		log.Println("❌ [ERROR] Failed to save event to database:", err)
		http.Error(w, "Failed to save event to database", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Event created successfully"})
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
