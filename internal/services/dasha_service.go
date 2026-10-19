package services

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/ports"
)

// DashaService handles Vimshottari Dasha (planetary period) calculations.
type DashaService struct {
	astrologyRepo ports.AstrologyRepository
}

// NewDashaService creates a new DashaService.
func NewDashaService(astrologyRepo ports.AstrologyRepository) *DashaService {
	return &DashaService{astrologyRepo: astrologyRepo}
}

// DashaPeriod represents one mahadasha + antardasha sub-period.
type DashaPeriod struct {
	Mahadasha       string    `json:"mahadasha"`
	MahadashaStart  time.Time `json:"mahadasha_start"`
	MahadashaEnd    time.Time `json:"mahadasha_end"`
	Antardasha      string    `json:"antardasha"`
	AntardashaStart time.Time `json:"antardasha_start"`
	AntardashaEnd   time.Time `json:"antardasha_end"`
	PratyantarDasha string    `json:"pratyantar_dasha,omitempty"`
}

// Vimshottari cycle: 9 planets, 120 years total.
var (
	dashaOrder = []string{"Sun", "Moon", "Mars", "Rahu", "Jupiter", "Saturn", "Mercury", "Ketu", "Venus"}
	dashaYears = map[string]float64{
		"Sun": 6, "Moon": 10, "Mars": 7, "Rahu": 18, "Jupiter": 16,
		"Saturn": 19, "Mercury": 17, "Ketu": 7, "Venus": 20,
	}
)

// CalculateVimshottariDasha returns dasha periods for the next 30 years from birthDate.
func (s *DashaService) CalculateVimshottariDasha(chartID uuid.UUID, birthDate time.Time) ([]DashaPeriod, error) {
	chart, err := s.astrologyRepo.GetBirthChart(chartID)
	if err != nil {
		return nil, fmt.Errorf("failed to get birth chart: %w", err)
	}

	startingDasha := getStartingDashaFromMoon(chart.MoonSign)
	startIndex := indexInCycle(startingDasha)
	balanceYears := dashaYears[startingDasha] * 0.5 // simplified: assume 50% remaining at birth

	var periods []DashaPeriod
	current := birthDate
	end := birthDate.AddDate(30, 0, 0)

	for current.Before(end) {
		for i := 0; i < len(dashaOrder) && current.Before(end); i++ {
			planet := dashaOrder[(startIndex+i)%len(dashaOrder)]
			years := dashaYears[planet]
			if i == 0 {
				years = balanceYears
			}

			mahaStart := current
			mahaEnd := current.AddDate(0, 0, int(years*365.25))

			for _, sub := range calculateAntardashas(planet, mahaStart, years) {
				if sub.AntardashaEnd.After(end) {
					sub.AntardashaEnd = end
				}
				periods = append(periods, DashaPeriod{
					Mahadasha:       planet,
					MahadashaStart:  mahaStart,
					MahadashaEnd:    mahaEnd,
					Antardasha:      sub.Antardasha,
					AntardashaStart: sub.AntardashaStart,
					AntardashaEnd:   sub.AntardashaEnd,
				})
			}
			current = mahaEnd
		}
	}
	return periods, nil
}

// GetCurrentDasha returns a mock current dasha period.
// TODO: replace with a real DB query once dasha persistence is implemented.
func (s *DashaService) GetCurrentDasha(chartID uuid.UUID) (*DashaPeriod, error) {
	now := time.Now()
	return &DashaPeriod{
		Mahadasha:       "Jupiter",
		MahadashaStart:  now.AddDate(0, -6, 0),
		MahadashaEnd:    now.AddDate(0, 10, 0),
		Antardasha:      "Venus",
		AntardashaStart: now.AddDate(0, -2, 0),
		AntardashaEnd:   now.AddDate(0, 4, 0),
	}, nil
}

// getStartingDashaFromMoon maps a Moon sign to its ruling dasha planet.
func getStartingDashaFromMoon(moonSign string) string {
	signToDasha := map[string]string{
		"Aries": "Mars", "Taurus": "Venus", "Gemini": "Mercury",
		"Cancer": "Moon", "Leo": "Sun", "Virgo": "Mercury",
		"Libra": "Venus", "Scorpio": "Mars", "Sagittarius": "Jupiter",
		"Capricorn": "Saturn", "Aquarius": "Saturn", "Pisces": "Jupiter",
	}
	if d, ok := signToDasha[moonSign]; ok {
		return d
	}
	return "Moon"
}

// indexInCycle returns the position of a planet in the dasha cycle.
func indexInCycle(planet string) int {
	for i, p := range dashaOrder {
		if p == planet {
			return i
		}
	}
	return 0
}

// calculateAntardashas returns the nine sub-periods within a mahadasha.
func calculateAntardashas(mahadasha string, start time.Time, mahaYears float64) []DashaPeriod {
	totalDays := mahaYears * 365.25
	startIdx := indexInCycle(mahadasha)
	var periods []DashaPeriod
	current := start

	for i := 0; i < len(dashaOrder); i++ {
		planet := dashaOrder[(startIdx+i)%len(dashaOrder)]
		days := int(math.Round(totalDays * dashaYears[planet] / 120.0))
		end := current.AddDate(0, 0, days)
		periods = append(periods, DashaPeriod{
			Antardasha:      planet,
			AntardashaStart: current,
			AntardashaEnd:   end,
		})
		current = end
	}
	return periods
}
