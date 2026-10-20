package services

import (
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"github.com/vikhyat-sharma/astrology-ai/internal/ports"
)

// CompatibilityService handles Guna Milan compatibility analysis between two birth charts.
type CompatibilityService struct {
	astrologyRepo ports.AstrologyRepository
}

// NewCompatibilityService creates a new CompatibilityService.
func NewCompatibilityService(astrologyRepo ports.AstrologyRepository) *CompatibilityService {
	return &CompatibilityService{astrologyRepo: astrologyRepo}
}

// CompatibilityResult holds the full Guna Milan breakdown.
// Max score per kuta: Varna=1, Vashya=2, Tara=3, Yoni=4, GrahaMaitri=5, Gana=6, Bhakut=7, Nadi=8 → total 36.
type CompatibilityResult struct {
	OverallScore     int                  `json:"overall_score"`
	VarnaScore       int                  `json:"varna_score"`
	VashyaScore      int                  `json:"vashya_score"`
	TaraScore        int                  `json:"tara_score"`
	YoniScore        int                  `json:"yoni_score"`
	GrahaMaitriScore int                  `json:"graha_maitri_score"`
	GanaScore        int                  `json:"gana_score"`
	BhakutScore      int                  `json:"bhakut_score"`
	NadiScore        int                  `json:"nadi_score"`
	Analysis         string               `json:"analysis"`
	Chart1           *database.BirthChart `json:"chart1"`
	Chart2           *database.BirthChart `json:"chart2"`
}

// CheckCompatibility fetches both charts and runs all eight Guna Milan kutas.
func (s *CompatibilityService) CheckCompatibility(chartID1, chartID2 uuid.UUID) (*CompatibilityResult, error) {
	chart1, err := s.astrologyRepo.GetBirthChart(chartID1)
	if err != nil {
		return nil, fmt.Errorf("failed to get chart 1: %w", err)
	}
	chart2, err := s.astrologyRepo.GetBirthChart(chartID2)
	if err != nil {
		return nil, fmt.Errorf("failed to get chart 2: %w", err)
	}

	varna := s.calculateVarnaCompatibility(chart1.RisingSign, chart2.RisingSign)
	vashya := s.calculateVashyaCompatibility(chart1.SunSign, chart2.SunSign)
	tara := s.calculateTaraCompatibility(chart1.MoonSign, chart2.MoonSign)
	yoni := s.calculateYoniCompatibility(chart1.MoonSign, chart2.MoonSign)
	grahaMaitri := s.calculateGrahaMaitriCompatibility(chart1.SunSign, chart2.SunSign)
	gana := s.calculateGanaCompatibility(chart1.MoonSign, chart2.MoonSign)
	bhakut := s.calculateBhakutCompatibility(chart1.Ascendant, chart2.Ascendant)
	nadi := s.calculateNadiCompatibility(chart1.MoonSign, chart2.MoonSign)

	overall := varna + vashya + tara + yoni + grahaMaitri + gana + bhakut + nadi

	return &CompatibilityResult{
		OverallScore:     overall,
		VarnaScore:       varna,
		VashyaScore:      vashya,
		TaraScore:        tara,
		YoniScore:        yoni,
		GrahaMaitriScore: grahaMaitri,
		GanaScore:        gana,
		BhakutScore:      bhakut,
		NadiScore:        nadi,
		Analysis:         s.generateCompatibilityAnalysis(overall, varna, vashya, tara, yoni, grahaMaitri, gana, bhakut, nadi),
		Chart1:           chart1,
		Chart2:           chart2,
	}, nil
}

func (s *CompatibilityService) calculateVarnaCompatibility(rising1, rising2 string) int {
	if s.getVarnaFromSign(rising1) == s.getVarnaFromSign(rising2) {
		return 1
	}
	return 0
}

func (s *CompatibilityService) calculateVashyaCompatibility(sun1, sun2 string) int {
	v1, v2 := s.getVashyaFromSign(sun1), s.getVashyaFromSign(sun2)
	if (v1 == "human" && v2 == "human") || (v1 != "human" && v2 != "human") {
		return 2
	}
	return 1
}

func (s *CompatibilityService) calculateTaraCompatibility(moon1, moon2 string) int {
	for _, c1 := range s.getConstellations(moon1) {
		for _, c2 := range s.getConstellations(moon2) {
			if s.areTaraCompatible(c1, c2) {
				return 3
			}
		}
	}
	return 1
}

func (s *CompatibilityService) calculateYoniCompatibility(moon1, moon2 string) int {
	y1, y2 := s.getYoniFromSign(moon1), s.getYoniFromSign(moon2)
	if y1 == y2 {
		return 4
	}
	// All yoni types are only friendly with themselves in this simplified model.
	return 1
}

func (s *CompatibilityService) calculateGrahaMaitriCompatibility(sun1, sun2 string) int {
	if s.arePlanetsFriendly(s.getRulingPlanet(sun1), s.getRulingPlanet(sun2)) {
		return 5
	}
	return 2
}

func (s *CompatibilityService) calculateGanaCompatibility(moon1, moon2 string) int {
	g1, g2 := s.getGanaFromSign(moon1), s.getGanaFromSign(moon2)
	if g1 == g2 {
		return 6
	}
	if (g1 == "deva" && g2 == "manushya") || (g1 == "manushya" && g2 == "deva") {
		return 5
	}
	return 1
}

func (s *CompatibilityService) calculateBhakutCompatibility(asc1, asc2 float64) int {
	diff := math.Abs(asc1 - asc2)
	if diff > 180 {
		diff = 360 - diff
	}
	if diff >= 150 && diff <= 210 {
		return 7
	}
	return 4
}

func (s *CompatibilityService) calculateNadiCompatibility(moon1, moon2 string) int {
	if s.getNadiFromSign(moon1) != s.getNadiFromSign(moon2) {
		return 8 // Different nadi = maximum compatibility
	}
	return 0 // Same nadi = incompatibility
}

// --- Lookup tables ---

func (s *CompatibilityService) getVarnaFromSign(sign string) string {
	return map[string]string{
		"Aries": "kshatriya", "Leo": "kshatriya", "Sagittarius": "kshatriya",
		"Taurus": "vaishya", "Virgo": "vaishya", "Capricorn": "vaishya",
		"Gemini": "vaishya", "Libra": "vaishya", "Aquarius": "vaishya",
		"Cancer": "shudra", "Scorpio": "shudra", "Pisces": "shudra",
	}[sign]
}

func (s *CompatibilityService) getVashyaFromSign(sign string) string {
	return map[string]string{
		"Aries": "quadruped", "Taurus": "quadruped", "Cancer": "quadruped",
		"Capricorn": "quadruped", "Pisces": "water",
		"Gemini": "human", "Virgo": "human", "Libra": "human",
		"Sagittarius": "human", "Aquarius": "human",
		"Leo": "wild", "Scorpio": "insect",
	}[sign]
}

func (s *CompatibilityService) getConstellations(sign string) []string {
	return map[string][]string{
		"Aries": {"Aries", "Taurus"}, "Taurus": {"Taurus", "Gemini"},
		"Gemini": {"Gemini", "Cancer"}, "Cancer": {"Cancer", "Leo"},
		"Leo": {"Leo", "Virgo"}, "Virgo": {"Virgo", "Libra"},
		"Libra": {"Libra", "Scorpio"}, "Scorpio": {"Scorpio", "Sagittarius"},
		"Sagittarius": {"Sagittarius", "Capricorn"}, "Capricorn": {"Capricorn", "Aquarius"},
		"Aquarius": {"Aquarius", "Pisces"}, "Pisces": {"Pisces", "Aries"},
	}[sign]
}

func (s *CompatibilityService) areTaraCompatible(c1, c2 string) bool {
	incompatible := map[string][]string{
		"Aries":  {"Cancer", "Capricorn"},
		"Taurus": {"Leo", "Aquarius"},
		"Gemini": {"Virgo", "Pisces"},
	}
	for _, inc := range incompatible[c1] {
		if inc == c2 {
			return false
		}
	}
	return true
}

func (s *CompatibilityService) getYoniFromSign(sign string) string {
	return map[string]string{
		"Aries": "sheep", "Taurus": "cow", "Gemini": "horse", "Cancer": "elephant",
		"Leo": "lion", "Virgo": "buffalo", "Libra": "rat", "Scorpio": "snake",
		"Sagittarius": "horse", "Capricorn": "buffalo", "Aquarius": "lion", "Pisces": "elephant",
	}[sign]
}

func (s *CompatibilityService) getRulingPlanet(sign string) string {
	return map[string]string{
		"Aries": "Mars", "Taurus": "Venus", "Gemini": "Mercury", "Cancer": "Moon",
		"Leo": "Sun", "Virgo": "Mercury", "Libra": "Venus", "Scorpio": "Mars",
		"Sagittarius": "Jupiter", "Capricorn": "Saturn", "Aquarius": "Saturn", "Pisces": "Jupiter",
	}[sign]
}

func (s *CompatibilityService) arePlanetsFriendly(p1, p2 string) bool {
	friends := map[string][]string{
		"Sun":     {"Moon", "Mars", "Jupiter"},
		"Moon":    {"Sun", "Mercury"},
		"Mars":    {"Sun", "Moon", "Jupiter"},
		"Mercury": {"Sun", "Venus"},
		"Jupiter": {"Sun", "Moon", "Mars"},
		"Venus":   {"Mercury", "Saturn"},
		"Saturn":  {"Mercury", "Venus"},
	}
	for _, f := range friends[p1] {
		if f == p2 {
			return true
		}
	}
	return false
}

func (s *CompatibilityService) getGanaFromSign(sign string) string {
	return map[string]string{
		"Aries": "rakshasa", "Taurus": "manushya", "Gemini": "deva", "Cancer": "rakshasa",
		"Leo": "deva", "Virgo": "manushya", "Libra": "manushya", "Scorpio": "rakshasa",
		"Sagittarius": "deva", "Capricorn": "rakshasa", "Aquarius": "manushya", "Pisces": "deva",
	}[sign]
}

func (s *CompatibilityService) getNadiFromSign(sign string) string {
	return map[string]string{
		"Aries": "wind", "Taurus": "earth", "Gemini": "wind", "Cancer": "water",
		"Leo": "fire", "Virgo": "earth", "Libra": "wind", "Scorpio": "water",
		"Sagittarius": "fire", "Capricorn": "earth", "Aquarius": "wind", "Pisces": "water",
	}[sign]
}

// --- Analysis ---

func (s *CompatibilityService) generateCompatibilityAnalysis(overall, varna, vashya, tara, yoni, grahaMaitri, gana, bhakut, nadi int) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Overall Compatibility Score: %d/36\n\n", overall))

	switch {
	case overall >= 28:
		b.WriteString("Excellent compatibility! This is a highly auspicious match.\n\n")
	case overall >= 21:
		b.WriteString("Good compatibility. Overall prospects are positive.\n\n")
	case overall >= 14:
		b.WriteString("Average compatibility. Some areas require understanding and compromise.\n\n")
	default:
		b.WriteString("Below average compatibility. Significant challenges may be present.\n\n")
	}

	b.WriteString("Detailed Analysis:\n")
	b.WriteString(fmt.Sprintf("- Varna (Spiritual):    %d/1 — %s\n", varna, scoreLabel(varna, 1)))
	b.WriteString(fmt.Sprintf("- Vashya (Dominance):   %d/2 — %s\n", vashya, scoreLabel(vashya, 2)))
	b.WriteString(fmt.Sprintf("- Tara (Destiny):       %d/3 — %s\n", tara, scoreLabel(tara, 3)))
	b.WriteString(fmt.Sprintf("- Yoni (Physical):      %d/4 — %s\n", yoni, scoreLabel(yoni, 4)))
	b.WriteString(fmt.Sprintf("- Graha Maitri (Mental):%d/5 — %s\n", grahaMaitri, scoreLabel(grahaMaitri, 5)))
	b.WriteString(fmt.Sprintf("- Gana (Temperament):   %d/6 — %s\n", gana, scoreLabel(gana, 6)))
	b.WriteString(fmt.Sprintf("- Bhakut (Happiness):   %d/7 — %s\n", bhakut, scoreLabel(bhakut, 7)))
	b.WriteString(fmt.Sprintf("- Nadi (Health):        %d/8 — %s\n", nadi, scoreLabel(nadi, 8)))
	return b.String()
}

func scoreLabel(score, max int) string {
	pct := float64(score) / float64(max) * 100
	switch {
	case pct >= 80:
		return "Excellent"
	case pct >= 60:
		return "Good"
	case pct >= 40:
		return "Average"
	default:
		return "Needs attention"
	}
}
