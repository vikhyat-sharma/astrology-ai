package repositories

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}, &database.BirthChart{}, &database.Horoscope{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestCreateBirthChart(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAstrologyRepository(db)

	chart := &database.BirthChart{UserID: uuid.New(), SunSign: "Leo"}
	err := repo.CreateBirthChart(chart)

	if err != nil {
		t.Fatalf("CreateBirthChart failed: %v", err)
	}
	if chart.ID == uuid.Nil {
		t.Fatal("expected chart ID to be set")
	}
}

func TestGetBirthChart(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAstrologyRepository(db)

	chart := &database.BirthChart{UserID: uuid.New(), SunSign: "Leo"}
	repo.CreateBirthChart(chart)

	retrieved, err := repo.GetBirthChart(chart.ID)
	if err != nil {
		t.Fatalf("GetBirthChart failed: %v", err)
	}
	if retrieved.SunSign != "Leo" {
		t.Fatalf("expected Leo, got %s", retrieved.SunSign)
	}
}

func TestCreateHoroscope(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAstrologyRepository(db)

	h := &database.Horoscope{
		Sign:    "Leo",
		Type:    "daily",
		Date:    time.Now().Truncate(24 * time.Hour),
		Content: "Test",
	}
	err := repo.CreateHoroscope(h)

	if err != nil {
		t.Fatalf("CreateHoroscope failed: %v", err)
	}
	if h.ID == uuid.Nil {
		t.Fatal("expected horoscope ID to be set")
	}
}
