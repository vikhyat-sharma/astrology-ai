package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/constants"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"github.com/vikhyat-sharma/astrology-ai/internal/ports"
	"github.com/vikhyat-sharma/astrology-ai/internal/repositories"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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

func TestCreateAndGetBirthChartIntegration(t *testing.T) {
	db := setupInMemoryDB(t)
	repo := repositories.NewAstrologyRepository(db)
	svc := NewAstrologyService(repo, "", "")

	userID := uuid.New()
	chart, err := svc.CreateBirthChart(context.Background(), ports.BirthChartData{
		UserID:     userID,
		BirthDate:  "1995-12-15",
		BirthTime:  "10:30",
		BirthPlace: "Bangalore",
		Latitude:   12.9716,
		Longitude:  77.5946,
		Timezone:   "Asia/Kolkata",
	})
	if err != nil {
		t.Fatalf("CreateBirthChart failed: %v", err)
	}
	if chart.UserID != userID {
		t.Fatalf("wrong user id: got %v, want %v", chart.UserID, userID)
	}
	// Dec 15 1995 — Sun in Sagittarius.
	if chart.SunSign != "Sagittarius" {
		t.Fatalf("expected Sagittarius, got %s", chart.SunSign)
	}

	fetched, err := svc.GetBirthChart(context.Background(), chart.ID)
	if err != nil {
		t.Fatalf("GetBirthChart failed: %v", err)
	}
	if fetched.ID != chart.ID {
		t.Fatalf("wrong chart id: got %v, want %v", fetched.ID, chart.ID)
	}
}

func TestGetDailyHoroscopeIntegration(t *testing.T) {
	db := setupInMemoryDB(t)
	repo := repositories.NewAstrologyRepository(db)
	svc := NewAstrologyService(repo, "", "")

	horoscope, err := svc.GetHoroscope(context.Background(), constants.Gemini, constants.HoroscopeTypeDaily)
	if err != nil {
		t.Fatalf("GetHoroscope failed: %v", err)
	}
	if horoscope.Sign != constants.Gemini {
		t.Fatalf("expected Gemini, got %s", horoscope.Sign)
	}
	if horoscope.Type != constants.HoroscopeTypeDaily {
		t.Fatalf("expected daily, got %s", horoscope.Type)
	}
	if horoscope.Content == "" {
		t.Fatal("expected non-empty content")
	}
}

func TestGetHoroscopeIdempotentIntegration(t *testing.T) {
	// Calling GetHoroscope twice for the same sign/type/date must return the same record.
	db := setupInMemoryDB(t)
	repo := repositories.NewAstrologyRepository(db)
	svc := NewAstrologyService(repo, "", "")

	h1, err := svc.GetHoroscope(context.Background(), constants.Leo, constants.HoroscopeTypeDaily)
	if err != nil {
		t.Fatalf("first GetHoroscope failed: %v", err)
	}
	h2, err := svc.GetHoroscope(context.Background(), constants.Leo, constants.HoroscopeTypeDaily)
	if err != nil {
		t.Fatalf("second GetHoroscope failed: %v", err)
	}
	if h1.ID != h2.ID {
		t.Fatalf("expected same horoscope ID on second call, got %v vs %v", h1.ID, h2.ID)
	}
}

func TestBirthChartAscendantIntegration(t *testing.T) {
	db := setupInMemoryDB(t)
	repo := repositories.NewAstrologyRepository(db)
	svc := NewAstrologyService(repo, "", "")

	// Bangalore coordinates — ascendant must be a valid degree 0–360.
	chart, err := svc.CreateBirthChart(context.Background(), ports.BirthChartData{
		UserID:    uuid.New(),
		BirthDate: "1990-06-15",
		BirthTime: "06:00",
		Latitude:  12.9716,
		Longitude: 77.5946,
		Timezone:  "Asia/Kolkata",
	})
	if err != nil {
		t.Fatalf("CreateBirthChart failed: %v", err)
	}
	if chart.Ascendant < 0 || chart.Ascendant >= 360 {
		t.Fatalf("ascendant %f out of range 0–360", chart.Ascendant)
	}
}

func TestDashaServiceCalculation(t *testing.T) {
	db := setupInMemoryDB(t)
	repo := repositories.NewAstrologyRepository(db)

	// Create a chart to reference.
	chart := &database.BirthChart{
		UserID:   uuid.New(),
		MoonSign: "Aries",
	}
	if err := repo.CreateBirthChart(chart); err != nil {
		t.Fatalf("CreateBirthChart failed: %v", err)
	}

	dashaSvc := NewDashaService(repo)
	birthDate := time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC)
	periods, err := dashaSvc.CalculateVimshottariDasha(chart.ID, birthDate)
	if err != nil {
		t.Fatalf("CalculateVimshottariDasha failed: %v", err)
	}
	if len(periods) == 0 {
		t.Fatal("expected at least one dasha period")
	}
	// All periods must have non-empty mahadasha and antardasha names.
	for i, p := range periods {
		if p.Mahadasha == "" {
			t.Fatalf("period[%d] has empty Mahadasha", i)
		}
		if p.Antardasha == "" {
			t.Fatalf("period[%d] has empty Antardasha", i)
		}
		if !p.MahadashaEnd.After(p.MahadashaStart) {
			t.Fatalf("period[%d] MahadashaEnd not after MahadashaStart", i)
		}
	}
}
