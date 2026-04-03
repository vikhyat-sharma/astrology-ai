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
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database connected and migrated successfully")
	return db
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
