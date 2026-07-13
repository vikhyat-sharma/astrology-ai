package services

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"github.com/vikhyat-sharma/astrology-ai/internal/interfaces"
	"gorm.io/gorm"
)

// DashaService handles Vimshottari Dasha (planetary period) calculations.
type DashaService struct {
	astrologyRepo interfaces.AstrologyRepositoryInterface
	db            dashaDB
}

// dashaDB is the minimal DB interface DashaService needs for persistence.
// Keeping it narrow makes the service testable without a full GORM instance.
type dashaDB interface {
	Create(value interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
}

// NewDashaService creates a new DashaService.
// db may be nil — in that case Save/Get operations are no-ops (useful in tests
// that only exercise the calculation logic).
func NewDashaService(astrologyRepo interfaces.AstrologyRepositoryInterface) *DashaService {
	return &DashaService{astrologyRepo: astrologyRepo}
}

// NewDashaServiceWithDB creates a DashaService with a live DB handle for persistence.
func NewDashaServiceWithDB(astrologyRepo interfaces.AstrologyRepositoryInterface, db dashaDB) *DashaService {
	return &DashaService{astrologyRepo: astrologyRepo, db: db}
}

// DashaPeriod represents one mahadasha + antardasha combination.
type DashaPeriod struct {
	Mahadasha       string    `json:"mahadasha"`
	MahadashaStart  time.Time `json:"mahadasha_start"`
	MahadashaEnd    time.Time `json:"mahadasha_end"`
	Antardasha      string    `json:"antardasha"`
	AntardashaStart time.Time `json:"antardasha_start"`
	AntardashaEnd   time.Time `json:"antardasha_end"`
	PratyantarDasha string    `json:"pratyantar_dasha,omitempty"`
}

// Vimshottari constants — total cycle is 120 years.
var (
	dashaOrder = []string{"Sun", "Moon", "Mars", "Rahu", "Jupiter", "Saturn", "Mercury", "Ketu", "Venus"}
	dashaYears = map[string]float64{
		"Sun": 6, "Moon": 10, "Mars": 7, "Rahu": 18, "Jupiter": 16,
		"Saturn": 19, "Mercury": 17, "Ketu": 7, "Venus": 20,
	}
	// nakshatraLord maps each of the 27 nakshatras (0-indexed) to its dasha lord.
	// This is the canonical Vimshottari mapping — the dasha sequence repeats every
	// 9 nakshatras in the same order as dashaOrder.
	nakshatraLord = [27]string{
		"Ketu", "Venus", "Sun", "Moon", "Mars", "Rahu", "Jupiter", "Saturn", "Mercury", // 0–8
		"Ketu", "Venus", "Sun", "Moon", "Mars", "Rahu", "Jupiter", "Saturn", "Mercury", // 9–17
		"Ketu", "Venus", "Sun", "Moon", "Mars", "Rahu", "Jupiter", "Saturn", "Mercury", // 18–26
	}
)

// CalculateVimshottariDasha calculates Vimshottari Dasha periods for 30 years from birthDate.
// The Moon's nakshatra (stored on the chart) determines the starting dasha lord and balance.
func (s *DashaService) CalculateVimshottariDasha(chartID uuid.UUID, birthDate time.Time) ([]DashaPeriod, error) {
	chart, err := s.astrologyRepo.GetBirthChart(chartID)
	if err != nil {
		return nil, fmt.Errorf("failed to get birth chart: %w", err)
	}

	startingDasha, balanceYears := s.dashaStartAndBalance(chart)

	startIndex := 0
	for i, p := range dashaOrder {
		if p == startingDasha {
			startIndex = i
			break
		}
	}

	var periods []DashaPeriod
	currentDate := birthDate
	endDate := birthDate.AddDate(30, 0, 0)

	for currentDate.Before(endDate) {
		for i := 0; i < len(dashaOrder) && currentDate.Before(endDate); i++ {
			idx := (startIndex + i) % len(dashaOrder)
			mahadasha := dashaOrder[idx]

			years := dashaYears[mahadasha]
			if i == 0 {
				years = balanceYears
			}

			mahaStart := currentDate
			mahaEnd := addYears(currentDate, years)

			for _, ap := range s.antardashas(mahadasha, mahaStart, years) {
				end := ap.AntardashaEnd
				if end.After(endDate) {
					end = endDate
				}
				periods = append(periods, DashaPeriod{
					Mahadasha:       mahadasha,
					MahadashaStart:  mahaStart,
					MahadashaEnd:    mahaEnd,
					Antardasha:      ap.Antardasha,
					AntardashaStart: ap.AntardashaStart,
					AntardashaEnd:   end,
				})
			}
			currentDate = mahaEnd
		}
	}

	return periods, nil
}

// SaveDashaPeriods persists calculated dasha periods to the database.
// If no DB handle is configured the call is a no-op (test-safe).
func (s *DashaService) SaveDashaPeriods(chartID uuid.UUID, periods []DashaPeriod) error {
	if s.db == nil {
		return nil
	}
	for _, p := range periods {
		row := &database.Dasha{
			ChartID:         chartID,
			Type:            "vimshottari",
			Mahadasha:       p.Mahadasha,
			MahadashaStart:  p.MahadashaStart,
			MahadashaEnd:    p.MahadashaEnd,
			Antardasha:      p.Antardasha,
			AntardashaStart: p.AntardashaStart,
			AntardashaEnd:   p.AntardashaEnd,
			PratyantarDasha: p.PratyantarDasha,
		}
		if res := s.db.Create(row); res.Error != nil {
			return fmt.Errorf("failed to save dasha period: %w", res.Error)
		}
	}
	return nil
}

// GetCurrentDasha returns the active dasha period for a chart at the current time.
// It queries persisted periods; if none exist it returns an error rather than
// returning fabricated data.
func (s *DashaService) GetCurrentDasha(chartID uuid.UUID) (*DashaPeriod, error) {
	if s.db == nil {
		return nil, errors.New("dasha periods not available: no database configured")
	}

	now := time.Now().UTC()
	var row database.Dasha
	res := s.db.Where(
		"chart_id = ? AND antardasha_start <= ? AND antardasha_end >= ?",
		chartID, now, now,
	).(*gorm.DB).First(&row)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no active dasha period found for chart %s — run CalculateVimshottariDasha first", chartID)
		}
		return nil, fmt.Errorf("failed to query current dasha: %w", res.Error)
	}

	return &DashaPeriod{
		Mahadasha:       row.Mahadasha,
		MahadashaStart:  row.MahadashaStart,
		MahadashaEnd:    row.MahadashaEnd,
		Antardasha:      row.Antardasha,
		AntardashaStart: row.AntardashaStart,
		AntardashaEnd:   row.AntardashaEnd,
		PratyantarDasha: row.PratyantarDasha,
	}, nil
}

// dashaStartAndBalance returns the starting dasha lord and the fractional years
// remaining in that dasha at birth, derived from the Moon's nakshatra.
//
// The Vimshottari balance is proportional to how far the Moon has travelled
// through its birth nakshatra:
//
//	balance = dashaYears[lord] × (1 − moonDegreeInNakshatra / nakshatraSpan)
func (s *DashaService) dashaStartAndBalance(chart *database.BirthChart) (string, float64) {
	nakshatraSpan := 360.0 / 27.0 // ≈ 13.333°

	// Derive nakshatra index from the stored nakshatra name.
	nakshatraIndex := nakshatraIndexFromName(chart.Nakshatra)

	lord := nakshatraLord[nakshatraIndex]

	// We don't store the Moon's exact longitude on the chart, so we approximate
	// the degree within the nakshatra from NakshatraPad (each pada = 1/4 span).
	// This gives ±1 pada accuracy (~3.3°). A future improvement is to store
	// the Moon's exact longitude and compute this precisely.
	padaFraction := float64(chart.NakshatraPad-1) / 4.0 // 0.0, 0.25, 0.50, 0.75
	degreeInNakshatra := padaFraction * nakshatraSpan
	balance := dashaYears[lord] * (1.0 - degreeInNakshatra/nakshatraSpan)

	// Clamp to a sensible range.
	if balance <= 0 {
		balance = 0.1
	}
	if balance > dashaYears[lord] {
		balance = dashaYears[lord]
	}

	return lord, balance
}

// antardashas calculates the 9 antardasha sub-periods within a mahadasha.
// Each antardasha duration = (mahadashaYears × antardashaYears) / 120.
func (s *DashaService) antardashas(mahadasha string, start time.Time, mahaYears float64) []DashaPeriod {
	startIdx := 0
	for i, p := range dashaOrder {
		if p == mahadasha {
			startIdx = i
			break
		}
	}

	var periods []DashaPeriod
	current := start
	for i := 0; i < len(dashaOrder); i++ {
		idx := (startIdx + i) % len(dashaOrder)
		antardasha := dashaOrder[idx]
		days := (mahaYears * dashaYears[antardasha] / 120.0) * 365.25
		end := current.AddDate(0, 0, int(math.Round(days)))
		periods = append(periods, DashaPeriod{
			Antardasha:      antardasha,
			AntardashaStart: current,
			AntardashaEnd:   end,
		})
		current = end
	}
	return periods
}

// addYears adds a fractional number of years to a time using day-level precision.
func addYears(t time.Time, years float64) time.Time {
	return t.AddDate(0, 0, int(math.Round(years*365.25)))
}

// nakshatraIndexFromName returns the 0-based index of a nakshatra by name.
// Returns 0 (Ashwini / Ketu) as a safe default for unknown names.
func nakshatraIndexFromName(name string) int {
	names := [27]string{
		"Ashwini", "Bharani", "Krittika", "Rohini", "Mrigashira", "Ardra",
		"Punarvasu", "Pushya", "Ashlesha", "Magha", "Purva Phalguni", "Uttara Phalguni",
		"Hasta", "Chitra", "Swati", "Vishakha", "Anuradha", "Jyeshtha",
		"Mula", "Purva Ashadha", "Uttara Ashadha", "Shravana", "Dhanishta", "Shatabhisha",
		"Purva Bhadrapada", "Uttara Bhadrapada", "Revati",
	}
	for i, n := range names {
		if n == name {
			return i
		}
	}
	return 0
}
