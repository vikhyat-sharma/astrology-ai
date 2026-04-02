package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"github.com/vikhyat-sharma/astrology-ai/internal/repositories"
)

// AstrologyService handles astrology business logic
type AstrologyService struct {
	astrologyRepo *repositories.AstrologyRepository
	ollamaURL     string
	ollamaModel   string
}

// NewAstrologyService creates a new astrology service
func NewAstrologyService(astrologyRepo *repositories.AstrologyRepository, ollamaURL, ollamaModel string) *AstrologyService {
	return &AstrologyService{
		astrologyRepo: astrologyRepo,
		ollamaURL:     ollamaURL,
		ollamaModel:   ollamaModel,
	}
}

// BirthChartData represents the data needed to create a birth chart
type BirthChartData struct {
	UserID     uuid.UUID `json:"user_id"`
	BirthDate  time.Time `json:"birth_date"`
	BirthTime  string    `json:"birth_time"`
	BirthPlace string    `json:"birth_place"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Timezone   string    `json:"timezone"`
}

// CreateBirthChart creates a new birth chart for a user
func (s *AstrologyService) CreateBirthChart(data BirthChartData) (*database.BirthChart, error) {
	// Calculate astrological data (simplified for now)
	sunSign := s.calculateSunSign(data.BirthDate)
	moonSign := s.calculateMoonSign(data.BirthDate, data.BirthTime)
	risingSign := s.calculateRisingSign(data.BirthDate, data.BirthTime, data.Latitude, data.Longitude)

	// Mock planet positions (in a real app, this would use astronomical calculations)
	planets := map[string]interface{}{
		"sun":     map[string]string{"sign": sunSign, "degree": "15°23'"},
		"moon":    map[string]string{"sign": moonSign, "degree": "7°45'"},
		"mercury": map[string]string{"sign": "Virgo", "degree": "22°10'"},
		"venus":   map[string]string{"sign": "Libra", "degree": "3°30'"},
		"mars":    map[string]string{"sign": "Sagittarius", "degree": "18°15'"},
	}

	planetsJSON, _ := json.Marshal(planets)

	// Mock house cusps
	houses := map[string]string{
		"1":  "0° Aries",
		"2":  "30° Aries",
		"3":  "60° Taurus",
		"4":  "90° Gemini",
		"5":  "120° Cancer",
		"6":  "150° Leo",
		"7":  "180° Virgo",
		"8":  "210° Libra",
		"9":  "240° Scorpio",
		"10": "270° Sagittarius",
		"11": "300° Capricorn",
		"12": "330° Aquarius",
	}

	housesJSON, _ := json.Marshal(houses)

	// Mock aspects
	aspects := []map[string]interface{}{
		{"planet1": "sun", "planet2": "moon", "aspect": "trine", "orb": 2.5},
		{"planet1": "venus", "planet2": "mars", "aspect": "conjunction", "orb": 1.2},
	}

	aspectsJSON, _ := json.Marshal(aspects)

	chart := &database.BirthChart{
		UserID:     data.UserID,
		SunSign:    sunSign,
		MoonSign:   moonSign,
		RisingSign: risingSign,
		Planets:    string(planetsJSON),
		Houses:     string(housesJSON),
		Aspects:    string(aspectsJSON),
	}

	if err := s.astrologyRepo.CreateBirthChart(chart); err != nil {
		return nil, err
	}

	return chart, nil
}

// GetBirthChart gets a birth chart by ID
func (s *AstrologyService) GetBirthChart(id uuid.UUID) (*database.BirthChart, error) {
	return s.astrologyRepo.GetBirthChart(id)
}

// GetUserBirthCharts gets all birth charts for a user
func (s *AstrologyService) GetUserBirthCharts(userID uuid.UUID) ([]*database.BirthChart, error) {
	return s.astrologyRepo.GetBirthChartsByUserID(userID)
}

// GetDailyHoroscope gets the daily horoscope for a sign
func (s *AstrologyService) GetDailyHoroscope(sign string) (*database.Horoscope, error) {
	// Check if today's horoscope exists
	horoscope, err := s.astrologyRepo.GetHoroscope(sign, "daily")
	if err == nil {
		return horoscope, nil
	}

	// If not found, generate a new one (in a real app, this might come from an AI service)
	horoscope = &database.Horoscope{
		Sign:         sign,
		Type:         "daily",
		Date:         time.Now().Truncate(24 * time.Hour),
		Content:      s.generateDailyHoroscope(sign),
		LoveRating:   7,
		MoneyRating:  6,
		HealthRating: 8,
	}

	if err := s.astrologyRepo.CreateHoroscope(horoscope); err != nil {
		return nil, err
	}

	return horoscope, nil
}

// fetchOllamaPrediction calls Ollama /api/predictions for advice text
func (s *AstrologyService) fetchOllamaPrediction(prompt string) (string, error) {
	if s.ollamaURL == "" || s.ollamaModel == "" {
		return "", fmt.Errorf("ollama config missing")
	}

	reqBody := map[string]interface{}{
		"model":  s.ollamaModel,
		"prompt": prompt,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("%s/api/predictions", s.ollamaURL)
	resp, err := http.Post(endpoint, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error: %d %s", resp.StatusCode, string(respBytes))
	}

	var parsed struct {
		Status string      `json:"status"`
		Output interface{} `json:"output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}

	// output may be array or string
	if parsed.Output == nil {
		return "", fmt.Errorf("ollama output empty")
	}

	// Prefer string output
	if sOut, ok := parsed.Output.(string); ok {
		return sOut, nil
	}

	if arrOut, ok := parsed.Output.([]interface{}); ok && len(arrOut) > 0 {
		if sOut, ok := arrOut[0].(string); ok {
			return sOut, nil
		}
	}

	return "", fmt.Errorf("unexpected ollama output format")
}

func (s *AstrologyService) generateDailyHoroscope(sign string) string {
	prompt := fmt.Sprintf("Write a friendly and concise daily horoscope for %s. Include advice for love, money, and health, with wise but realistic language.", sign)
	if text, err := s.fetchOllamaPrediction(prompt); err == nil && text != "" {
		return text
	}

	// fallback if Ollama is unavailable
	return fmt.Sprintf("Today brings new opportunities for %s. Trust your intuition and embrace change. Your natural %s energy will guide you to success.", sign, s.getSignElement(sign))
}

// CheckCompatibility checks compatibility between two birth charts
func (s *AstrologyService) CheckCompatibility(chartID1, chartID2 uuid.UUID) (map[string]interface{}, error) {
	chart1, err := s.astrologyRepo.GetBirthChart(chartID1)
	if err != nil {
		return nil, err
	}

	chart2, err := s.astrologyRepo.GetBirthChart(chartID2)
	if err != nil {
		return nil, err
	}

	// Simple compatibility calculation (in a real app, this would be more complex)
	compatibility := s.calculateCompatibility(chart1, chart2)

	return map[string]interface{}{
		"chart1":        chart1,
		"chart2":        chart2,
		"compatibility": compatibility,
	}, nil
}

// Helper functions for astrological calculations

func (s *AstrologyService) calculateSunSign(birthDate time.Time) string {
	month := birthDate.Month()
	day := birthDate.Day()

	switch month {
	case 1: // January
		if day <= 19 {
			return "Capricorn"
		}
		return "Aquarius"
	case 2: // February
		if day <= 18 {
			return "Aquarius"
		}
		return "Pisces"
	case 3: // March
		if day <= 20 {
			return "Pisces"
		}
		return "Aries"
	case 4: // April
		if day <= 19 {
			return "Aries"
		}
		return "Taurus"
	case 5: // May
		if day <= 20 {
			return "Taurus"
		}
		return "Gemini"
	case 6: // June
		if day <= 20 {
			return "Gemini"
		}
		return "Cancer"
	case 7: // July
		if day <= 22 {
			return "Cancer"
		}
		return "Leo"
	case 8: // August
		if day <= 22 {
			return "Leo"
		}
		return "Virgo"
	case 9: // September
		if day <= 22 {
			return "Virgo"
		}
		return "Libra"
	case 10: // October
		if day <= 22 {
			return "Libra"
		}
		return "Scorpio"
	case 11: // November
		if day <= 21 {
			return "Scorpio"
		}
		return "Sagittarius"
	case 12: // December
		if day <= 21 {
			return "Sagittarius"
		}
		return "Capricorn"
	}
	return "Unknown"
}

func (s *AstrologyService) calculateMoonSign(birthDate time.Time, birthTime string) string {
	// Simplified moon sign calculation - in reality this requires astronomical calculations
	return "Cancer" // Placeholder
}

func (s *AstrologyService) calculateRisingSign(birthDate time.Time, birthTime string, lat, lng float64) string {
	// Simplified rising sign calculation
	return "Leo" // Placeholder
}

func (s *AstrologyService) getSignElement(sign string) string {
	elements := map[string]string{
		"Aries": "fiery", "Leo": "fiery", "Sagittarius": "fiery",
		"Taurus": "earthy", "Virgo": "earthy", "Capricorn": "earthy",
		"Gemini": "airy", "Libra": "airy", "Aquarius": "airy",
		"Cancer": "watery", "Scorpio": "watery", "Pisces": "watery",
	}
	return elements[sign]
}

func (s *AstrologyService) calculateCompatibility(chart1, chart2 *database.BirthChart) map[string]interface{} {
	// Simple compatibility calculation based on sun signs
	sunSigns := []string{chart1.SunSign, chart2.SunSign}

	// Mock compatibility score
	score := 75 // Placeholder

	return map[string]interface{}{
		"score":     score,
		"summary":   "These signs have good compatibility potential",
		"sun_signs": sunSigns,
	}
}
