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

func setupInMemoryDB(t *testing.T) *gorm.DB {
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
	r.POST("/api/v1/astrology/birth-chart", func(c *gin.Context) {
		c.Set("userID", uuid.New())
		astrologyHandler.CreateBirthChart(c)
	})
	r.GET("/api/v1/astrology/horoscope/daily", astrologyHandler.GetDailyHoroscope)

	payload := `{"birth_date":"1985-07-23","birth_time":"14:20","birth_place":"NYC","latitude":40.7128,"longitude":-74.0060,"timezone":"America/New_York"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/astrology/birth-chart", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 created, got %d, body=%s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/astrology/horoscope/daily?sign=Leo", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "horoscope") {
		t.Fatalf("expected horoscope key in response body, got %s", w.Body.String())
	}
}
