package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/constants"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"github.com/vikhyat-sharma/astrology-ai/internal/interfaces"
	"github.com/vikhyat-sharma/astrology-ai/internal/ports"
	"github.com/vikhyat-sharma/astrology-ai/internal/repositories"
)

// AstrologyService implements ports.AstrologyService.
type AstrologyService struct {
	astrologyRepo        interfaces.AstrologyRepositoryInterface
	calculationService   *CalculationService
	compatibilityService *CompatibilityService
	dashaService         *DashaService
	ollamaURL            string
	ollamaModel          string
	httpClient           interfaces.HTTPClientInterface
}

// Compile-time assertion that AstrologyService satisfies the port interface.
var _ interface {
	CreateBirthChart(ctx context.Context, data ports.BirthChartData) (*database.BirthChart, error)
} = (*AstrologyService)(nil)

// newOllamaHTTPClient returns an http.Client with production-grade transport settings.
func newOllamaHTTPClient() *http.Client {
	return &http.Client{
		Timeout: time.Duration(constants.OllamaTimeoutSeconds) * time.Second,
		Transport: &http.Transport{
			DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 15 * time.Second,
			IdleConnTimeout:       90 * time.Second,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
		},
	}
}

// NewAstrologyService creates a production AstrologyService.
func NewAstrologyService(astrologyRepo *repositories.AstrologyRepository, ollamaURL, ollamaModel string) *AstrologyService {
	return &AstrologyService{
		astrologyRepo:        astrologyRepo,
		calculationService:   NewCalculationService(),
		compatibilityService: NewCompatibilityService(astrologyRepo),
		dashaService:         NewDashaService(astrologyRepo),
		ollamaURL:            ollamaURL,
		ollamaModel:          ollamaModel,
		httpClient:           newOllamaHTTPClient(),
	}
}

// NewAstrologyServiceWithClient creates an AstrologyService with a custom HTTP client (for testing).
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

// PersonalizationPreferences carries optional personalization inputs.
// Kept here for backward compatibility with existing tests; ports.PersonalizationPreferences
// is the canonical definition used by handlers.
type PersonalizationPreferences = ports.PersonalizationPreferences

// BirthChartData is an alias to the canonical ports type.
type BirthChartData = ports.BirthChartData

// CreateBirthChart calculates and persists a birth chart.
func (s *AstrologyService) CreateBirthChart(ctx context.Context, data ports.BirthChartData) (*database.BirthChart, error) {
	birthDate, err := time.Parse(constants.DateFormat, data.BirthDate)
	if err != nil {
		return nil, fmt.Errorf("invalid birth_date %q: expected YYYY-MM-DD", data.BirthDate)
	}

	birthTime, err := parseBirthTime(data.BirthTime)
	if err != nil {
		return nil, err
	}

	chartData, err := s.calculationService.CalculateBirthChart(birthDate, birthTime, data.Latitude, data.Longitude)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate birth chart: %w", err)
	}

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
		Ascendant:         chartData.Ascendant,
		Midheaven:         chartData.Midheaven,
		Planets:           string(planetsJSON),
		Houses:            string(housesJSON),
		Aspects:           string(aspectsJSON),
		Yogas:             string(yogasJSON),
		CalculationMethod: "Equal",
	}

	if err := s.astrologyRepo.CreateBirthChart(chart); err != nil {
		return nil, err
	}
	return chart, nil
}

// GetBirthChart retrieves a birth chart by ID.
func (s *AstrologyService) GetBirthChart(ctx context.Context, id uuid.UUID) (*database.BirthChart, error) {
	return s.astrologyRepo.GetBirthChart(id)
}

// GetUserBirthCharts retrieves all birth charts for a user.
func (s *AstrologyService) GetUserBirthCharts(ctx context.Context, userID uuid.UUID) ([]*database.BirthChart, error) {
	return s.astrologyRepo.GetBirthChartsByUserID(userID)
}

// GetHoroscope returns today's horoscope for a sign, generating and persisting one if absent.
// Uses GetOrCreateHoroscope to avoid the check-then-insert race condition.
func (s *AstrologyService) GetHoroscope(ctx context.Context, sign, horoscopeType string) (*database.Horoscope, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	candidate := &database.Horoscope{
		Sign:         sign,
		Type:         horoscopeType,
		Date:         today,
		Content:      s.generateHoroscope(sign, horoscopeType),
		LoveRating:   constants.DefaultLoveRating,
		MoneyRating:  constants.DefaultMoneyRating,
		HealthRating: constants.DefaultHealthRating,
	}
	return s.astrologyRepo.GetOrCreateHoroscope(candidate)
}

// GetDailyHoroscope is a convenience wrapper for backward compatibility.
func (s *AstrologyService) GetDailyHoroscope(sign string) (*database.Horoscope, error) {
	return s.GetHoroscope(context.Background(), sign, constants.HoroscopeTypeDaily)
}

// CheckCompatibility performs Guna Milan analysis between two birth charts.
func (s *AstrologyService) CheckCompatibility(ctx context.Context, chartID1, chartID2 uuid.UUID) (map[string]interface{}, error) {
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

// GetRemedies generates personalized remedies via AI with a deterministic fallback.
func (s *AstrologyService) GetRemedies(ctx context.Context, chart *database.BirthChart) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Based on this birth chart, provide personalized astrological remedies:

Sun Sign: %s | Moon Sign: %s | Rising Sign: %s
Planets: %s
Houses: %s
Aspects: %s

Provide specific remedies for: health, career, relationships, spiritual growth, gemstones, mantras, daily practices.`,
		chart.SunSign, chart.MoonSign, chart.RisingSign,
		chart.Planets, chart.Houses, chart.Aspects)

	remediesText, err := s.fetchOllamaPrediction(prompt)
	if err != nil {
		remediesText = s.generateFallbackRemedies(chart)
	}

	return map[string]interface{}{
		"chart_id":     chart.ID,
		"sun_sign":     chart.SunSign,
		"remedies":     remediesText,
		"generated_at": time.Now().UTC(),
	}, nil
}

// GeneratePersonalizedHoroscope generates a chart-specific horoscope via AI.
func (s *AstrologyService) GeneratePersonalizedHoroscope(ctx context.Context, chart *database.BirthChart, prefs ports.PersonalizationPreferences) (map[string]interface{}, error) {
	prompt := s.buildPersonalizedHoroscopePrompt(chart, prefs)
	text, err := s.fetchOllamaPrediction(prompt)
	if err != nil {
		text = s.generatePersonalizedHoroscopeFallback(chart, prefs)
	}
	return map[string]interface{}{
		"chart_id":                chart.ID,
		"sun_sign":                chart.SunSign,
		"personalized_horoscope":  text,
		"personalization_details": prefs,
		"generated_at":            time.Now().UTC(),
	}, nil
}

// GetCurrentTransits returns current planetary positions for a given location.
func (s *AstrologyService) GetCurrentTransits(ctx context.Context, lat, lng float64) ([]map[string]interface{}, error) {
	now := time.Now().UTC()
	jd := s.calculationService.calculateJulianDay(now)
	planets, err := s.calculationService.calculatePlanetPositions(jd)
	if err != nil {
		return nil, err
	}
	houses := s.calculationService.calculateHouses(0) // ascendant=0 for generic transits
	if lat != 0 || lng != 0 {
		asc, err := s.calculationService.calculateAscendant(jd, lat, lng)
		if err == nil {
			houses = s.calculationService.calculateHouses(asc)
		}
	}
	planets = s.calculationService.assignPlanetsToHouses(planets, houses)

	result := make([]map[string]interface{}, len(planets))
	for i, p := range planets {
		result[i] = map[string]interface{}{
			"planet":      p.Name,
			"sign":        p.Sign,
			"degree":      p.Degree,
			"longitude":   p.Longitude,
			"house":       p.House,
			"retrograde":  p.Retrograde,
			"exalted":     p.Exalted,
			"debilitated": p.Debilitated,
			"own_sign":    p.OwnSign,
		}
	}
	return result, nil
}

// fetchOllamaPrediction calls the Ollama /api/predictions endpoint.
func (s *AstrologyService) fetchOllamaPrediction(prompt string) (string, error) {
	if s.ollamaURL == "" || s.ollamaModel == "" {
		return "", fmt.Errorf("ollama not configured")
	}

	body, err := json.Marshal(map[string]interface{}{
		"model":  s.ollamaModel,
		"prompt": prompt,
	})
	if err != nil {
		return "", err
	}

	endpoint := s.ollamaURL + constants.OllamaPredictionsEndpoint
	resp, err := s.httpClient.Post(endpoint, constants.ContentTypeJSON, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned %d: %s", resp.StatusCode, string(b))
	}

	var parsed struct {
		Output interface{} `json:"output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %w", err)
	}
	if parsed.Output == nil {
		return "", fmt.Errorf("ollama output is empty")
	}
	if s, ok := parsed.Output.(string); ok {
		return s, nil
	}
	if arr, ok := parsed.Output.([]interface{}); ok && len(arr) > 0 {
		if s, ok := arr[0].(string); ok {
			return s, nil
		}
	}
	return "", fmt.Errorf("unexpected ollama output format")
}

func (s *AstrologyService) generateHoroscope(sign, horoscopeType string) string {
	prompts := map[string]string{
		constants.HoroscopeTypeDaily:   fmt.Sprintf("Write a concise daily horoscope for %s. Include advice for love, money, and health.", sign),
		constants.HoroscopeTypeWeekly:  fmt.Sprintf("Write a detailed weekly horoscope for %s covering career, relationships, health, and spiritual growth.", sign),
		constants.HoroscopeTypeMonthly: fmt.Sprintf("Write a comprehensive monthly horoscope for %s including major themes, opportunities, and challenges.", sign),
		constants.HoroscopeTypeYearly:  fmt.Sprintf("Write an annual horoscope for %s covering career, relationships, health, and personal growth.", sign),
		constants.HoroscopeTypeLove:    fmt.Sprintf("Write a love and relationship horoscope for %s with romantic opportunities and emotional guidance.", sign),
	}
	prompt, ok := prompts[horoscopeType]
	if !ok {
		prompt = fmt.Sprintf("Write a friendly daily horoscope for %s.", sign)
	}

	if text, err := s.fetchOllamaPrediction(prompt); err == nil && text != "" {
		return text
	}
	return s.generateFallbackHoroscope(sign, horoscopeType)
}

func (s *AstrologyService) generateFallbackHoroscope(sign, horoscopeType string) string {
	element := s.getSignElement(sign)
	switch horoscopeType {
	case constants.HoroscopeTypeWeekly:
		return fmt.Sprintf("This week brings new opportunities for %s. Your natural %s energy will guide you to success.", sign, element)
	case constants.HoroscopeTypeMonthly:
		return fmt.Sprintf("This month focuses on growth for %s. Your %s nature will help you navigate challenges.", sign, element)
	case constants.HoroscopeTypeYearly:
		return fmt.Sprintf("This year promises growth for %s. Your %s energy will bring success through perseverance.", sign, element)
	case constants.HoroscopeTypeLove:
		return fmt.Sprintf("Love brings warmth for %s. Your %s nature attracts meaningful relationships.", sign, element)
	default:
		return fmt.Sprintf("Today brings new opportunities for %s. Trust your %s energy.", sign, element)
	}
}

func (s *AstrologyService) buildPersonalizedHoroscopePrompt(chart *database.BirthChart, prefs ports.PersonalizationPreferences) string {
	focusAreas := "general life guidance"
	if len(prefs.FocusAreas) > 0 {
		focusAreas = strings.Join(prefs.FocusAreas, ", ")
	}
	tone := prefs.Tone
	if tone == "" {
		tone = "supportive and grounded"
	}
	return fmt.Sprintf(`You are an expert astrology AI. Write a deeply tailored horoscope.

Birth Chart: Sun=%s | Moon=%s | Rising=%s
Planets: %s

Goals: %s | Focus: %s | Tone: %s

Provide practical, positive guidance tailored to this individual.`,
		chart.SunSign, chart.MoonSign, chart.RisingSign, chart.Planets,
		prefs.Goals, focusAreas, tone)
}

func (s *AstrologyService) generatePersonalizedHoroscopeFallback(chart *database.BirthChart, prefs ports.PersonalizationPreferences) string {
	focusAreas := "general life guidance"
	if len(prefs.FocusAreas) > 0 {
		focusAreas = strings.Join(prefs.FocusAreas, ", ")
	}
	tone := prefs.Tone
	if tone == "" {
		tone = "supportive and balanced"
	}
	return fmt.Sprintf("Personalized horoscope for %s (%s tone). Focus: %s. Goals: %s. Lean into your natural strengths and take small actions that support your long-term goals.",
		chart.SunSign, tone, focusAreas, prefs.Goals)
}

func (s *AstrologyService) generateFallbackRemedies(chart *database.BirthChart) string {
	element := s.getSignElement(chart.SunSign)
	return fmt.Sprintf(`General remedies for %s (%s element):

Health: Daily meditation, balanced diet, regular exercise.
Career: Build stability, network intentionally, set realistic goals.
Relationships: Communicate openly, practice active listening.
Spiritual: Connect with nature, practice gratitude daily.
Gemstones: Wear colors resonating with your %s energy.
Daily Practice: Morning affirmations, evening reflection.`,
		chart.SunSign, element, element)
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

// parseBirthTime parses "HH:MM" and returns an error on invalid input.
func parseBirthTime(birthTime string) (time.Time, error) {
	t, err := time.Parse(constants.TimeFormat, birthTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid birth_time %q: expected HH:MM format", birthTime)
	}
	return t, nil
}
