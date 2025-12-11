package main

import (
	"log"
	"os"

	"saas-go-app/internal/auth"
	"saas-go-app/internal/db"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file (if it exists)
	_ = godotenv.Load()

	// Initialize JWT (needed for password hashing)
	if err := auth.InitJWT(); err != nil {
		log.Fatal("Failed to initialize JWT:", err)
	}

	// Initialize database connections
	if err := db.InitPrimaryDB(); err != nil {
		log.Fatal("Failed to initialize primary database:", err)
	}
	defer db.CloseDB()

	if err := db.InitAnalyticsDB(); err != nil {
		log.Printf("Warning: Failed to initialize analytics database: %v", err)
	}

	// Create database tables (in case they don't exist)
	if err := db.CreateTables(); err != nil {
		log.Fatal("Failed to create database tables:", err)
	}

	// Clear and reseed
	if err := db.ClearAndReseed(); err != nil {
		log.Fatal("Failed to clear and reseed database:", err)
	}

	log.Println("Database cleared and reseeded successfully!")
}

