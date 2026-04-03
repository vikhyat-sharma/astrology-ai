package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq" // Import pq driver for sql.Open
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes the database connection and creates database/tables if they don't exist
func InitDB(databaseURL string) *gorm.DB {
	// Parse the database URL to extract database name
	dbName := extractDatabaseName(databaseURL)
	if dbName == "" {
		log.Fatal("Could not extract database name from DATABASE_URL")
	}

	// Create database if it doesn't exist
	if err := createDatabaseIfNotExists(databaseURL, dbName); err != nil {
		log.Fatal("Failed to create database:", err)
	}

	// Connect to the specific database
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate the schema - this creates tables if they don't exist
	log.Println("Running database migrations for tables: users, birth_charts, horoscopes")
	if err := db.AutoMigrate(&User{}, &BirthChart{}, &Horoscope{}); err != nil {
		log.Printf("AutoMigrate failed: %v", err)
		log.Println("Attempting manual table creation...")

		// Try to create tables manually if AutoMigrate fails
		if err := createTablesManually(db); err != nil {
			log.Fatal("Failed to create tables manually:", err)
		}
	}

	log.Println("Database connected and migrated successfully")
	return db
}

// createTablesManually creates tables manually if AutoMigrate fails
func createTablesManually(db *gorm.DB) error {
	// Create users table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			name TEXT,
			birth_date TIMESTAMP WITH TIME ZONE,
			birth_time TEXT,
			birth_place TEXT,
			latitude NUMERIC,
			longitude NUMERIC,
			timezone TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`).Error; err != nil {
		return err
	}

	// Create birth_charts table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS birth_charts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id),
			sun_sign TEXT,
			moon_sign TEXT,
			rising_sign TEXT,
			planets JSONB,
			houses JSONB,
			aspects JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`).Error; err != nil {
		return err
	}

	// Create horoscopes table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS horoscopes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			sign TEXT NOT NULL,
			type TEXT NOT NULL,
			date DATE NOT NULL,
			content TEXT,
			love_rating INTEGER,
			money_rating INTEGER,
			health_rating INTEGER,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`).Error; err != nil {
		return err
	}

	return nil
}

// extractDatabaseName extracts the database name from a PostgreSQL connection string
func extractDatabaseName(databaseURL string) string {
	// Simple parsing for postgres://user:pass@host:port/dbname
	parts := strings.Split(databaseURL, "/")
	if len(parts) < 4 {
		return ""
	}
	dbPart := parts[len(parts)-1]
	// Remove query parameters if any
	if idx := strings.Index(dbPart, "?"); idx != -1 {
		dbPart = dbPart[:idx]
	}
	return dbPart
}

// createDatabaseIfNotExists creates the database if it doesn't exist
func createDatabaseIfNotExists(databaseURL, dbName string) error {
	// Create connection string without database name
	baseURL := strings.Replace(databaseURL, "/"+dbName, "/postgres", 1)

	// Connect to postgres database
	db, err := sql.Open("postgres", baseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer db.Close()

	// Check if database exists
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", dbName)
	err = db.QueryRow(query).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	// Create database if it doesn't exist
	if !exists {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		log.Printf("Database '%s' created successfully", dbName)
	}

	return nil
}
