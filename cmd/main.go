package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vikhyat-sharma/astrology-ai/internal/config"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"github.com/vikhyat-sharma/astrology-ai/internal/handlers"
	"github.com/vikhyat-sharma/astrology-ai/internal/repositories"
	"github.com/vikhyat-sharma/astrology-ai/internal/services"
)

func main() {
	// Structured JSON logging for production log aggregation.
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Fail fast on missing or invalid configuration — never start with defaults.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	db := database.InitDB(cfg.DatabaseURL)

	// Wire dependencies.
	userRepo := repositories.NewUserRepository(db)
	astrologyRepo := repositories.NewAstrologyRepository(db)

	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	astrologyService := services.NewAstrologyService(astrologyRepo, cfg.OllamaURL, cfg.OllamaModel)

	authHandler := handlers.NewAuthHandler(authService)
	astrologyHandler := handlers.NewAstrologyHandler(astrologyService)

	// Use gin.New() — not gin.Default() — so we control every middleware explicitly.
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	SetupRoutes(router, db, authHandler, astrologyHandler, authService, cfg.AllowedOrigins)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.Port, "environment", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	slog.Info("server exited")
}
