package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"github.com/vikhyat-sharma/astrology-ai/internal/ports"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authService ports.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService ports.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRequest is the registration request payload.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

// LoginRequest is the login request payload.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register handles user registration.
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.RegisterUser(c.Request.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"user":    userView(user),
	})
}

// Login handles user login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.authService.AuthenticateUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"token":   token,
		"user":    userView(user),
	})
}

// GetProfile handles fetching the authenticated user's profile.
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := mustUserID(c)
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": profileView(user)})
}

// UpdateProfileRequest uses pointer fields so zero values are distinguishable from absent fields.
type UpdateProfileRequest struct {
	Name       *string  `json:"name"`
	BirthDate  *string  `json:"birth_date"`
	BirthTime  *string  `json:"birth_time"`
	BirthPlace *string  `json:"birth_place"`
	Latitude   *float64 `json:"latitude"`
	Longitude  *float64 `json:"longitude"`
	Timezone   *string  `json:"timezone"`
}

// UpdateProfile handles partial profile updates.
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := mustUserID(c)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.BirthDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid birth_date format, expected YYYY-MM-DD"})
			return
		}
		user.BirthDate = parsed
	}
	if req.BirthTime != nil {
		user.BirthTime = *req.BirthTime
	}
	if req.BirthPlace != nil {
		user.BirthPlace = *req.BirthPlace
	}
	// Only update coordinates when explicitly provided — prevents silent zeroing.
	if req.Latitude != nil {
		user.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		user.Longitude = *req.Longitude
	}
	if req.Timezone != nil {
		user.Timezone = *req.Timezone
	}

	if err := h.authService.UpdateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "profile updated successfully",
		"user":    profileView(user),
	})
}

// BirthInfoRequest is the birth info update payload.
type BirthInfoRequest struct {
	BirthDate  string  `json:"birth_date" binding:"required"`
	BirthTime  string  `json:"birth_time" binding:"required"`
	BirthPlace string  `json:"birth_place" binding:"required"`
	Latitude   float64 `json:"latitude" binding:"required"`
	Longitude  float64 `json:"longitude" binding:"required"`
	Timezone   string  `json:"timezone" binding:"required"`
}

// UpdateBirthInfo handles updating birth date/time/location.
func (h *AuthHandler) UpdateBirthInfo(c *gin.Context) {
	userID := mustUserID(c)

	var req BirthInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid birth_date format, expected YYYY-MM-DD"})
		return
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	user.BirthDate = birthDate
	user.BirthTime = req.BirthTime
	user.BirthPlace = req.BirthPlace
	user.Latitude = req.Latitude
	user.Longitude = req.Longitude
	user.Timezone = req.Timezone

	if err := h.authService.UpdateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update birth info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "birth info updated successfully", "user": profileView(user)})
}

// mustUserID extracts the authenticated user ID from the Gin context.
// Panics if the middleware did not set it — which is a programming error, not a runtime error.
func mustUserID(c *gin.Context) uuid.UUID {
	v, _ := c.Get("userID")
	return v.(uuid.UUID)
}

// userView returns a safe subset of user fields for registration/login responses.
func userView(u *database.User) gin.H {
	return gin.H{
		"id":    u.ID,
		"email": u.Email,
		"name":  u.Name,
	}
}

// profileView returns the full profile (no password).
func profileView(u *database.User) gin.H {
	return gin.H{
		"id":          u.ID,
		"email":       u.Email,
		"name":        u.Name,
		"birth_date":  u.BirthDate,
		"birth_time":  u.BirthTime,
		"birth_place": u.BirthPlace,
		"latitude":    u.Latitude,
		"longitude":   u.Longitude,
		"timezone":    u.Timezone,
	}
}
