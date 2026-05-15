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
	astrologyRepo        interfaces.AstrologyRepositoryInterface
	calculationService   *CalculationService
	compatibilityService *CompatibilityService
	dashaService         *DashaService
	ollamaURL            string
	ollamaModel          string
	httpClient           interfaces.HTTPClientInterface
}

// NewAstrologyService creates a new astrology service
func NewAstrologyService(astrologyRepo *repositories.AstrologyRepository, ollamaURL, ollamaModel string) *AstrologyService {
	return &AstrologyService{
		astrologyRepo:        astrologyRepo,
		calculationService:   NewCalculationService(),
		compatibilityService: NewCompatibilityService(astrologyRepo),
		dashaService:         NewDashaService(astrologyRepo),
		ollamaURL:            ollamaURL,
		ollamaModel:          ollamaModel,
		httpClient:           &http.Client{Timeout: constants.OllamaTimeoutSeconds * time.Second},
	}
}

// NewAstrologyServiceWithClient creates a new astrology service with custom HTTP client
func NewAstrologyServiceWithClient(astrologyRepo interfaces.AstrologyRepositoryInterface, ollamaURL, ollamaModel string, httpClient interfaces.HTTPClientInterface) *AstrologyService {
	return &AstrologyService{
		astrologyRepo:        astrologyRepo,
		calculationService:   NewCalculationService(),
		compatibilityService: NewCompatibilityService(astrologyRepo),
		dashaService:         NewDashaService(astrologyRepo),
		ollamaURL:            ollamaURL,
		ollamaModel:          ollamaModel,
		httpClient:           httpClient,
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
	// Calculate accurate astrological data
	chartData, err := s.calculationService.CalculateBirthChart(
		data.BirthDate,
		s.parseBirthTime(data.BirthTime),
		data.Latitude,
		data.Longitude,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate birth chart: %w", err)
	}

	// Convert to JSON for storage
	planetsJSON, _ := json.Marshal(chartData.Planets)
	housesJSON, _ := json.Marshal(chartData.Houses)
	aspectsJSON, _ := json.Marshal(chartData.Aspects)
	yogasJSON, _ := json.Marshal(chartData.Yogas)

	chart := &database.BirthChart{
		UserID:            data.UserID,
		SunSign:           chartData.SunSign,
		MoonSign:          chartData.MoonSign,
		RisingSign:        chartData.RisingSign,
		Nakshatra:         chartData.Nakshatra,
		NakshatraPad:      chartData.NakshatraPad,
		Ayanamsha:         0, // TODO: Implement Ayanamsha calculation
		Ascendant:         chartData.Ascendant,
		Midheaven:         chartData.Midheaven,
		Planets:           string(planetsJSON),
		Houses:            string(housesJSON),
		Aspects:           string(aspectsJSON),
		Yogas:             string(yogasJSON),
		CalculationMethod: "Placidus",
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

// GetHoroscope gets horoscope for a sign and type
func (s *AstrologyService) GetHoroscope(sign, horoscopeType string) (*database.Horoscope, error) {
	// Check if today's horoscope exists
	horoscope, err := s.astrologyRepo.GetHoroscope(sign, horoscopeType)
	if err == nil {
		return horoscope, nil
	}

	// If not found, generate a new one
	horoscope = &database.Horoscope{
		Sign:         sign,
		Type:         horoscopeType,
		Date:         time.Now().Truncate(24 * time.Hour),
		Content:      s.generateHoroscope(sign, horoscopeType),
		LoveRating:   constants.DefaultLoveRating,
		MoneyRating:  constants.DefaultMoneyRating,
		HealthRating: constants.DefaultHealthRating,
	}

	if err := s.astrologyRepo.CreateHoroscope(horoscope); err != nil {
		return nil, err
	}

	return horoscope, nil
}

// GetDailyHoroscope gets the daily horoscope for a sign (backward compatibility)
func (s *AstrologyService) GetDailyHoroscope(sign string) (*database.Horoscope, error) {
	return s.GetHoroscope(sign, constants.HoroscopeTypeDaily)
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

func (s *AstrologyService) generateHoroscope(sign, horoscopeType string) string {
	var prompt string

	switch horoscopeType {
	case constants.HoroscopeTypeDaily:
		prompt = fmt.Sprintf("Write a concise daily horoscope for %s. Include advice for love, money, and health, with realistic language.", sign)
	case constants.HoroscopeTypeWeekly:
		prompt = fmt.Sprintf("Write a detailed weekly horoscope for %s. Cover career, relationships, health, and spiritual growth for the coming week.", sign)
	case constants.HoroscopeTypeMonthly:
		prompt = fmt.Sprintf("Write a comprehensive monthly horoscope for %s. Include major themes, opportunities, challenges, and advice for the entire month.", sign)
	case constants.HoroscopeTypeYearly:
		prompt = fmt.Sprintf("Write an annual horoscope for %s. Cover major life areas including career, relationships, health, and personal growth for the year ahead.", sign)
	case constants.HoroscopeTypeLove:
		prompt = fmt.Sprintf("Write a focused love and relationship horoscope for %s. Include romantic opportunities, relationship advice, and emotional guidance.", sign)
	default:
		prompt = fmt.Sprintf("Write a friendly and concise daily horoscope for %s. Include advice for love, money, and health.", sign)
	}

	if text, err := s.fetchOllamaPrediction(prompt); err == nil && text != "" {
		return text
	}

	// fallback if Ollama is unavailable
	return s.generateFallbackHoroscope(sign, horoscopeType)
}

func (s *AstrologyService) generateFallbackHoroscope(sign, horoscopeType string) string {
	element := s.getSignElement(sign)

	switch horoscopeType {
	case constants.HoroscopeTypeWeekly:
		return fmt.Sprintf("This week brings new opportunities for %s. Trust your intuition and embrace change. Your natural %s energy will guide you to success in relationships and career.", sign, element)
	case constants.HoroscopeTypeMonthly:
		return fmt.Sprintf("This month focuses on growth and stability for %s. Pay attention to your health and relationships. Your %s nature will help you navigate challenges successfully.", sign, element)
	case constants.HoroscopeTypeYearly:
		return fmt.Sprintf("This year promises growth and new beginnings for %s. Focus on building strong foundations in career and relationships. Your %s energy will bring success through perseverance.", sign, element)
	case constants.HoroscopeTypeLove:
		return fmt.Sprintf("Love brings warmth and connection for %s this week. Open your heart to new possibilities. Your %s nature attracts meaningful relationships.", sign, element)
	default: // daily
		return fmt.Sprintf("Today brings new opportunities for %s. Trust your intuition and embrace change. Your natural %s energy will guide you to success.", sign, s.getSignElement(sign))
	}
}

// CheckCompatibility checks compatibility between two birth charts
func (s *AstrologyService) CheckCompatibility(chartID1, chartID2 uuid.UUID) (map[string]interface{}, error) {
	result, err := s.compatibilityService.CheckCompatibility(chartID1, chartID2)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"overall_score":      result.OverallScore,
		"varna_score":        result.VarnaScore,
		"vashya_score":       result.VashyaScore,
		"tara_score":         result.TaraScore,
		"yoni_score":         result.YoniScore,
		"graha_maitri_score": result.GrahaMaitriScore,
		"gana_score":         result.GanaScore,
		"bhakut_score":       result.BhakutScore,
		"nadi_score":         result.NadiScore,
		"analysis":           result.Analysis,
		"chart1":             result.Chart1,
		"chart2":             result.Chart2,
	}, nil
}

// parseBirthTime parses birth time string to time.Time
func (s *AstrologyService) parseBirthTime(birthTime string) time.Time {
	// Parse time in HH:MM format
	parsed, err := time.Parse("15:04", birthTime)
	if err != nil {
		// Default to noon if parsing fails
		return time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC)
	}
	return parsed
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
		"chart_id":     chart.ID,
		"sun_sign":     chart.SunSign,
		"remedies":     remediesText,
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
