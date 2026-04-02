package main

import (
	"log"
	"os"

	"github.com/vikhyat-sharma/astrology-ai/internal/config"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"github.com/vikhyat-sharma/astrology-ai/internal/handlers"
	"github.com/vikhyat-sharma/astrology-ai/internal/middleware"
	"github.com/vikhyat-sharma/astrology-ai/internal/repositories"
	"github.com/vikhyat-sharma/astrology-ai/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db := database.InitDB(cfg.DatabaseURL)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	astrologyRepo := repositories.NewAstrologyRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	astrologyService := services.NewAstrologyService(astrologyRepo, cfg.OllamaURL, cfg.OllamaModel)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	astrologyHandler := handlers.NewAstrologyHandler(astrologyService)

	// Setup Gin router
	router := gin.Default()

	// Global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.ErrorHandler())

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthRequired())
		{
			// User routes
			user := protected.Group("/user")
			{
				user.GET("/profile", authHandler.GetProfile)
				user.PUT("/profile", authHandler.UpdateProfile)
				user.POST("/birth-info", authHandler.UpdateBirthInfo)
			}

			// Astrology routes
			astro := protected.Group("/astrology")
			{
				astro.POST("/birth-chart", astrologyHandler.CreateBirthChart)
				astro.GET("/birth-chart/:id", astrologyHandler.GetBirthChart)
				astro.GET("/horoscope/daily", astrologyHandler.GetDailyHoroscope)
				astro.POST("/compatibility", astrologyHandler.CheckCompatibility)
			}
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

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
