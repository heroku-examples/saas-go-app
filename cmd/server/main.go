package main

import (
	"log"
	"os"

	"saas-go-app/internal/api"
	"saas-go-app/internal/auth"
	"saas-go-app/internal/db"
	"saas-go-app/internal/jobs"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load environment variables from .env file (if it exists)
	_ = godotenv.Load()

	// Initialize JWT
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

	// Create database tables
	if err := db.CreateTables(); err != nil {
		log.Fatal("Failed to create database tables:", err)
	}

	// Seed database with sample data if SEED_DATA is set
	if os.Getenv("SEED_DATA") == "true" {
		if err := db.SeedDataIfEmpty(); err != nil {
			log.Printf("Warning: Failed to seed database: %v", err)
		}
	}

	// Initialize background job processor
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		srv := asynq.NewServer(
			asynq.RedisClientOpt{Addr: redisURL},
			asynq.Config{
				Concurrency: 10,
				Queues: map[string]int{
					"critical": 6,
					"default":  3,
					"low":      1,
				},
			},
		)

		mux := asynq.NewServeMux()
		mux.HandleFunc(jobs.TypeAggregateData, jobs.HandleAggregationTask)

		go func() {
			log.Println("Starting background job processor...")
			if err := srv.Run(mux); err != nil {
				log.Fatalf("Failed to start background job processor: %v", err)
			}
		}()
	} else {
		log.Println("REDIS_URL not set, background jobs will not be processed")
	}

	// Set up Gin router
	router := gin.Default()

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check endpoint
	router.GET("/health", api.HealthCheck)

	// Public routes
	apiRoutes := router.Group("/api")
	{
		apiRoutes.POST("/auth/login", api.Login)
		apiRoutes.POST("/auth/register", api.Register)
	}

	// Protected routes
	protectedRoutes := apiRoutes.Group("")
	protectedRoutes.Use(auth.AuthMiddleware())
	{
		// Customer routes
		customers := protectedRoutes.Group("/customers")
		{
			customers.GET("", api.GetCustomers)
			customers.GET("/:id", api.GetCustomer)
			customers.POST("", api.CreateCustomer)
			customers.PUT("/:id", api.UpdateCustomer)
			customers.DELETE("/:id", api.DeleteCustomer)
		}

		// Account routes
		accounts := protectedRoutes.Group("/accounts")
		{
			accounts.GET("", api.GetAccounts)
			accounts.GET("/:id", api.GetAccount)
			accounts.POST("", api.CreateAccount)
			accounts.PUT("/:id", api.UpdateAccount)
			accounts.DELETE("/:id", api.DeleteAccount)
		}

		// Analytics routes
		analytics := protectedRoutes.Group("/analytics")
		{
			analytics.GET("", api.GetAnalytics)
			analytics.GET("/customers/:customer_id", api.GetCustomerAnalytics)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

