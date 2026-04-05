package database

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserModel(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	if !db.Migrator().HasTable(&User{}) {
		t.Fatal("User table not created")
	}
	if !db.Migrator().HasColumn(&User{}, "email") {
		t.Fatal("email column not created")
	}
}

func TestBirthChartModel(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	err = db.AutoMigrate(&BirthChart{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	if !db.Migrator().HasTable(&BirthChart{}) {
		t.Fatal("BirthChart table not created")
	}
	if !db.Migrator().HasColumn(&BirthChart{}, "sun_sign") {
		t.Fatal("sun_sign column not created")
	}
}

func TestHoroscopeModel(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	err = db.AutoMigrate(&Horoscope{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	if !db.Migrator().HasTable(&Horoscope{}) {
		t.Fatal("Horoscope table not created")
	}
	if !db.Migrator().HasColumn(&Horoscope{}, "sign") {
		t.Fatal("sign column not created")
	}
}

func TestAllModels(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	err = db.AutoMigrate(&User{}, &BirthChart{}, &Horoscope{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	if !db.Migrator().HasTable(&User{}) {
		t.Fatal("User table not created")
	}
	if !db.Migrator().HasTable(&BirthChart{}) {
		t.Fatal("BirthChart table not created")
	}
	if !db.Migrator().HasTable(&Horoscope{}) {
		t.Fatal("Horoscope table not created")
	}
}
