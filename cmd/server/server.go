package server

import (
	"google-calendar-api/internal/handler"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type Server struct {
	router *mux.Router
}

func NewServer(db *gorm.DB) *Server {
	s := &Server{
		router: mux.NewRouter(),
	}
	s.setupRoutes(db)
	return s
}

func (s *Server) setupRoutes(db *gorm.DB) {
	h := handler.NewHandler(db)

	// Auth routes
	s.router.HandleFunc("/login", h.LoginPage).Methods("GET")
	s.router.HandleFunc("/auth/google/login", h.GoogleLogin).Methods("GET")
	s.router.HandleFunc("/auth/google/callback", h.GoogleCallback).Methods("GET")

	// Event routes (protected by middleware)
	api := s.router.PathPrefix("/api").Subrouter()
	// api.Handle("/dashboard", h.AuthMiddleware(http.HandlerFunc(h.Dashboard))).Methods("GET")
	api.Use(h.AuthMiddleware)
	api.HandleFunc("/dashboard", h.Dashboard).Methods("GET")
    api.HandleFunc("/events/create", h.CreateEvent).Methods("POST") // Create event
    api.HandleFunc("/events/list", h.ListEvents).Methods("GET")     // List events    

    // api.HandleFunc("/event/create-meeting", h.CreateGoogleCalendarEvent).Methods("POST") // ✅ New route added
	api.HandleFunc("/logout", h.Logout)
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
