package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vikhyat-sharma/astrology-ai/internal/constants"
	"github.com/vikhyat-sharma/astrology-ai/internal/handlers"
	"github.com/vikhyat-sharma/astrology-ai/internal/middleware"
	"github.com/vikhyat-sharma/astrology-ai/internal/ports"
	"gorm.io/gorm"
)

// SetupRoutes configures all API routes.
// authService is injected so middleware never retrieves it from the Gin context.
func SetupRoutes(
	router *gin.Engine,
	db *gorm.DB,
	authHandler *handlers.AuthHandler,
	astrologyHandler *handlers.AstrologyHandler,
	authService ports.AuthService,
	allowedOrigins []string,
) {
	// Security: only trust the loopback proxy. Must be set before routes.
	router.SetTrustedProxies([]string{"127.0.0.1"}) //nolint:errcheck

	// Global middleware — order matters.
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.CORS(allowedOrigins))
	router.Use(middleware.Logger())
	router.Use(middleware.ErrorHandler())

	// Health check — no auth, no rate limit.
	router.GET(constants.HealthEndpoint, func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group(constants.APIV1Prefix)

	// Auth routes — strict rate limit (5 req/min per IP) to prevent brute force.
	auth := api.Group(constants.AuthPrefix)
	auth.Use(middleware.RateLimit(5.0/60.0, 5))
	{
		auth.POST(constants.RegisterEndpoint, authHandler.Register)
		auth.POST(constants.LoginEndpoint, authHandler.Login)
	}

	// Protected routes — JWT required.
	protected := api.Group("")
	protected.Use(middleware.AuthRequired(authService))
	{
		// User routes
		user := protected.Group(constants.UserPrefix)
		{
			user.GET(constants.ProfileEndpoint, authHandler.GetProfile)
			user.PUT(constants.ProfileEndpoint, authHandler.UpdateProfile)
			user.POST(constants.BirthInfoEndpoint, authHandler.UpdateBirthInfo)
		}

		// Astrology routes — moderate rate limit (30 req/min per IP).
		astro := protected.Group(constants.AstrologyPrefix)
		astro.Use(middleware.RateLimit(30.0/60.0, 10))
		{
			astro.POST(constants.BirthChartEndpoint, astrologyHandler.CreateBirthChart)
			astro.GET(constants.BirthChartEndpoint+"/:id", astrologyHandler.GetBirthChart)
			astro.GET(constants.DailyHoroscopeEndpoint+"/:sign", astrologyHandler.GetHoroscope)
			astro.POST(constants.CompatibilityEndpoint, astrologyHandler.CheckCompatibility)
			astro.GET(constants.RemediesEndpoint+"/:id", astrologyHandler.GetRemedies)
			astro.POST(constants.PersonalizedHoroscopeEndpoint, astrologyHandler.GeneratePersonalizedHoroscope)
			astro.POST("/graphql", astrologyHandler.GraphQLHandler())
			astro.GET("/transits/now", astrologyHandler.GetCurrentTransits)
		}
	}
}
