package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Set trusted proxies for security
	router.SetTrustedProxies([]string{"127.0.0.1"})

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
