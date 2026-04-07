package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/services"
)

// AstrologyHandler handles astrology HTTP requests
type AstrologyHandler struct {
	astrologyService *services.AstrologyService
}

// NewAstrologyHandler creates a new astrology handler
func NewAstrologyHandler(astrologyService *services.AstrologyService) *AstrologyHandler {
	return &AstrologyHandler{astrologyService: astrologyService}
}

// CreateBirthChartRequest represents the create birth chart request payload
type CreateBirthChartRequest struct {
	BirthDate  string  `json:"birth_date" binding:"required"` // ISO 8601 date string
	BirthTime  string  `json:"birth_time" binding:"required"`
	BirthPlace string  `json:"birth_place" binding:"required"`
	Latitude   float64 `json:"latitude" binding:"required"`
	Longitude  float64 `json:"longitude" binding:"required"`
	Timezone   string  `json:"timezone" binding:"required"`
}

// CreateBirthChart handles creating a birth chart
func (h *AstrologyHandler) CreateBirthChart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req CreateBirthChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse birth date
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid birth date format. Use YYYY-MM-DD"})
		return
	}

	data := services.BirthChartData{
		UserID:     userID.(uuid.UUID),
		BirthDate:  birthDate,
		BirthTime:  req.BirthTime,
		BirthPlace: req.BirthPlace,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		Timezone:   req.Timezone,
	}

	chart, err := h.astrologyService.CreateBirthChart(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Birth chart created successfully",
		"chart":   chart,
	})
}

// GetBirthChart handles getting a birth chart by ID
func (h *AstrologyHandler) GetBirthChart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	chartIDStr := c.Param("id")
	chartID, err := uuid.Parse(chartIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chart ID"})
		return
	}

	chart, err := h.astrologyService.GetBirthChart(chartID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Check if the chart belongs to the authenticated user
	if chart.UserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"chart": chart})
}

// GetDailyHoroscope handles getting daily horoscope
func (h *AstrologyHandler) GetDailyHoroscope(c *gin.Context) {
	sign := c.Query("sign")
	if sign == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sign parameter is required"})
		return
	}

	horoscope, err := h.astrologyService.GetDailyHoroscope(sign)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"horoscope": horoscope})
}

// CheckCompatibilityRequest represents the compatibility check request payload
type CheckCompatibilityRequest struct {
	ChartID1 string `json:"chart_id_1" binding:"required"`
	ChartID2 string `json:"chart_id_2" binding:"required"`
}

// CheckCompatibility handles checking compatibility between two birth charts
func (h *AstrologyHandler) CheckCompatibility(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req CheckCompatibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chartID1, err := uuid.Parse(req.ChartID1)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chart ID 1"})
		return
	}

	chartID2, err := uuid.Parse(req.ChartID2)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chart ID 2"})
		return
	}

	// Verify that both charts belong to the authenticated user
	chart1, err := h.astrologyService.GetBirthChart(chartID1)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chart 1 not found"})
		return
	}

	chart2, err := h.astrologyService.GetBirthChart(chartID2)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chart 2 not found"})
		return
	}

	if chart1.UserID != userID.(uuid.UUID) || chart2.UserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	compatibility, err := h.astrologyService.CheckCompatibility(chartID1, chartID2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"compatibility": compatibility})
}

// GetRemedies handles getting remedies based on a birth chart
func (h *AstrologyHandler) GetRemedies(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	chartIDStr := c.Param("id")
	chartID, err := uuid.Parse(chartIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chart ID"})
		return
	}

	chart, err := h.astrologyService.GetBirthChart(chartID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Check if the chart belongs to the authenticated user
	if chart.UserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	remedies, err := h.astrologyService.GetRemedies(chart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"remedies": remedies})
}
