package server

import (
    "net/http"
    "github.com/gorilla/mux"
    "google-calendar-api/internal/handler"
)

type Server struct {
    router *mux.Router
}

func NewServer() *Server {
    s := &Server{
        router: mux.NewRouter(),
    }
    s.setupRoutes()
    return s
}

func (s *Server) setupRoutes() {
    h := handler.NewHandler()
    
    // Auth routes
    s.router.HandleFunc("/login", h.LoginPage).Methods("GET")
    s.router.HandleFunc("/auth/google/login", h.GoogleLogin).Methods("GET")
    s.router.HandleFunc("/auth/google/callback", h.GoogleCallback).Methods("GET")
    
    // Event routes (protected by middleware)
    api := s.router.PathPrefix("/api").Subrouter()
    // api.Handle("/dashboard", h.AuthMiddleware(http.HandlerFunc(h.Dashboard))).Methods("GET")
    api.Use(h.AuthMiddleware)
    api.HandleFunc("/dashboard", h.Dashboard).Methods("GET")
    api.HandleFunc("/events", h.CreateEvent).Methods("POST")
    api.HandleFunc("/events", h.ListEvents).Methods("GET")
    api.HandleFunc("/event/create", h.CreateEvent)
    api.HandleFunc("/event/all", h.ListEvents).Methods("GET")
    api.HandleFunc("/logout", h.Logout)
}

func (s *Server) Run(addr string) error {
    return http.ListenAndServe(addr, s.router)
}

// internal/handler/handler.go
