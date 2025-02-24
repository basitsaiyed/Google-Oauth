package handler

import (
    "encoding/json"
    "net/http"
    "google.golang.org/api/calendar/v3"
)

type EventRequest struct {
    Summary     string `json:"summary"`
    Description string `json:"description"`
    StartTime   string `json:"start_time"`
    EndTime     string `json:"end_time"`
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
    var req EventRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    srv, err := calendar.New(r.Context().Value("client").(*http.Client))
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    event := &calendar.Event{
        Summary:     req.Summary,
        Description: req.Description,
        Start: &calendar.EventDateTime{
            DateTime: req.StartTime,
            TimeZone: "UTC",
        },
        End: &calendar.EventDateTime{
            DateTime: req.EndTime,
            TimeZone: "UTC",
        },
    }

    event, err = srv.Events.Insert("primary", event).Do()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(event)
}

func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
    srv, err := calendar.New(r.Context().Value("client").(*http.Client))
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    events, err := srv.Events.List("primary").MaxResults(10).
        OrderBy("startTime").SingleEvents(true).Do()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(events.Items)
}

