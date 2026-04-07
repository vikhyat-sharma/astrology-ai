package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/constants"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"github.com/vikhyat-sharma/astrology-ai/internal/interfaces"
	"github.com/vikhyat-sharma/astrology-ai/internal/repositories"
)

// AstrologyService handles astrology business logic
type AstrologyService struct {
	astrologyRepo interfaces.AstrologyRepositoryInterface
	ollamaURL     string
	ollamaModel   string
	httpClient    interfaces.HTTPClientInterface
}

// NewAstrologyService creates a new astrology service
func NewAstrologyService(astrologyRepo *repositories.AstrologyRepository, ollamaURL, ollamaModel string) *AstrologyService {
	return &AstrologyService{
		astrologyRepo: astrologyRepo,
		ollamaURL:     ollamaURL,
		ollamaModel:   ollamaModel,
		httpClient:    &http.Client{Timeout: constants.OllamaTimeoutSeconds * time.Second},
	}
}

// NewAstrologyServiceWithClient creates a new astrology service with custom HTTP client
func NewAstrologyServiceWithClient(astrologyRepo interfaces.AstrologyRepositoryInterface, ollamaURL, ollamaModel string, httpClient interfaces.HTTPClientInterface) *AstrologyService {
	return &AstrologyService{
		astrologyRepo: astrologyRepo,
		ollamaURL:     ollamaURL,
		ollamaModel:   ollamaModel,
		httpClient:    httpClient,
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
	horoscope, err := s.astrologyRepo.GetHoroscope(sign, constants.HoroscopeTypeDaily)
	if err == nil {
		return horoscope, nil
	}

	// If not found, generate a new one (in a real app, this might come from an AI service)
	horoscope = &database.Horoscope{
		Sign:         sign,
		Type:         constants.HoroscopeTypeDaily,
		Date:         time.Now().Truncate(24 * time.Hour),
		Content:      s.generateDailyHoroscope(sign),
		LoveRating:   constants.DefaultLoveRating,
		MoneyRating:  constants.DefaultMoneyRating,
		HealthRating: constants.DefaultHealthRating,
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

	endpoint := fmt.Sprintf("%s%s", s.ollamaURL, constants.OllamaPredictionsEndpoint)
	resp, err := s.httpClient.Post(endpoint, constants.ContentTypeJSON, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != constants.StatusOK && resp.StatusCode != constants.StatusCreated {
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
			return constants.Capricorn
		}
		return constants.Aquarius
	case 2: // February
		if day <= 18 {
			return constants.Aquarius
		}
		return constants.Pisces
	case 3: // March
		if day <= 20 {
			return constants.Pisces
		}
		return constants.Aries
	case 4: // April
		if day <= 19 {
			return constants.Aries
		}
		return constants.Taurus
	case 5: // May
		if day <= 20 {
			return constants.Taurus
		}
		return constants.Gemini
	case 6: // June
		if day <= 20 {
			return constants.Gemini
		}
		return constants.Cancer
	case 7: // July
		if day <= 22 {
			return constants.Cancer
		}
		return constants.Leo
	case 8: // August
		if day <= 22 {
			return constants.Leo
		}
		return constants.Virgo
	case 9: // September
		if day <= 22 {
			return constants.Virgo
		}
		return constants.Libra
	case 10: // October
		if day <= 22 {
			return constants.Libra
		}
		return constants.Scorpio
	case 11: // November
		if day <= 21 {
			return constants.Scorpio
		}
		return constants.Sagittarius
	case 12: // December
		if day <= 21 {
			return constants.Sagittarius
		}
		return constants.Capricorn
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
		constants.Aries: constants.ElementFiery, constants.Leo: constants.ElementFiery, constants.Sagittarius: constants.ElementFiery,
		constants.Taurus: constants.ElementEarthy, constants.Virgo: constants.ElementEarthy, constants.Capricorn: constants.ElementEarthy,
		constants.Gemini: constants.ElementAiry, constants.Libra: constants.ElementAiry, constants.Aquarius: constants.ElementAiry,
		constants.Cancer: constants.ElementWatery, constants.Scorpio: constants.ElementWatery, constants.Pisces: constants.ElementWatery,
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

// GetRemedies generates personalized remedies based on a birth chart
func (s *AstrologyService) GetRemedies(chart *database.BirthChart) (map[string]interface{}, error) {
	// Generate remedies using AI based on the birth chart
	prompt := fmt.Sprintf(`Based on this birth chart, provide personalized astrological remedies and recommendations:

Sun Sign: %s
Moon Sign: %s
Rising Sign: %s
Planets: %s
Houses: %s
Aspects: %s

Please provide specific remedies for:
1. Health and well-being
2. Career and financial success
3. Relationships and love
4. Spiritual growth
5. Gemstones and colors to wear
6. Mantras or affirmations
7. Daily practices or rituals

Make the remedies practical, positive, and tailored to this chart.`, chart.SunSign, chart.MoonSign, chart.RisingSign, chart.Planets, chart.Houses, chart.Aspects)

	remediesText, err := s.fetchOllamaPrediction(prompt)
	if err != nil {
		// Fallback remedies if AI is unavailable
		remediesText = s.generateFallbackRemedies(chart)
	}

	return map[string]interface{}{
		"chart_id": chart.ID,
		"sun_sign": chart.SunSign,
		"remedies": remediesText,
		"generated_at": time.Now(),
	}, nil
}

// generateFallbackRemedies provides basic remedies when AI is unavailable
func (s *AstrologyService) generateFallbackRemedies(chart *database.BirthChart) string {
	element := s.getSignElement(chart.SunSign)

	return fmt.Sprintf(`Based on your %s sun sign (%s element), here are some general remedies:

**Health & Well-being:**
- Practice daily meditation for 10-15 minutes
- Stay hydrated and maintain a balanced diet
- Regular exercise according to your energy levels

**Career & Finance:**
- Focus on building stability and patience
- Network with like-minded individuals
- Set realistic financial goals

**Relationships:**
- Communicate openly and honestly
- Practice active listening
- Show appreciation for your loved ones

**Spiritual Growth:**
- Read spiritual books or scriptures
- Connect with nature regularly
- Practice gratitude daily

**Gemstones & Colors:**
- Wear colors that resonate with your %s energy
- Consider gemstones like amethyst for protection

**Daily Practices:**
- Morning affirmations
- Evening reflection
- Maintain a positive mindset

Remember, these are general suggestions. Consult with a professional astrologer for personalized guidance.`, chart.SunSign, element, element)
}
