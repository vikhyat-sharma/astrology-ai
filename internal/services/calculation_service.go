package services

import (
	"fmt"
	"math"
	"time"
)

// CalculationService handles astronomical calculations for astrology.
type CalculationService struct{}

// NewCalculationService creates a new CalculationService.
func NewCalculationService() *CalculationService {
	return &CalculationService{}
}

// PlanetPosition represents a planet's ecliptic position.
type PlanetPosition struct {
	Name         string  `json:"name"`
	Longitude    float64 `json:"longitude"`     // Tropical longitude 0–360°
	Degree       float64 `json:"degree"`        // Degree within sign 0–30°
	Sign         string  `json:"sign"`
	SignNumber   int     `json:"sign_number"`   // 1=Aries … 12=Pisces
	Retrograde   bool    `json:"retrograde"`
	Speed        float64 `json:"speed"`         // °/day
	House        int     `json:"house"`         // 1–12
	Exalted      bool    `json:"exalted"`
	Debilitated  bool    `json:"debilitated"`
	OwnSign      bool    `json:"own_sign"`
	FriendlySign bool    `json:"friendly_sign"`
}

// HousePosition represents a house cusp.
type HousePosition struct {
	Number     int     `json:"number"`
	CuspDegree float64 `json:"cusp_degree"`
	Sign       string  `json:"sign"`
	SignNumber int     `json:"sign_number"`
}

// Aspect represents an angular relationship between two planets.
type Aspect struct {
	Planet1    string  `json:"planet1"`
	Planet2    string  `json:"planet2"`
	AspectType string  `json:"aspect_type"`
	Orb        float64 `json:"orb"`
	Exact      bool    `json:"exact"`
}

// ChartData contains all calculated astrological data for a birth chart.
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

// CalculateBirthChart computes a complete birth chart for the given date, time, and location.
func (s *CalculationService) CalculateBirthChart(birthDate, birthTime time.Time, latitude, longitude float64) (*ChartData, error) {
	dt := time.Date(
		birthDate.Year(), birthDate.Month(), birthDate.Day(),
		birthTime.Hour(), birthTime.Minute(), birthTime.Second(),
		0, time.UTC,
	)

	jd := julianDay(dt)

	planets, err := s.calculatePlanetPositions(jd)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate planet positions: %w", err)
	}

	ascendant, err := s.calculateAscendant(jd, latitude, longitude)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate ascendant: %w", err)
	}

	midheaven := s.calculateMidheaven(jd, longitude)
	houses := s.calculateHouses(ascendant)
	planets = s.assignPlanetsToHouses(planets, houses)
	aspects := s.calculateAspects(planets)

	sunSign := zodiacSign(planets[0].Longitude)
	moonSign := zodiacSign(planets[1].Longitude)
	risingSign := zodiacSign(ascendant)

	nakshatra, nakshatraPad := calculateNakshatra(planets[1].Longitude)
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

// julianDay converts a UTC time to Julian Day Number.
// Algorithm: Meeus, "Astronomical Algorithms", 2nd ed., ch. 7.
func julianDay(t time.Time) float64 {
	y := float64(t.Year())
	m := float64(t.Month())
	d := float64(t.Day()) +
		float64(t.Hour())/24.0 +
		float64(t.Minute())/1440.0 +
		float64(t.Second())/86400.0

	if m <= 2 {
		y--
		m += 12
	}
	a := math.Floor(y / 100)
	b := 2 - a + math.Floor(a/4)
	return math.Floor(365.25*(y+4716)) + math.Floor(30.6001*(m+1)) + d + b - 1524.5
}

// calculateSunLongitude returns the Sun's apparent ecliptic longitude (degrees, 0–360).
// Accuracy: ~1 arcminute. Algorithm: low-precision VSOP87 (Meeus ch. 25).
func calculateSunLongitude(jd float64) float64 {
	// Julian centuries from J2000.0
	T := (jd - 2451545.0) / 36525.0

	// Geometric mean longitude of the Sun (degrees)
	L0 := 280.46646 + 36000.76983*T + 0.0003032*T*T
	L0 = normalizeDeg(L0)

	// Mean anomaly of the Sun (degrees)
	M := 357.52911 + 35999.05029*T - 0.0001537*T*T
	M = normalizeDeg(M)
	Mrad := deg2rad(M)

	// Equation of centre
	C := (1.914602-0.004817*T-0.000014*T*T)*math.Sin(Mrad) +
		(0.019993-0.000101*T)*math.Sin(2*Mrad) +
		0.000289*math.Sin(3*Mrad)

	// Sun's true longitude
	sunLon := L0 + C

	// Apparent longitude (correct for aberration and nutation — simplified)
	omega := 125.04 - 1934.136*T
	apparent := sunLon - 0.00569 - 0.00478*math.Sin(deg2rad(omega))

	return normalizeDeg(apparent)
}

// calculateMoonLongitude returns the Moon's ecliptic longitude (degrees, 0–360).
// Accuracy: ~1°. Algorithm: Meeus ch. 47 (truncated series).
func calculateMoonLongitude(jd float64) float64 {
	T := (jd - 2451545.0) / 36525.0

	// Moon's mean longitude
	L1 := 218.3164477 + 481267.88123421*T - 0.0015786*T*T + T*T*T/538841.0
	L1 = normalizeDeg(L1)

	// Moon's mean anomaly
	M1 := 134.9633964 + 477198.8675055*T + 0.0087414*T*T + T*T*T/69699.0
	M1 = normalizeDeg(M1)

	// Sun's mean anomaly
	M := 357.5291092 + 35999.0502909*T - 0.0001536*T*T
	M = normalizeDeg(M)

	// Moon's argument of latitude
	F := 93.2720950 + 483202.0175233*T - 0.0036539*T*T
	F = normalizeDeg(F)

	// Elongation of the Moon
	D := 297.8501921 + 445267.1114034*T - 0.0018819*T*T
	D = normalizeDeg(D)

	// Principal periodic terms (degrees)
	lon := L1 +
		6.288774*math.Sin(deg2rad(M1)) +
		1.274027*math.Sin(deg2rad(2*D-M1)) +
		0.658314*math.Sin(deg2rad(2*D)) +
		0.213618*math.Sin(deg2rad(2*M1)) -
		0.185116*math.Sin(deg2rad(M)) -
		0.114332*math.Sin(deg2rad(2*F)) +
		0.058793*math.Sin(deg2rad(2*D-2*M1)) +
		0.057066*math.Sin(deg2rad(2*D-M-M1)) +
		0.053322*math.Sin(deg2rad(2*D+M1)) +
		0.045758*math.Sin(deg2rad(2*D-M))

	return normalizeDeg(lon)
}

// calculatePlanetPositions returns positions for Sun, Moon, and outer planets.
// Sun and Moon use proper series; outer planets use mean motion approximations
// accurate to ~1–2° for dates within ±50 years of J2000.
func (s *CalculationService) calculatePlanetPositions(jd float64) ([]PlanetPosition, error) {
	T := (jd - 2451545.0) / 36525.0

	// Mean longitudes and daily motions (J2000 elements, degrees).
	// Source: Meeus, "Astronomical Algorithms", Table 31.a
	type planetElem struct {
		name string
		L0   float64 // mean longitude at J2000
		dL   float64 // degrees per Julian century
	}
	outerPlanets := []planetElem{
		{"Mercury", 252.250906, 149472.6746358},
		{"Venus", 181.979801, 58517.8156760},
		{"Mars", 355.433275, 19140.2993313},
		{"Jupiter", 34.351519, 3034.9056606},
		{"Saturn", 50.077444, 1222.1138488},
		{"Uranus", 314.055005, 428.4669983},
		{"Neptune", 304.348665, 218.4862002},
		{"Pluto", 238.92881, 145.2078},
	}

	sunLon := calculateSunLongitude(jd)
	moonLon := calculateMoonLongitude(jd)

	positions := []PlanetPosition{
		s.makePlanet("Sun", sunLon, 1.0, false),
		s.makePlanet("Moon", moonLon, 13.176, false),
	}

	for _, p := range outerPlanets {
		lon := normalizeDeg(p.L0 + p.dL*T/100)
		// Daily motion in degrees
		speed := p.dL / 36525.0
		retrograde := speed < 0
		positions = append(positions, s.makePlanet(p.name, lon, math.Abs(speed), retrograde))
	}

	// Rahu (mean ascending node of Moon): regresses ~19.3°/year
	rahu := normalizeDeg(125.0445479 - 1934.1362608*T)
	ketu := normalizeDeg(rahu + 180)
	positions = append(positions, s.makePlanet("Rahu", rahu, 0.053, true))
	positions = append(positions, s.makePlanet("Ketu", ketu, 0.053, true))

	return positions, nil
}

func (s *CalculationService) makePlanet(name string, lon, speed float64, retrograde bool) PlanetPosition {
	sign, signNum := zodiacSignAndNumber(lon)
	return PlanetPosition{
		Name:         name,
		Longitude:    lon,
		Degree:       math.Mod(lon, 30),
		Sign:         sign,
		SignNumber:   signNum,
		Retrograde:   retrograde,
		Speed:        speed,
		Exalted:      isExalted(name, signNum),
		Debilitated:  isDebilitated(name, signNum),
		OwnSign:      isOwnSign(name, signNum),
		FriendlySign: false, // TODO: implement full Shadbala friendship table
	}
}

// calculateAscendant computes the tropical Ascendant using LMST and obliquity.
// Algorithm: Meeus ch. 13 & 15. Accurate to ~1 arcminute for latitudes ±60°.
func (s *CalculationService) calculateAscendant(jd, latitude, longitude float64) (float64, error) {
	if latitude < -89.9 || latitude > 89.9 {
		return 0, fmt.Errorf("ascendant is undefined at polar latitudes (got %.4f°)", latitude)
	}

	T := (jd - 2451545.0) / 36525.0

	// Mean obliquity of the ecliptic (Laskar 1986, degrees)
	eps0 := 23.0 + 26.0/60.0 + 21.448/3600.0 -
		(4680.93/3600.0)*T -
		(1.55/3600.0)*T*T +
		(1999.25/3600.0)*T*T*T
	// Nutation in obliquity (simplified)
	omega := normalizeDeg(125.04452 - 1934.136261*T)
	eps := eps0 + (0.00256 * math.Cos(deg2rad(omega)))
	epsRad := deg2rad(eps)

	// Greenwich Mean Sidereal Time (degrees)
	gmst := 280.46061837 +
		360.98564736629*(jd-2451545.0) +
		0.000387933*T*T -
		T*T*T/38710000.0
	gmst = normalizeDeg(gmst)

	// Local Mean Sidereal Time
	lmst := normalizeDeg(gmst + longitude)
	lmstRad := deg2rad(lmst)
	latRad := deg2rad(latitude)

	// Ascendant formula (Meeus eq. 14.3)
	y := -math.Cos(lmstRad)
	x := math.Sin(epsRad)*math.Tan(latRad) + math.Cos(epsRad)*math.Sin(lmstRad)
	asc := rad2deg(math.Atan2(y, x))
	asc = normalizeDeg(asc)

	return asc, nil
}

// calculateMidheaven computes the Midheaven (MC) from LMST and obliquity.
func (s *CalculationService) calculateMidheaven(jd, longitude float64) float64 {
	T := (jd - 2451545.0) / 36525.0
	eps := 23.439291111 - 0.013004167*T
	epsRad := deg2rad(eps)

	gmst := normalizeDeg(280.46061837 + 360.98564736629*(jd-2451545.0) + 0.000387933*T*T)
	lmst := normalizeDeg(gmst + longitude)

	mc := rad2deg(math.Atan2(math.Sin(deg2rad(lmst)), math.Cos(deg2rad(lmst))*math.Cos(epsRad)))
	return normalizeDeg(mc)
}

// calculateHouses returns 12 equal house cusps starting from the Ascendant.
// Equal house system is used as a safe default; Placidus requires iterative solving.
func (s *CalculationService) calculateHouses(ascendant float64) []HousePosition {
	houses := make([]HousePosition, 12)
	for i := 0; i < 12; i++ {
		cusp := normalizeDeg(ascendant + float64(i)*30)
		sign, signNum := zodiacSignAndNumber(cusp)
		houses[i] = HousePosition{
			Number:     i + 1,
			CuspDegree: cusp,
			Sign:       sign,
			SignNumber: signNum,
		}
	}
	return houses
}

// assignPlanetsToHouses assigns each planet to its house number.
func (s *CalculationService) assignPlanetsToHouses(planets []PlanetPosition, houses []HousePosition) []PlanetPosition {
	for i := range planets {
		planets[i].House = houseForLongitude(planets[i].Longitude, houses)
	}
	return planets
}

func houseForLongitude(lon float64, houses []HousePosition) int {
	for i := 0; i < 12; i++ {
		next := (i + 1) % 12
		c1, c2 := houses[i].CuspDegree, houses[next].CuspDegree
		if c1 < c2 {
			if lon >= c1 && lon < c2 {
				return i + 1
			}
		} else {
			if lon >= c1 || lon < c2 {
				return i + 1
			}
		}
	}
	return 1
}

// calculateAspects finds major aspects between all planet pairs.
func (s *CalculationService) calculateAspects(planets []PlanetPosition) []Aspect {
	type aspectDef struct {
		name string
		deg  float64
		orb  float64
	}
	defs := []aspectDef{
		{"conjunction", 0, 10},
		{"sextile", 60, 6},
		{"square", 90, 8},
		{"trine", 120, 8},
		{"opposition", 180, 10},
	}

	var aspects []Aspect
	for i := 0; i < len(planets); i++ {
		for j := i + 1; j < len(planets); j++ {
			sep := math.Abs(planets[i].Longitude - planets[j].Longitude)
			if sep > 180 {
				sep = 360 - sep
			}
			for _, def := range defs {
				orb := math.Abs(sep - def.deg)
				if orb <= def.orb {
					aspects = append(aspects, Aspect{
						Planet1:    planets[i].Name,
						Planet2:    planets[j].Name,
						AspectType: def.name,
						Orb:        orb,
						Exact:      orb <= 1.0,
					})
					break
				}
			}
		}
	}
	return aspects
}

// calculateNakshatra returns the nakshatra name and pada for a given Moon longitude.
func calculateNakshatra(moonLon float64) (string, int) {
	nakshatras := [27]string{
		"Ashwini", "Bharani", "Krittika", "Rohini", "Mrigashira", "Ardra",
		"Punarvasu", "Pushya", "Ashlesha", "Magha", "Purva Phalguni", "Uttara Phalguni",
		"Hasta", "Chitra", "Swati", "Vishakha", "Anuradha", "Jyeshtha",
		"Mula", "Purva Ashadha", "Uttara Ashadha", "Shravana", "Dhanishta", "Shatabhisha",
		"Purva Bhadrapada", "Uttara Bhadrapada", "Revati",
	}
	// Each nakshatra spans 360/27 ≈ 13.333°; each pada is 1/4 of that.
	nakshatraSpan := 360.0 / 27.0
	idx := int(moonLon/nakshatraSpan) % 27
	pada := int(math.Mod(moonLon/(nakshatraSpan/4), 4)) + 1
	return nakshatras[idx], pada
}

// detectYogas detects classical Vedic yogas in the chart.
func (s *CalculationService) detectYogas(planets []PlanetPosition, houses []HousePosition) []string {
	var yogas []string
	if s.hasGajakesariYoga(planets) {
		yogas = append(yogas, "Gajakesari Yoga")
	}
	if s.hasRajaYoga(planets) {
		yogas = append(yogas, "Raja Yoga")
	}
	return yogas
}

// hasGajakesariYoga: Moon and Jupiter in mutual kendra (1,4,7,10) from each other.
func (s *CalculationService) hasGajakesariYoga(planets []PlanetPosition) bool {
	var moonHouse, jupiterHouse int
	for _, p := range planets {
		switch p.Name {
		case "Moon":
			moonHouse = p.House
		case "Jupiter":
			jupiterHouse = p.House
		}
	}
	if moonHouse == 0 || jupiterHouse == 0 {
		return false
	}
	diff := abs(moonHouse - jupiterHouse)
	if diff > 6 {
		diff = 12 - diff
	}
	return diff == 0 || diff == 3 || diff == 6 || diff == 9
}

// hasRajaYoga: Jupiter and Venus both in kendra houses (1,4,7,10).
func (s *CalculationService) hasRajaYoga(planets []PlanetPosition) bool {
	kendras := map[int]bool{1: true, 4: true, 7: true, 10: true}
	var jupiterKendra, venusKendra bool
	for _, p := range planets {
		switch p.Name {
		case "Jupiter":
			jupiterKendra = kendras[p.House]
		case "Venus":
			venusKendra = kendras[p.House]
		}
	}
	return jupiterKendra && venusKendra
}

// calculateJulianDay is exported for use by astrology_service.go (transit calculations).
func (s *CalculationService) calculateJulianDay(t time.Time) float64 {
	return julianDay(t)
}

func (s *CalculationService) calculatePlanetPositionsPublic(jd float64) ([]PlanetPosition, error) {
	return s.calculatePlanetPositions(jd)
}

// Dignity tables (Vedic / traditional Western)

func isExalted(planet string, signNum int) bool {
	exaltations := map[string]int{
		"Sun": 1, "Moon": 2, "Mercury": 6, "Venus": 12,
		"Mars": 10, "Jupiter": 4, "Saturn": 7,
	}
	s, ok := exaltations[planet]
	return ok && s == signNum
}

func isDebilitated(planet string, signNum int) bool {
	debilitations := map[string]int{
		"Sun": 7, "Moon": 8, "Mercury": 12, "Venus": 6,
		"Mars": 4, "Jupiter": 10, "Saturn": 1,
	}
	s, ok := debilitations[planet]
	return ok && s == signNum
}

func isOwnSign(planet string, signNum int) bool {
	ownSigns := map[string][]int{
		"Sun":     {5},
		"Moon":    {4},
		"Mercury": {3, 6},
		"Venus":   {2, 7},
		"Mars":    {1, 8},
		"Jupiter": {9, 12},
		"Saturn":  {10, 11},
	}
	for _, s := range ownSigns[planet] {
		if s == signNum {
			return true
		}
	}
	return false
}

// Zodiac helpers

var zodiacNames = [12]string{
	"Aries", "Taurus", "Gemini", "Cancer", "Leo", "Virgo",
	"Libra", "Scorpio", "Sagittarius", "Capricorn", "Aquarius", "Pisces",
}

func zodiacSign(lon float64) string {
	s, _ := zodiacSignAndNumber(lon)
	return s
}

func zodiacSignAndNumber(lon float64) (string, int) {
	idx := int(lon/30) % 12
	return zodiacNames[idx], idx + 1
}

// Math helpers

func normalizeDeg(d float64) float64 {
	d = math.Mod(d, 360)
	if d < 0 {
		d += 360
	}
	return d
}

func deg2rad(d float64) float64 { return d * math.Pi / 180 }
func rad2deg(r float64) float64 { return r * 180 / math.Pi }

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
