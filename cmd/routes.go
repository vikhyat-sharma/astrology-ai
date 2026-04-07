package main

import (
	"github.com/vikhyat-sharma/astrology-ai/internal/constants"
	"github.com/vikhyat-sharma/astrology-ai/internal/handlers"
	"github.com/vikhyat-sharma/astrology-ai/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all the API routes for the application
func SetupRoutes(router *gin.Engine, authHandler *handlers.AuthHandler, astrologyHandler *handlers.AstrologyHandler) {
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
				astro.GET(constants.RemediesEndpoint+"/:id", astrologyHandler.GetRemedies)
			}
		}
	}

	// Health check
	router.GET(constants.HealthEndpoint, func(c *gin.Context) {
		c.JSON(constants.StatusOK, gin.H{"status": "ok"})
	})

	// Set trusted proxies for security
	router.SetTrustedProxies([]string{"127.0.0.1"})
}
