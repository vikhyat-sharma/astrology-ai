package e2e

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"github.com/vikhyat-sharma/astrology-ai/internal/handlers"
	"github.com/vikhyat-sharma/astrology-ai/internal/repositories"
	"github.com/vikhyat-sharma/astrology-ai/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupInMemoryDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}, &database.BirthChart{}, &database.Horoscope{}); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}
	return db
}

func TestE2ECreateBirthChartAndGetDailyHoroscope(t *testing.T) {
	db := setupInMemoryDB(t)
	astrologyRepo := repositories.NewAstrologyRepository(db)
	astrologyService := services.NewAstrologyService(astrologyRepo, "", "")
	astrologyHandler := handlers.NewAstrologyHandler(astrologyService)

	r := gin.New()

	// Inject a fixed userID so the handler can read it from context.
	fixedUserID := uuid.New()
	injectUser := func(c *gin.Context) {
		c.Set("userID", fixedUserID)
		c.Next()
	}

	r.POST("/api/v1/astrology/birth-chart", injectUser, astrologyHandler.CreateBirthChart)
	r.GET("/api/v1/astrology/horoscope/daily/:sign", injectUser, astrologyHandler.GetHoroscope)

	// Create birth chart.
	payload := `{"birth_date":"1985-07-23","birth_time":"14:20","birth_place":"NYC","latitude":40.7128,"longitude":-74.0060,"timezone":"America/New_York"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/astrology/birth-chart", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Get daily horoscope.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/astrology/horoscope/daily/Leo", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "horoscope") {
		t.Fatalf("expected 'horoscope' key in response, got: %s", w.Body.String())
	}
}

func TestE2EGetBirthChartForbidden(t *testing.T) {
	db := setupInMemoryDB(t)
	astrologyRepo := repositories.NewAstrologyRepository(db)
	astrologyService := services.NewAstrologyService(astrologyRepo, "", "")
	astrologyHandler := handlers.NewAstrologyHandler(astrologyService)

	r := gin.New()

	ownerID := uuid.New()
	otherID := uuid.New()

	// Create a chart owned by ownerID.
	chart := &database.BirthChart{UserID: ownerID, SunSign: "Leo"}
	if err := astrologyRepo.CreateBirthChart(chart); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Try to fetch it as otherID — should get 403.
	r.GET("/api/v1/astrology/birth-chart/:id", func(c *gin.Context) {
		c.Set("userID", otherID)
		astrologyHandler.GetBirthChart(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/astrology/birth-chart/"+chart.ID.String(), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestE2EInvalidBirthDate(t *testing.T) {
	db := setupInMemoryDB(t)
	astrologyRepo := repositories.NewAstrologyRepository(db)
	astrologyService := services.NewAstrologyService(astrologyRepo, "", "")
	astrologyHandler := handlers.NewAstrologyHandler(astrologyService)

	r := gin.New()
	r.POST("/api/v1/astrology/birth-chart", func(c *gin.Context) {
		c.Set("userID", uuid.New())
		astrologyHandler.CreateBirthChart(c)
	})

	payload := `{"birth_date":"not-a-date","birth_time":"10:00","birth_place":"NYC","latitude":40.7,"longitude":-74.0,"timezone":"UTC"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/astrology/birth-chart", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid date, got %d: %s", w.Code, w.Body.String())
	}
}
