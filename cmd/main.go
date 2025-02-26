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
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("❌ Error loading .env file:", err)
	}

	// Retrieve database connection string from environment variables
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		log.Fatal("❌ Database URL not provided in environment variables")
	}

	// Initialize PostgreSQL database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	log.Println("✅ Connected to database")

	// Run database migrations for required models
	if err := db.AutoMigrate(&models.User{}, &models.Meeting{}); err != nil {
		log.Fatal("❌ Migration failed:", err)
	}
	log.Println("✅ Database migration completed")

	// Initialize and start the HTTP server
	srv := server.NewServer(db)
	log.Println("🚀 Server is running on port 8080")
	if err := srv.Run(":8080"); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
