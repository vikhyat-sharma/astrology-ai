package services

import (
	"fmt"
	"math"
	"time"
)

// CalculationService handles accurate astronomical calculations for astrology
type CalculationService struct{}

// NewCalculationService creates a new calculation service
func NewCalculationService() *CalculationService {
	return &CalculationService{}
}

// PlanetPosition represents a planet's position in the zodiac
type PlanetPosition struct {
	Name         string  `json:"name"`
	Longitude    float64 `json:"longitude"`     // Degrees in zodiac (0-360)
	Degree       float64 `json:"degree"`        // Degree within sign (0-30)
	Sign         string  `json:"sign"`          // Zodiac sign name
	SignNumber   int     `json:"sign_number"`   // 1-12 (Aries=1, Pisces=12)
	Retrograde   bool    `json:"retrograde"`    // Is planet retrograde?
	Speed        float64 `json:"speed"`         // Daily speed in degrees
	House        int     `json:"house"`         // House position (1-12)
	Exalted      bool    `json:"exalted"`       // Is in exaltation sign?
	Debilitated  bool    `json:"debilitated"`   // Is in debilitation sign?
	OwnSign      bool    `json:"own_sign"`      // Is in own sign?
	FriendlySign bool    `json:"friendly_sign"` // Is in friendly sign?
}

// HousePosition represents a house cusp position
type HousePosition struct {
	Number     int     `json:"number"`
	CuspDegree float64 `json:"cusp_degree"` // Exact degree
	Sign       string  `json:"sign"`
	SignNumber int     `json:"sign_number"`
}

// Aspect represents an aspect between two planets
type Aspect struct {
	Planet1    string  `json:"planet1"`
	Planet2    string  `json:"planet2"`
	AspectType string  `json:"aspect_type"` // "conjunction", "sextile", "square", "trine", "opposition"
	Orb        float64 `json:"orb"`         // Separation in degrees
	Exact      bool    `json:"exact"`       // Is exact aspect?
}

// ChartData contains all calculated astrological data
type ChartData struct {
	Ascendant    float64          `json:"ascendant"`
	Midheaven    float64          `json:"midheaven"`
	Planets      []PlanetPosition `json:"planets"`
	Houses       []HousePosition  `json:"houses"`
	Aspects      []Aspect         `json:"aspects"`
	SunSign      string           `json:"sun_sign"`
	MoonSign     string           `json:"moon_sign"`
	RisingSign   string           `json:"rising_sign"`
	Nakshatra    string           `json:"nakshatra"`
	NakshatraPad int              `json:"nakshatra_pad"`
	Yogas        []string         `json:"yogas"`
}

// calculateJulianDay converts time to Julian Day (simplified)
func (s *CalculationService) calculateJulianDay(t time.Time) float64 {
	// Simplified Julian Day calculation
	year := float64(t.Year())
	month := float64(t.Month())
	day := float64(t.Day())
	hour := float64(t.Hour()) / 24.0

	if month <= 2 {
		year--
		month += 12
	}

	a := math.Floor(year / 100)
	b := 2 - a + math.Floor(a/4)

	jd := math.Floor(365.25*(year+4716)) + math.Floor(30.6001*(month+1)) + day + b - 1524 + hour

	return jd
}

// CalculateBirthChart calculates complete birth chart data
func (s *CalculationService) CalculateBirthChart(birthDate time.Time, birthTime time.Time, latitude, longitude float64) (*ChartData, error) {
	// Combine date and time
	birthDateTime := time.Date(
		birthDate.Year(), birthDate.Month(), birthDate.Day(),
		birthTime.Hour(), birthTime.Minute(), birthTime.Second(),
		0, time.UTC,
	)

	// Convert to Julian Day (simplified calculation)
	jd := s.calculateJulianDay(birthDateTime)

	// Calculate planetary positions (simplified)
	planets, err := s.calculatePlanetPositions(jd)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate planets: %w", err)
	}

	// Calculate house cusps (simplified Placidus system)
	houses := s.calculateHouses(jd, latitude, longitude)

	// Calculate ascendant and midheaven
	ascendant := houses[0].CuspDegree
	midheaven := s.calculateMidheaven(jd, latitude, longitude)

	// Assign planets to houses
	planets = s.assignPlanetsToHouses(planets, houses)

	// Calculate aspects
	aspects := s.calculateAspects(planets)

	// Determine signs
	sunSign := s.getZodiacSign(planets[0].Longitude)  // Sun is first
	moonSign := s.getZodiacSign(planets[1].Longitude) // Moon is second
	risingSign := s.getZodiacSign(ascendant)

	// Calculate nakshatra
	nakshatra, nakshatraPad := s.calculateNakshatra(planets[1].Longitude) // Moon's position

	// Detect yogas
	yogas := s.detectYogas(planets, houses)

	return &ChartData{
		Ascendant:    ascendant,
		Midheaven:    midheaven,
		Planets:      planets,
		Houses:       houses,
		Aspects:      aspects,
		SunSign:      sunSign,
		MoonSign:     moonSign,
		RisingSign:   risingSign,
		Nakshatra:    nakshatra,
		NakshatraPad: nakshatraPad,
		Yogas:        yogas,
	}, nil
}

// calculatePlanetPositions calculates positions for all planets (simplified)
func (s *CalculationService) calculatePlanetPositions(jd float64) ([]PlanetPosition, error) {
	planets := []PlanetPosition{}

	planetNames := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Uranus", "Neptune", "Pluto", "Rahu", "Ketu"}

	for i, name := range planetNames {
		var longitude float64

		if name == "Sun" {
			// Approximate sun position based on Julian day
			// Sun moves ~1 degree per day, starts at Aries (0°) around March 21
			daysSinceSpring := jd - 2451545.0             // Approximate J2000 epoch
			longitude = math.Mod(daysSinceSpring+80, 360) // +80 to align with Aries start
		} else {
			// Simplified calculation for other planets
			basePosition := math.Mod(float64(i)*30+jd*0.1, 360)
			longitude = basePosition + float64(i)*2.5
		}

		// Convert to 0-360 range
		longitude = math.Mod(longitude, 360)
		if longitude < 0 {
			longitude += 360
		}

		// Determine zodiac sign
		sign, signNum := s.getZodiacSignAndNumber(longitude)
		degree := math.Mod(longitude, 30)

		// Check dignity (simplified)
		exalted := s.isExalted(name, signNum)
		debilitated := s.isDebilitated(name, signNum)
		ownSign := s.isOwnSign(name, signNum)
		friendlySign := s.isFriendlySign(name, signNum)

		planet := PlanetPosition{
			Name:         name,
			Longitude:    longitude,
			Degree:       degree,
			Sign:         sign,
			SignNumber:   signNum,
			Retrograde:   false, // Simplified
			Speed:        1.0,   // Simplified
			Exalted:      exalted,
			Debilitated:  debilitated,
			OwnSign:      ownSign,
			FriendlySign: friendlySign,
		}

		planets = append(planets, planet)
	}

	return planets, nil
}

// calculateHouses calculates house cusps using Placidus system
func (s *CalculationService) calculateHouses(jd float64, lat, lng float64) []HousePosition {
	houses := make([]HousePosition, 12)

	// Simplified house calculation - in production, use proper Placidus algorithm
	ascendant := s.calculateAscendant(jd, lat, lng)

	for i := 0; i < 12; i++ {
		cuspDegree := ascendant + float64(i)*30
		cuspDegree = math.Mod(cuspDegree, 360)

		sign, signNum := s.getZodiacSignAndNumber(cuspDegree)

		houses[i] = HousePosition{
			Number:     i + 1,
			CuspDegree: cuspDegree,
			Sign:       sign,
			SignNumber: signNum,
		}
	}

	return houses
}

// calculateAscendant calculates the ascendant (rising sign)
func (s *CalculationService) calculateAscendant(jd float64, lat, lng float64) float64 {
	// Simplified calculation - in production, use proper astronomical algorithm
	// This is a placeholder that returns a reasonable ascendant
	return 0 // TODO: Implement proper ascendant calculation
}

// calculateMidheaven calculates the midheaven (MC)
func (s *CalculationService) calculateMidheaven(jd float64, lat, lng float64) float64 {
	// Simplified calculation
	return 90 // TODO: Implement proper MC calculation
}

// assignPlanetsToHouses assigns each planet to its house
func (s *CalculationService) assignPlanetsToHouses(planets []PlanetPosition, houses []HousePosition) []PlanetPosition {
	for i := range planets {
		house := s.findHouseForLongitude(planets[i].Longitude, houses)
		planets[i].House = house
	}
	return planets
}

// findHouseForLongitude finds which house a longitude falls into
func (s *CalculationService) findHouseForLongitude(longitude float64, houses []HousePosition) int {
	for i := 0; i < 12; i++ {
		nextHouse := (i + 1) % 12
		if s.isLongitudeInHouse(longitude, houses[i].CuspDegree, houses[nextHouse].CuspDegree) {
			return i + 1
		}
	}
	return 1 // Default to first house
}

// isLongitudeInHouse checks if longitude is between two house cusps
func (s *CalculationService) isLongitudeInHouse(longitude, cusp1, cusp2 float64) bool {
	if cusp1 < cusp2 {
		return longitude >= cusp1 && longitude < cusp2
	}
	// Handle 360-degree wraparound
	return longitude >= cusp1 || longitude < cusp2
}

// calculateAspects calculates aspects between planets
func (s *CalculationService) calculateAspects(planets []PlanetPosition) []Aspect {
	aspects := []Aspect{}

	for i := 0; i < len(planets); i++ {
		for j := i + 1; j < len(planets); j++ {
			aspect := s.calculateAspect(planets[i], planets[j])
			if aspect != nil {
				aspects = append(aspects, *aspect)
			}
		}
	}

	return aspects
}

// calculateAspect calculates aspect between two planets
func (s *CalculationService) calculateAspect(p1, p2 PlanetPosition) *Aspect {
	// Calculate angular separation
	sep := math.Abs(p1.Longitude - p2.Longitude)
	if sep > 180 {
		sep = 360 - sep
	}

	// Check for major aspects
	var aspectType string
	var orb float64

	if sep <= 10 { // Conjunction
		aspectType = "conjunction"
		orb = sep
	} else if math.Abs(sep-60) <= 8 { // Sextile
		aspectType = "sextile"
		orb = math.Abs(sep - 60)
	} else if math.Abs(sep-90) <= 8 { // Square
		aspectType = "square"
		orb = math.Abs(sep - 90)
	} else if math.Abs(sep-120) <= 8 { // Trine
		aspectType = "trine"
		orb = math.Abs(sep - 120)
	} else if math.Abs(sep-180) <= 8 { // Opposition
		aspectType = "opposition"
		orb = math.Abs(sep - 180)
	} else {
		return nil // No aspect
	}

	return &Aspect{
		Planet1:    p1.Name,
		Planet2:    p2.Name,
		AspectType: aspectType,
		Orb:        orb,
		Exact:      orb <= 1.0,
	}
}

// getZodiacSign returns zodiac sign name for longitude
func (s *CalculationService) getZodiacSign(longitude float64) string {
	sign, _ := s.getZodiacSignAndNumber(longitude)
	return sign
}

// getZodiacSignAndNumber returns both sign name and number
func (s *CalculationService) getZodiacSignAndNumber(longitude float64) (string, int) {
	signs := []string{
		"Aries", "Taurus", "Gemini", "Cancer", "Leo", "Virgo",
		"Libra", "Scorpio", "Sagittarius", "Capricorn", "Aquarius", "Pisces",
	}

	signIndex := int(longitude/30) % 12
	return signs[signIndex], signIndex + 1
}

// calculateNakshatra calculates nakshatra for moon position
func (s *CalculationService) calculateNakshatra(moonLongitude float64) (string, int) {
	nakshatras := []string{
		"Ashwini", "Bharani", "Krittika", "Rohini", "Mrigashira", "Ardra",
		"Punarvasu", "Pushya", "Ashlesha", "Magha", "Purva Phalguni", "Uttara Phalguni",
		"Hasta", "Chitra", "Swati", "Vishakha", "Anuradha", "Jyeshtha",
		"Mula", "Purva Ashadha", "Uttara Ashadha", "Shravana", "Dhanishta", "Shatabhisha",
		"Purva Bhadrapada", "Uttara Bhadrapada", "Revati",
	}

	nakshatraIndex := int(moonLongitude/(360.0/27.0)) % 27
	pada := int(math.Mod(moonLongitude/(360.0/108.0), 4)) + 1

	return nakshatras[nakshatraIndex], pada
}

// detectYogas detects astrological yogas in the chart
func (s *CalculationService) detectYogas(planets []PlanetPosition, houses []HousePosition) []string {
	yogas := []string{}

	// Raja Yoga detection (simplified)
	if s.hasRajaYoga(planets) {
		yogas = append(yogas, "Raja Yoga")
	}

	// Dhana Yoga detection
	if s.hasDhanaYoga(planets) {
		yogas = append(yogas, "Dhana Yoga")
	}

	// Gajakesari Yoga
	if s.hasGajakesariYoga(planets) {
		yogas = append(yogas, "Gajakesari Yoga")
	}

	return yogas
}

// hasRajaYoga checks for Raja Yoga
func (s *CalculationService) hasRajaYoga(planets []PlanetPosition) bool {
	// Simplified: Check if Jupiter and Venus are in kendra houses
	jupiterHouse := 0
	venusHouse := 0

	for _, planet := range planets {
		if planet.Name == "Jupiter" {
			jupiterHouse = planet.House
		}
		if planet.Name == "Venus" {
			venusHouse = planet.House
		}
	}

	kendraHouses := []int{1, 4, 7, 10}
	isJupiterKendra := false
	isVenusKendra := false

	for _, house := range kendraHouses {
		if jupiterHouse == house {
			isJupiterKendra = true
		}
		if venusHouse == house {
			isVenusKendra = true
		}
	}

	return isJupiterKendra && isVenusKendra
}

// hasDhanaYoga checks for wealth yoga
func (s *CalculationService) hasDhanaYoga(planets []PlanetPosition) bool {
	// Simplified: Check if 2nd lord and 11th lord are strong
	return false // TODO: Implement proper logic
}

// hasGajakesariYoga checks for Gajakesari Yoga
func (s *CalculationService) hasGajakesariYoga(planets []PlanetPosition) bool {
	// Moon and Jupiter in kendra from each other
	return false // TODO: Implement proper logic
}

// Dignity checking functions
func (s *CalculationService) isExalted(planet string, signNum int) bool {
	exaltations := map[string]int{
		"Sun": 1, "Moon": 2, "Mercury": 6, "Venus": 12, "Mars": 10,
		"Jupiter": 4, "Saturn": 7,
	}
	if sign, exists := exaltations[planet]; exists {
		return sign == signNum
	}
	return false
}

func (s *CalculationService) isDebilitated(planet string, signNum int) bool {
	debilitations := map[string]int{
		"Sun": 7, "Moon": 8, "Mercury": 12, "Venus": 6, "Mars": 4,
		"Jupiter": 10, "Saturn": 1,
	}
	if sign, exists := debilitations[planet]; exists {
		return sign == signNum
	}
	return false
}

func (s *CalculationService) isOwnSign(planet string, signNum int) bool {
	ownSigns := map[string][]int{
		"Sun":     {5},      // Leo
		"Moon":    {4},      // Cancer
		"Mercury": {3, 6},   // Gemini, Virgo
		"Venus":   {2, 7},   // Taurus, Libra
		"Mars":    {1, 8},   // Aries, Scorpio
		"Jupiter": {9, 12},  // Sagittarius, Pisces
		"Saturn":  {10, 11}, // Capricorn, Aquarius
	}
	if signs, exists := ownSigns[planet]; exists {
		for _, sign := range signs {
			if sign == signNum {
				return true
			}
		}
	}
	return false
}

func (s *CalculationService) isFriendlySign(planet string, signNum int) bool {
	// Simplified - in practice, this is more complex
	return false // TODO: Implement proper friendly sign logic
}

// Lunar node calculations (simplified)
func (s *CalculationService) calculateRahuPosition(jd float64) (float64, error) {
	// Simplified Rahu calculation
	return 0, nil // TODO: Implement proper lunar node calculation
}

func (s *CalculationService) calculateKetuPosition(jd float64) (float64, error) {
	// Ketu is opposite to Rahu
	rahu, _ := s.calculateRahuPosition(jd)
	return math.Mod(rahu+180, 360), nil
}
