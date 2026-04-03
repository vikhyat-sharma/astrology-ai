package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"github.com/vikhyat-sharma/astrology-ai/internal/repositories"
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

func TestCreateAndGetBirthChartIntegration(t *testing.T) {
	db := setupInMemoryDB(t)
	astrologyRepo := repositories.NewAstrologyRepository(db)
	service := NewAstrologyService(astrologyRepo, "", "")

	userID := uuid.New()
	input := BirthChartData{
		UserID:     userID,
		BirthDate:  time.Date(1995, time.December, 15, 0, 0, 0, 0, time.UTC),
		BirthTime:  "10:30",
		BirthPlace: "Bangalore",
		Latitude:   12.9716,
		Longitude:  77.5946,
		Timezone:   "Asia/Kolkata",
	}

	chart, err := service.CreateBirthChart(input)
	if err != nil {
		t.Fatalf("CreateBirthChart failed: %v", err)
	}
	if chart.UserID != userID {
		t.Fatalf("CreateBirthChart saved wrong user id: got %v, want %v", chart.UserID, userID)
	}

	fetched, err := service.GetBirthChart(chart.ID)
	if err != nil {
		t.Fatalf("GetBirthChart failed: %v", err)
	}
	if fetched.ID != chart.ID {
		t.Fatalf("GetBirthChart returned wrong chart id: got %v, want %v", fetched.ID, chart.ID)
	}
}

func TestGetDailyHoroscopeIntegration(t *testing.T) {
	db := setupInMemoryDB(t)
	astrologyRepo := repositories.NewAstrologyRepository(db)
	service := NewAstrologyService(astrologyRepo, "", "")

	horoscope, err := service.GetDailyHoroscope("Gemini")
	if err != nil {
		t.Fatalf("GetDailyHoroscope failed: %v", err)
	}
	if horoscope.Sign != "Gemini" || horoscope.Type != "daily" {
		t.Fatalf("unexpected horoscope: %v", horoscope)
	}
	if horoscope.Content == "" {
		t.Fatal("expected non-empty content")
	}
}
