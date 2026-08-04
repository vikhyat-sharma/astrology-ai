package database

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// validDBName enforces a strict allowlist for database names to prevent SQL injection.
var validDBName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,62}$`)

// InitDB initializes the database connection, creates the database if needed, and runs migrations.
func InitDB(databaseURL string) *gorm.DB {
	dbName := extractDatabaseName(databaseURL)
	if dbName == "" {
		log.Fatal("could not extract database name from DATABASE_URL")
	}

	if err := createDatabaseIfNotExists(databaseURL, dbName); err != nil {
		log.Fatalf("failed to ensure database exists: %v", err)
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Configure connection pool.
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	log.Println("running database migrations")
	if err := db.AutoMigrate(
		&User{}, &BirthChart{}, &Horoscope{},
		&Dasha{}, &Compatibility{}, &Panchang{}, &Transit{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	log.Println("database connected and migrated successfully")
	return db
}

// extractDatabaseName extracts the database name from a PostgreSQL DSN.
func extractDatabaseName(databaseURL string) string {
	parts := strings.Split(databaseURL, "/")
	if len(parts) < 4 {
		return ""
	}
	dbPart := parts[len(parts)-1]
	if idx := strings.Index(dbPart, "?"); idx != -1 {
		dbPart = dbPart[:idx]
	}
	return dbPart
}

// createDatabaseIfNotExists creates the target database if it does not already exist.
// Uses a parameterized query for the existence check and an allowlist-validated name
// for the CREATE DATABASE statement (which cannot be parameterized in PostgreSQL).
func createDatabaseIfNotExists(databaseURL, dbName string) error {
	if !validDBName.MatchString(dbName) {
		return fmt.Errorf("invalid database name %q: must match ^[a-zA-Z_][a-zA-Z0-9_]{0,62}$", dbName)
	}

	// Connect to the maintenance database to check/create the target database.
	baseURL := strings.Replace(databaseURL, "/"+dbName, "/postgres", 1)
	// Strip query params from base URL for the maintenance connection.
	if idx := strings.Index(baseURL, "?"); idx != -1 {
		baseURL = baseURL[:idx]
	}

	db, err := sql.Open("postgres", baseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres maintenance db: %w", err)
	}
	defer db.Close()

	// Parameterized — safe against injection.
	var exists bool
	if err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName,
	).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if !exists {
		// dbName is allowlist-validated above; safe to interpolate.
		if _, err := db.Exec("CREATE DATABASE " + dbName); err != nil {
			return fmt.Errorf("failed to create database %q: %w", dbName, err)
		}
		log.Printf("database %q created", dbName)
	}

	return nil
}
