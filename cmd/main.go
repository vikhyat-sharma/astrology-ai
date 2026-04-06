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
	"github.com/vikhyat-sharma/astrology-ai/internal/constants"
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
	api := router.Group(constants.APIV1Prefix)
	{
		// Public routes
		auth := api.Group(constants.AuthPrefix)
		{
			auth.POST(constants.RegisterEndpoint, authHandler.Register)
			auth.POST(constants.LoginEndpoint, authHandler.Login)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthRequired())
		{
			// User routes
			user := protected.Group(constants.UserPrefix)
			{
				user.GET(constants.ProfileEndpoint, authHandler.GetProfile)
				user.PUT(constants.ProfileEndpoint, authHandler.UpdateProfile)
				user.POST(constants.BirthInfoEndpoint, authHandler.UpdateBirthInfo)
			}

			// Astrology routes
			astro := protected.Group(constants.AstrologyPrefix)
			{
				astro.POST(constants.BirthChartEndpoint, astrologyHandler.CreateBirthChart)
				astro.GET(constants.BirthChartEndpoint+"/:id", astrologyHandler.GetBirthChart)
				astro.GET(constants.DailyHoroscopeEndpoint, astrologyHandler.GetDailyHoroscope)
				astro.POST(constants.CompatibilityEndpoint, astrologyHandler.CheckCompatibility)
			}
		}
	}

	// Health check
	router.GET(constants.HealthEndpoint, func(c *gin.Context) {
		c.JSON(constants.StatusOK, gin.H{"status": "ok"})
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
