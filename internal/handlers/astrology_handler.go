package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	graphqlhandler "github.com/graphql-go/handler"
	"github.com/vikhyat-sharma/astrology-ai/internal/constants"
	"github.com/vikhyat-sharma/astrology-ai/internal/ports"
)

// AstrologyHandler handles astrology HTTP requests.
type AstrologyHandler struct {
	astrologyService ports.AstrologyService
	graphqlHandler   http.Handler
}

// NewAstrologyHandler creates a new AstrologyHandler.
// The GraphQL schema is built once at construction time — not per request.
func NewAstrologyHandler(astrologyService ports.AstrologyService) *AstrologyHandler {
	h := &AstrologyHandler{astrologyService: astrologyService}
	h.graphqlHandler = h.buildGraphQLHandler()
	return h
}

// CreateBirthChartRequest is the create birth chart request payload.
type CreateBirthChartRequest struct {
	BirthDate  string  `json:"birth_date" binding:"required"`
	BirthTime  string  `json:"birth_time" binding:"required"`
	BirthPlace string  `json:"birth_place" binding:"required"`
	Latitude   float64 `json:"latitude" binding:"required"`
	Longitude  float64 `json:"longitude" binding:"required"`
	Timezone   string  `json:"timezone" binding:"required"`
}

// CreateBirthChart handles birth chart creation.
func (h *AstrologyHandler) CreateBirthChart(c *gin.Context) {
	userID := mustUserID(c)

	var req CreateBirthChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chart, err := h.astrologyService.CreateBirthChart(c.Request.Context(), ports.BirthChartData{
		UserID:     userID,
		BirthDate:  req.BirthDate,
		BirthTime:  req.BirthTime,
		BirthPlace: req.BirthPlace,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		Timezone:   req.Timezone,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "birth chart created successfully",
		"chart":   chart,
	})
}

// GetBirthChart handles fetching a birth chart by ID.
func (h *AstrologyHandler) GetBirthChart(c *gin.Context) {
	userID := mustUserID(c)

	chartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chart ID"})
		return
	}

	chart, err := h.astrologyService.GetBirthChart(c.Request.Context(), chartID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if chart.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"chart": chart})
}

// GetHoroscope handles fetching a horoscope by sign and optional type.
func (h *AstrologyHandler) GetHoroscope(c *gin.Context) {
	sign := c.Param("sign")
	if sign == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sign parameter is required"})
		return
	}

	horoscopeType := c.Query("type")
	if horoscopeType == "" {
		horoscopeType = constants.HoroscopeTypeDaily
	}

	horoscope, err := h.astrologyService.GetHoroscope(c.Request.Context(), sign, horoscopeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"horoscope": horoscope})
}

// CheckCompatibilityRequest is the compatibility check request payload.
type CheckCompatibilityRequest struct {
	ChartID1 string `json:"chart_id_1" binding:"required"`
	ChartID2 string `json:"chart_id_2" binding:"required"`
}

// CheckCompatibility handles compatibility analysis between two birth charts.
func (h *AstrologyHandler) CheckCompatibility(c *gin.Context) {
	userID := mustUserID(c)

	var req CheckCompatibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chartID1, err := uuid.Parse(req.ChartID1)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chart_id_1"})
		return
	}
	chartID2, err := uuid.Parse(req.ChartID2)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chart_id_2"})
		return
	}

	// Authorisation: both charts must belong to the authenticated user.
	chart1, err := h.astrologyService.GetBirthChart(c.Request.Context(), chartID1)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart 1 not found"})
		return
	}
	chart2, err := h.astrologyService.GetBirthChart(c.Request.Context(), chartID2)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart 2 not found"})
		return
	}
	if chart1.UserID != userID || chart2.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	result, err := h.astrologyService.CheckCompatibility(c.Request.Context(), chartID1, chartID2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"compatibility": result})
}

// PersonalizedHoroscopeRequest is the personalized horoscope request payload.
type PersonalizedHoroscopeRequest struct {
	ChartID    string   `json:"chart_id" binding:"required"`
	Goals      string   `json:"goals,omitempty"`
	FocusAreas []string `json:"focus_areas,omitempty"`
	Tone       string   `json:"tone,omitempty"`
}

// GeneratePersonalizedHoroscope handles AI-personalized horoscope generation.
func (h *AstrologyHandler) GeneratePersonalizedHoroscope(c *gin.Context) {
	userID := mustUserID(c)

	var req PersonalizedHoroscopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chartID, err := uuid.Parse(req.ChartID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chart_id"})
		return
	}

	chart, err := h.astrologyService.GetBirthChart(c.Request.Context(), chartID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if chart.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	result, err := h.astrologyService.GeneratePersonalizedHoroscope(c.Request.Context(), chart, ports.PersonalizationPreferences{
		Goals:      req.Goals,
		FocusAreas: req.FocusAreas,
		Tone:       req.Tone,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"personalized_horoscope": result})
}

// GetRemedies handles fetching AI-generated remedies for a birth chart.
func (h *AstrologyHandler) GetRemedies(c *gin.Context) {
	userID := mustUserID(c)

	chartID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chart ID"})
		return
	}

	chart, err := h.astrologyService.GetBirthChart(c.Request.Context(), chartID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if chart.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	remedies, err := h.astrologyService.GetRemedies(c.Request.Context(), chart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"remedies": remedies})
}

// GetCurrentTransits handles real-time planetary transit queries.
func (h *AstrologyHandler) GetCurrentTransits(c *gin.Context) {
	lat, err := parseOptionalFloat(c.Query("latitude"), -90, 90)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid latitude: %v", err)})
		return
	}
	lng, err := parseOptionalFloat(c.Query("longitude"), -180, 180)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid longitude: %v", err)})
		return
	}

	transits, err := h.astrologyService.GetCurrentTransits(c.Request.Context(), lat, lng)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transits":  transits,
		"timestamp": time.Now().UTC(),
	})
}

// GraphQLHandler returns the pre-built GraphQL gin handler.
func (h *AstrologyHandler) GraphQLHandler() gin.HandlerFunc {
	return gin.WrapH(h.graphqlHandler)
}

// buildGraphQLHandler constructs the GraphQL schema and handler once at startup.
// Panics on schema build failure — this is a programming error, not a runtime error.
func (h *AstrologyHandler) buildGraphQLHandler() http.Handler {
	horoscopeType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Horoscope",
		Fields: graphql.Fields{
			"id":      &graphql.Field{Type: graphql.String},
			"sign":    &graphql.Field{Type: graphql.String},
			"type":    &graphql.Field{Type: graphql.String},
			"content": &graphql.Field{Type: graphql.String},
			"date":    &graphql.Field{Type: graphql.String},
		},
	})

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"horoscope": &graphql.Field{
				Type: horoscopeType,
				Args: graphql.FieldConfigArgument{
					"sign": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"type": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					sign, _ := p.Args["sign"].(string)
					htype, _ := p.Args["type"].(string)
					if htype == "" {
						htype = constants.HoroscopeTypeDaily
					}
					horoscope, err := h.astrologyService.GetHoroscope(p.Context, sign, htype)
					if err != nil {
						return nil, err
					}
					return map[string]interface{}{
						"id":      horoscope.ID.String(),
						"sign":    horoscope.Sign,
						"type":    horoscope.Type,
						"content": horoscope.Content,
						"date":    horoscope.Date.Format(time.RFC3339),
					}, nil
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{Query: rootQuery})
	if err != nil {
		panic(fmt.Sprintf("failed to build GraphQL schema: %v", err))
	}

	return graphqlhandler.New(&graphqlhandler.Config{
		Schema: &schema,
		Pretty: true,
	})
}

// parseOptionalFloat parses a query string as float64, returning 0.0 if empty.
// Returns an error if the value is present but invalid or out of the given range.
func parseOptionalFloat(s string, min, max float64) (float64, error) {
	if s == "" {
		return 0, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("must be a number")
	}
	if v < min || v > max {
		return 0, fmt.Errorf("must be between %.0f and %.0f", min, max)
	}
	return v, nil
}
