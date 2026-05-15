package services

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"github.com/vikhyat-sharma/astrology-ai/internal/interfaces"
)

// DashaService handles Dasha (planetary period) calculations
type DashaService struct {
	astrologyRepo interfaces.AstrologyRepositoryInterface
}

// NewDashaService creates a new dasha service
func NewDashaService(astrologyRepo interfaces.AstrologyRepositoryInterface) *DashaService {
	return &DashaService{astrologyRepo: astrologyRepo}
}

// DashaPeriod represents a dasha period
type DashaPeriod struct {
	Mahadasha       string    `json:"mahadasha"`
	MahadashaStart  time.Time `json:"mahadasha_start"`
	MahadashaEnd    time.Time `json:"mahadasha_end"`
	Antardasha      string    `json:"antardasha"`
	AntardashaStart time.Time `json:"antardasha_start"`
	AntardashaEnd   time.Time `json:"antardasha_end"`
	PratyantarDasha string    `json:"pratyantar_dasha,omitempty"`
}

// CalculateVimshottariDasha calculates Vimshottari Dasha periods
func (s *DashaService) CalculateVimshottariDasha(chartID uuid.UUID, birthDate time.Time) ([]DashaPeriod, error) {
	chart, err := s.astrologyRepo.GetBirthChart(chartID)
	if err != nil {
		return nil, fmt.Errorf("failed to get birth chart: %w", err)
	}

	// Determine starting dasha based on Moon's position
	startingDasha := s.getStartingDashaFromMoon(chart.MoonSign)

	// Vimshottari Dasha cycle: 120 years total
	dashaOrder := []string{"Sun", "Moon", "Mars", "Rahu", "Jupiter", "Saturn", "Mercury", "Ketu", "Venus"}
	dashaYears := map[string]float64{
		"Sun": 6, "Moon": 10, "Mars": 7, "Rahu": 18, "Jupiter": 16,
		"Saturn": 19, "Mercury": 17, "Ketu": 7, "Venus": 20,
	}

	// Find starting point in cycle
	startIndex := 0
	for i, planet := range dashaOrder {
		if planet == startingDasha {
			startIndex = i
			break
		}
	}

	// Calculate balance of current dasha at birth
	balanceYears := s.calculateDashaBalance(chart, startingDasha)

	var periods []DashaPeriod
	currentDate := birthDate

	// Calculate periods for next 30 years (can be extended)
	endDate := birthDate.AddDate(30, 0, 0)

	for currentDate.Before(endDate) {
		for i := 0; i < len(dashaOrder) && currentDate.Before(endDate); i++ {
			planetIndex := (startIndex + i) % len(dashaOrder)
			mahadasha := dashaOrder[planetIndex]

			var mahadashaYears float64
			if i == 0 {
				mahadashaYears = balanceYears
			} else {
				mahadashaYears = dashaYears[mahadasha]
			}

			mahadashaStart := currentDate
			mahadashaEnd := currentDate.AddDate(0, 0, int(mahadashaYears*365.25))

			// Calculate antardashas within this mahadasha
			antardashaPeriods := s.calculateAntardashas(mahadasha, mahadashaStart, mahadashaYears)

			for _, antardasha := range antardashaPeriods {
				if antardasha.AntardashaEnd.After(endDate) {
					antardasha.AntardashaEnd = endDate
				}

				period := DashaPeriod{
					Mahadasha:       mahadasha,
					MahadashaStart:  mahadashaStart,
					MahadashaEnd:    mahadashaEnd,
					Antardasha:      antardasha.Antardasha,
					AntardashaStart: antardasha.AntardashaStart,
					AntardashaEnd:   antardasha.AntardashaEnd,
				}
				periods = append(periods, period)
			}

			currentDate = mahadashaEnd
		}
	}

	return periods, nil
}

// getStartingDashaFromMoon determines starting dasha from Moon's position
func (s *DashaService) getStartingDashaFromMoon(moonSign string) string {
	// Simplified mapping - in practice, this uses nakshatra
	signToDasha := map[string]string{
		"Aries": "Mars", "Taurus": "Venus", "Gemini": "Mercury",
		"Cancer": "Moon", "Leo": "Sun", "Virgo": "Mercury",
		"Libra": "Venus", "Scorpio": "Mars", "Sagittarius": "Jupiter",
		"Capricorn": "Saturn", "Aquarius": "Saturn", "Pisces": "Jupiter",
	}

	if dasha, exists := signToDasha[moonSign]; exists {
		return dasha
	}
	return "Moon" // Default
}

// calculateDashaBalance calculates remaining years in current dasha at birth
func (s *DashaService) calculateDashaBalance(chart *database.BirthChart, startingDasha string) float64 {
	// Simplified calculation - in practice, uses exact birth time and planetary positions
	dashaYears := map[string]float64{
		"Sun": 6, "Moon": 10, "Mars": 7, "Rahu": 18, "Jupiter": 16,
		"Saturn": 19, "Mercury": 17, "Ketu": 7, "Venus": 20,
	}

	totalYears := dashaYears[startingDasha]
	// Assume 50% of dasha remaining (simplified)
	return totalYears * 0.5
}

// calculateAntardashas calculates sub-periods within a mahadasha
func (s *DashaService) calculateAntardashas(mahadasha string, startDate time.Time, mahadashaYears float64) []DashaPeriod {
	dashaOrder := []string{"Sun", "Moon", "Mars", "Rahu", "Jupiter", "Saturn", "Mercury", "Ketu", "Venus"}
	dashaYears := map[string]float64{
		"Sun": 6, "Moon": 10, "Mars": 7, "Rahu": 18, "Jupiter": 16,
		"Saturn": 19, "Mercury": 17, "Ketu": 7, "Venus": 20,
	}

	var periods []DashaPeriod
	currentDate := startDate

	// Find starting index for antardashas
	startIndex := 0
	for i, planet := range dashaOrder {
		if planet == mahadasha {
			startIndex = i
			break
		}
	}

	totalDays := mahadashaYears * 365.25

	for i := 0; i < len(dashaOrder); i++ {
		planetIndex := (startIndex + i) % len(dashaOrder)
		antardasha := dashaOrder[planetIndex]

		// Calculate proportional time for this antardasha
		antardashaPortion := dashaYears[antardasha] / 120.0 // 120 is total Vimshottari years
		antardashaDays := totalDays * antardashaPortion

		antardashaStart := currentDate
		antardashaEnd := currentDate.AddDate(0, 0, int(math.Round(antardashaDays)))

		period := DashaPeriod{
			Antardasha:      antardasha,
			AntardashaStart: antardashaStart,
			AntardashaEnd:   antardashaEnd,
		}
		periods = append(periods, period)

		currentDate = antardashaEnd
	}

	return periods
}

// SaveDashaPeriods saves calculated dasha periods to database
func (s *DashaService) SaveDashaPeriods(chartID uuid.UUID, periods []DashaPeriod) error {
	for _, period := range periods {
		dasha := &database.Dasha{
			ChartID:         chartID,
			Type:            "vimshottari",
			Mahadasha:       period.Mahadasha,
			MahadashaStart:  period.MahadashaStart,
			MahadashaEnd:    period.MahadashaEnd,
			Antardasha:      period.Antardasha,
			AntardashaStart: period.AntardashaStart,
			AntardashaEnd:   period.AntardashaEnd,
			PratyantarDasha: period.PratyantarDasha,
		}

		// In a real implementation, you'd save to database
		// For now, we'll skip database operations
		_ = dasha
	}

	return nil
}

// GetCurrentDasha gets the current dasha period for a chart
func (s *DashaService) GetCurrentDasha(chartID uuid.UUID) (*DashaPeriod, error) {
	now := time.Now()

	// In a real implementation, you'd query the database
	// For now, return a mock current dasha
	return &DashaPeriod{
		Mahadasha:       "Jupiter",
		MahadashaStart:  now.AddDate(0, -6, 0),
		MahadashaEnd:    now.AddDate(0, 10, 0),
		Antardasha:      "Venus",
		AntardashaStart: now.AddDate(0, -2, 0),
		AntardashaEnd:   now.AddDate(0, 4, 0),
	}, nil
}
