package main

import (
	"google-calendar-api/cmd/server"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file here: ", err)
	}

	// Create handler with environment variables
	// h := handler.NewHandler()

	srv := server.NewServer()
	if err := srv.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
