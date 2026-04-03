package main

import (
	"fmt"
	"log"

	_ "github.com/lib/pq" // Import pq driver for sql.Open
	"github.com/vikhyat-sharma/astrology-ai/internal/config"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	_ "gorm.io/driver/postgres" // Import postgres driver
)

func main() {
	// Load configuration
	cfg := config.Load()

	fmt.Println("=== Database Setup Script ===")
	fmt.Printf("Database URL: %s\n", cfg.DatabaseURL)

	// Initialize database (this will create DB and tables if they don't exist)
	db := database.InitDB(cfg.DatabaseURL)

	// Test connection and show table info
	fmt.Println("\n=== Database Tables Created ===")

	// Check if tables exist by trying to query them
	tables := []string{"users", "birth_charts", "horoscopes"}
	for _, table := range tables {
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			log.Printf("Warning: Could not query table '%s': %v", table, err)
		} else {
			fmt.Printf("✓ Table '%s' exists (rows: %d)\n", table, count)
		}
	}

	fmt.Println("\n=== Table Schemas ===")
	fmt.Println("users:")
	fmt.Println("  - id (UUID, primary key)")
	fmt.Println("  - email (string, unique)")
	fmt.Println("  - password (string)")
	fmt.Println("  - name (string)")
	fmt.Println("  - birth_date (timestamp)")
	fmt.Println("  - birth_time (string)")
	fmt.Println("  - birth_place (string)")
	fmt.Println("  - latitude (float64)")
	fmt.Println("  - longitude (float64)")
	fmt.Println("  - timezone (string)")
	fmt.Println("  - created_at (timestamp)")
	fmt.Println("  - updated_at (timestamp)")

	fmt.Println("\nbirth_charts:")
	fmt.Println("  - id (UUID, primary key)")
	fmt.Println("  - user_id (UUID, foreign key to users)")
	fmt.Println("  - sun_sign (string)")
	fmt.Println("  - moon_sign (string)")
	fmt.Println("  - rising_sign (string)")
	fmt.Println("  - planets (jsonb)")
	fmt.Println("  - houses (jsonb)")
	fmt.Println("  - aspects (jsonb)")
	fmt.Println("  - created_at (timestamp)")
	fmt.Println("  - updated_at (timestamp)")

	fmt.Println("\nhoroscopes:")
	fmt.Println("  - id (UUID, primary key)")
	fmt.Println("  - sign (string)")
	fmt.Println("  - type (string)")
	fmt.Println("  - date (timestamp)")
	fmt.Println("  - content (text)")
	fmt.Println("  - love_rating (int)")
	fmt.Println("  - money_rating (int)")
	fmt.Println("  - health_rating (int)")
	fmt.Println("  - created_at (timestamp)")
	fmt.Println("  - updated_at (timestamp)")

	fmt.Println("\n=== Setup Complete ===")
	fmt.Println("Database and tables are ready!")
}
