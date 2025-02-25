package main

import (
	"google-calendar-api/cmd/server"
	"google-calendar-api/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file here: ", err)
	}

	// Connect to PostgreSQL database
	dsn := os.Getenv("DB_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("✅ Connected to database")

	// AutoMigrate the schema
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Meeting{})
	if err != nil {
        log.Fatal("❌ Migration failed:", err)
    }
	// Create handler with environment variables
	// h := handler.NewHandler()

	srv := server.NewServer(db)
	if err := srv.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
