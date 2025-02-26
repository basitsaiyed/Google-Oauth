package server

import (
	"net/http"

	"google-calendar-api/internal/handler"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// Server struct represents the application server with a router.
type Server struct {
	router *mux.Router
}

// NewServer initializes a new Server instance and sets up the routes.
func NewServer(db *gorm.DB) *Server {
	s := &Server{
		router: mux.NewRouter(),
	}
	s.setupRoutes(db)
	return s
}

// setupRoutes configures all the API routes and assigns them to the router.
func (s *Server) setupRoutes(db *gorm.DB) {
	h := handler.NewHandler(db)

	// Authentication routes
	s.router.HandleFunc("/login", h.LoginPage).Methods("GET")
	s.router.HandleFunc("/auth/google/login", h.GoogleLogin).Methods("GET")
	s.router.HandleFunc("/auth/google/callback", h.GoogleCallback).Methods("GET")

	// Protected API routes (require authentication)
	api := s.router.PathPrefix("/api").Subrouter()
	api.Use(h.AuthMiddleware) // Apply authentication middleware

	api.HandleFunc("/dashboard", h.Dashboard).Methods("GET") // Dashboard route
	api.HandleFunc("/events/create", h.CreateEvent).Methods("POST") // Create event
	api.HandleFunc("/events/list", h.ListEvents).Methods("GET") // List events

	// Logout route
	s.router.HandleFunc("/logout", h.Logout)
}

// Run starts the HTTP server on the specified address.
func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
