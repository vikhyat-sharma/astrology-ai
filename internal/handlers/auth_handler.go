package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/services"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.RegisterUser(req.Email, req.Password, req.Name)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.authService.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

// GetProfile handles getting user profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := h.authService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":          user.ID,
			"email":       user.Email,
			"name":        user.Name,
			"birth_date":  user.BirthDate,
			"birth_time":  user.BirthTime,
			"birth_place": user.BirthPlace,
			"latitude":    user.Latitude,
			"longitude":   user.Longitude,
			"timezone":    user.Timezone,
		},
	})
}

// UpdateProfileRequest represents the update profile request payload
type UpdateProfileRequest struct {
	Name       string  `json:"name"`
	BirthDate  string  `json:"birth_date"` // ISO 8601 date string
	BirthTime  string  `json:"birth_time"`
	BirthPlace string  `json:"birth_place"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Timezone   string  `json:"timezone"`
}

// UpdateProfile handles updating user profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Update user fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.BirthDate != "" {
		// Parse birth date (simplified - in production, use proper date parsing)
		user.BirthDate, _ = time.Parse("2006-01-02", req.BirthDate)
	}
	if req.BirthTime != "" {
		user.BirthTime = req.BirthTime
	}
	if req.BirthPlace != "" {
		user.BirthPlace = req.BirthPlace
	}
	user.Latitude = req.Latitude
	user.Longitude = req.Longitude
	if req.Timezone != "" {
		user.Timezone = req.Timezone
	}

	if err := h.authService.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":          user.ID,
			"email":       user.Email,
			"name":        user.Name,
			"birth_date":  user.BirthDate,
			"birth_time":  user.BirthTime,
			"birth_place": user.BirthPlace,
			"latitude":    user.Latitude,
			"longitude":   user.Longitude,
			"timezone":    user.Timezone,
		},
	})
}

// BirthInfoRequest represents the birth info request payload
type BirthInfoRequest struct {
	BirthDate  string  `json:"birth_date" binding:"required"`
	BirthTime  string  `json:"birth_time" binding:"required"`
	BirthPlace string  `json:"birth_place" binding:"required"`
	Latitude   float64 `json:"latitude" binding:"required"`
	Longitude  float64 `json:"longitude" binding:"required"`
	Timezone   string  `json:"timezone" binding:"required"`
}

// UpdateBirthInfo handles updating date/time/location details
func (h *AuthHandler) UpdateBirthInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req BirthInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid birth date format. Use YYYY-MM-DD"})
		return
	}

	user.BirthDate = birthDate
	user.BirthTime = req.BirthTime
	user.BirthPlace = req.BirthPlace
	user.Latitude = req.Latitude
	user.Longitude = req.Longitude
	user.Timezone = req.Timezone

	if err := h.authService.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Birth info updated successfully", "user": user})
}
