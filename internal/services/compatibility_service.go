package services

import (
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"github.com/vikhyat-sharma/astrology-ai/internal/interfaces"
)

// CompatibilityService handles compatibility analysis between birth charts
type CompatibilityService struct {
	astrologyRepo interfaces.AstrologyRepositoryInterface
}

// NewCompatibilityService creates a new compatibility service
func NewCompatibilityService(astrologyRepo interfaces.AstrologyRepositoryInterface) *CompatibilityService {
	return &CompatibilityService{astrologyRepo: astrologyRepo}
}

// CompatibilityResult represents detailed compatibility analysis
type CompatibilityResult struct {
	OverallScore     int                  `json:"overall_score"`      // 1-36
	VarnaScore       int                  `json:"varna_score"`        // 1-1
	VashyaScore      int                  `json:"vashya_score"`       // 1-2
	TaraScore        int                  `json:"tara_score"`         // 1-3
	YoniScore        int                  `json:"yoni_score"`         // 1-4
	GrahaMaitriScore int                  `json:"graha_maitri_score"` // 1-5
	GanaScore        int                  `json:"gana_score"`         // 1-6
	BhakutScore      int                  `json:"bhakut_score"`       // 1-7
	NadiScore        int                  `json:"nadi_score"`         // 1-8
	Analysis         string               `json:"analysis"`
	Chart1           *database.BirthChart `json:"chart1"`
	Chart2           *database.BirthChart `json:"chart2"`
}

// CheckCompatibility performs detailed compatibility analysis
func (s *CompatibilityService) CheckCompatibility(chartID1, chartID2 uuid.UUID) (*CompatibilityResult, error) {
	chart1, err := s.astrologyRepo.GetBirthChart(chartID1)
	if err != nil {
		return nil, fmt.Errorf("failed to get chart 1: %w", err)
	}

	chart2, err := s.astrologyRepo.GetBirthChart(chartID2)
	if err != nil {
		return nil, fmt.Errorf("failed to get chart 2: %w", err)
	}

	// Calculate Guna Milan scores
	varnaScore := s.calculateVarnaCompatibility(chart1.RisingSign, chart2.RisingSign)
	vashyaScore := s.calculateVashyaCompatibility(chart1.SunSign, chart2.SunSign)
	taraScore := s.calculateTaraCompatibility(chart1.MoonSign, chart2.MoonSign)
	yoniScore := s.calculateYoniCompatibility(chart1.MoonSign, chart2.MoonSign)
	grahaMaitriScore := s.calculateGrahaMaitriCompatibility(chart1.SunSign, chart2.SunSign)
	ganaScore := s.calculateGanaCompatibility(chart1.MoonSign, chart2.MoonSign)
	bhakutScore := s.calculateBhakutCompatibility(chart1.Ascendant, chart2.Ascendant)
	nadiScore := s.calculateNadiCompatibility(chart1.MoonSign, chart2.MoonSign)

	overallScore := varnaScore + vashyaScore + taraScore + yoniScore + grahaMaitriScore + ganaScore + bhakutScore + nadiScore

	analysis := s.generateCompatibilityAnalysis(overallScore, varnaScore, vashyaScore, taraScore, yoniScore, grahaMaitriScore, ganaScore, bhakutScore, nadiScore)

	result := &CompatibilityResult{
		OverallScore:     overallScore,
		VarnaScore:       varnaScore,
		VashyaScore:      vashyaScore,
		TaraScore:        taraScore,
		YoniScore:        yoniScore,
		GrahaMaitriScore: grahaMaitriScore,
		GanaScore:        ganaScore,
		BhakutScore:      bhakutScore,
		NadiScore:        nadiScore,
		Analysis:         analysis,
		Chart1:           chart1,
		Chart2:           chart2,
	}

	return result, nil
}

// calculateVarnaCompatibility calculates Varna (caste) compatibility
func (s *CompatibilityService) calculateVarnaCompatibility(rising1, rising2 string) int {
	// Simplified Varna calculation based on ascendant
	varna1 := s.getVarnaFromSign(rising1)
	varna2 := s.getVarnaFromSign(rising2)

	if varna1 == varna2 {
		return 1
	}
	return 0
}

// calculateVashyaCompatibility calculates Vashya (dominance) compatibility
func (s *CompatibilityService) calculateVashyaCompatibility(sun1, sun2 string) int {
	vashya1 := s.getVashyaFromSign(sun1)
	vashya2 := s.getVashyaFromSign(sun2)

	// Higher score if compatible vashya types
	if (vashya1 == "human" && vashya2 == "human") || (vashya1 != "human" && vashya2 != "human") {
		return 2
	}
	return 1
}

// calculateTaraCompatibility calculates Tara (star) compatibility
func (s *CompatibilityService) calculateTaraCompatibility(moon1, moon2 string) int {
	// Simplified tara calculation
	constellations1 := s.getConstellations(moon1)
	constellations2 := s.getConstellations(moon2)

	// Check if moons are in compatible constellations
	for _, c1 := range constellations1 {
		for _, c2 := range constellations2 {
			if s.areTaraCompatible(c1, c2) {
				return 3
			}
		}
	}
	return 1
}

// calculateYoniCompatibility calculates Yoni (sexual) compatibility
func (s *CompatibilityService) calculateYoniCompatibility(moon1, moon2 string) int {
	yoni1 := s.getYoniFromSign(moon1)
	yoni2 := s.getYoniFromSign(moon2)

	if yoni1 == yoni2 {
		return 4 // Same yoni - maximum compatibility
	}

	// Check friendly yonis
	friendlyYonis := map[string][]string{
		"elephant": {"elephant"},
		"horse":    {"horse"},
		"sheep":    {"sheep"},
		"snake":    {"snake"},
		"dog":      {"dog"},
		"cat":      {"cat"},
		"rat":      {"rat"},
		"cow":      {"cow"},
		"buffalo":  {"buffalo"},
		"lion":     {"lion"},
		"tiger":    {"tiger"},
	}

	if friends, exists := friendlyYonis[yoni1]; exists {
		for _, friend := range friends {
			if friend == yoni2 {
				return 3
			}
		}
	}

	return 1 // Neutral
}

// calculateGrahaMaitriCompatibility calculates planetary friendship compatibility
func (s *CompatibilityService) calculateGrahaMaitriCompatibility(sun1, sun2 string) int {
	planet1 := s.getRulingPlanet(sun1)
	planet2 := s.getRulingPlanet(sun2)

	if s.arePlanetsFriendly(planet1, planet2) {
		return 5
	}
	return 2
}

// calculateGanaCompatibility calculates Gana (temperament) compatibility
func (s *CompatibilityService) calculateGanaCompatibility(moon1, moon2 string) int {
	gana1 := s.getGanaFromSign(moon1)
	gana2 := s.getGanaFromSign(moon2)

	if gana1 == gana2 {
		return 6
	}

	// Deva and Manushya are compatible
	if (gana1 == "deva" && gana2 == "manushya") || (gana1 == "manushya" && gana2 == "deva") {
		return 5
	}

	return 1
}

// calculateBhakutCompatibility calculates Bhakut (happiness) compatibility
func (s *CompatibilityService) calculateBhakutCompatibility(asc1, asc2 float64) int {
	// Simplified bhakut based on ascendant positions
	diff := math.Abs(asc1 - asc2)
	if diff > 180 {
		diff = 360 - diff
	}

	// Better compatibility with certain angular differences
	if diff >= 150 && diff <= 210 {
		return 7
	}
	return 4
}

// calculateNadiCompatibility calculates Nadi (health) compatibility
func (s *CompatibilityService) calculateNadiCompatibility(moon1, moon2 string) int {
	nadi1 := s.getNadiFromSign(moon1)
	nadi2 := s.getNadiFromSign(moon2)

	if nadi1 != nadi2 {
		return 8 // Different nadi - maximum compatibility
	}
	return 0 // Same nadi - incompatibility
}

// Helper functions for compatibility calculations

func (s *CompatibilityService) getVarnaFromSign(sign string) string {
	varnas := map[string]string{
		"Aries": "kshatriya", "Leo": "kshatriya", "Sagittarius": "kshatriya",
		"Taurus": "vaishya", "Virgo": "vaishya", "Capricorn": "vaishya",
		"Gemini": "vaishya", "Libra": "vaishya", "Aquarius": "vaishya",
		"Cancer": "shudra", "Scorpio": "shudra", "Pisces": "shudra",
	}
	return varnas[sign]
}

func (s *CompatibilityService) getVashyaFromSign(sign string) string {
	vashyas := map[string]string{
		"Aries": "quadruped", "Taurus": "quadruped", "Cancer": "quadruped", "Virgo": "human", "Libra": "human",
		"Capricorn": "quadruped", "Pisces": "water", "Gemini": "human", "Leo": "wild", "Scorpio": "insect",
		"Sagittarius": "human", "Aquarius": "human",
	}
	return vashyas[sign]
}

func (s *CompatibilityService) getConstellations(sign string) []string {
	// Simplified - each sign has multiple constellations
	constellations := map[string][]string{
		"Aries":       {"Aries", "Taurus"},
		"Taurus":      {"Taurus", "Gemini"},
		"Gemini":      {"Gemini", "Cancer"},
		"Cancer":      {"Cancer", "Leo"},
		"Leo":         {"Leo", "Virgo"},
		"Virgo":       {"Virgo", "Libra"},
		"Libra":       {"Libra", "Scorpio"},
		"Scorpio":     {"Scorpio", "Sagittarius"},
		"Sagittarius": {"Sagittarius", "Capricorn"},
		"Capricorn":   {"Capricorn", "Aquarius"},
		"Aquarius":    {"Aquarius", "Pisces"},
		"Pisces":      {"Pisces", "Aries"},
	}
	return constellations[sign]
}

func (s *CompatibilityService) areTaraCompatible(const1, const2 string) bool {
	// Simplified tara compatibility
	incompatible := map[string][]string{
		"Aries":  {"Cancer", "Capricorn"},
		"Taurus": {"Leo", "Aquarius"},
		"Gemini": {"Virgo", "Pisces"},
	}
	if incompat, exists := incompatible[const1]; exists {
		for _, inc := range incompat {
			if inc == const2 {
				return false
			}
		}
	}
	return true
}

func (s *CompatibilityService) getYoniFromSign(sign string) string {
	yonis := map[string]string{
		"Aries": "sheep", "Taurus": "cow", "Gemini": "horse", "Cancer": "elephant",
		"Leo": "lion", "Virgo": "buffalo", "Libra": "rat", "Scorpio": "snake",
		"Sagittarius": "horse", "Capricorn": "buffalo", "Aquarius": "lion", "Pisces": "elephant",
	}
	return yonis[sign]
}

func (s *CompatibilityService) getRulingPlanet(sign string) string {
	rulers := map[string]string{
		"Aries": "Mars", "Taurus": "Venus", "Gemini": "Mercury", "Cancer": "Moon",
		"Leo": "Sun", "Virgo": "Mercury", "Libra": "Venus", "Scorpio": "Mars",
		"Sagittarius": "Jupiter", "Capricorn": "Saturn", "Aquarius": "Saturn", "Pisces": "Jupiter",
	}
	return rulers[sign]
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

	if friendList, exists := friends[p1]; exists {
		for _, friend := range friendList {
			if friend == p2 {
				return true
			}
		}
	}
	return false
}

func (s *CompatibilityService) getGanaFromSign(sign string) string {
	ganas := map[string]string{
		"Aries": "rakshasa", "Taurus": "manushya", "Gemini": "deva", "Cancer": "rakshasa",
		"Leo": "deva", "Virgo": "manushya", "Libra": "manushya", "Scorpio": "rakshasa",
		"Sagittarius": "deva", "Capricorn": "rakshasa", "Aquarius": "manushya", "Pisces": "deva",
	}
	return ganas[sign]
}

func (s *CompatibilityService) getNadiFromSign(sign string) string {
	nadis := map[string]string{
		"Aries": "wind", "Taurus": "earth", "Gemini": "wind", "Cancer": "water",
		"Leo": "fire", "Virgo": "earth", "Libra": "wind", "Scorpio": "water",
		"Sagittarius": "fire", "Capricorn": "earth", "Aquarius": "wind", "Pisces": "water",
	}
	return nadis[sign]
}

func (s *CompatibilityService) generateCompatibilityAnalysis(overall, varna, vashya, tara, yoni, grahaMaitri, gana, bhakut, nadi int) string {
	var analysis strings.Builder

	analysis.WriteString(fmt.Sprintf("Overall Compatibility Score: %d/36\n\n", overall))

	if overall >= 28 {
		analysis.WriteString("Excellent compatibility! This is a highly auspicious match with strong potential for a harmonious marriage.\n\n")
	} else if overall >= 21 {
		analysis.WriteString("Good compatibility. The couple may need to work on some areas but overall prospects are positive.\n\n")
	} else if overall >= 14 {
		analysis.WriteString("Average compatibility. Some challenges may arise that require understanding and compromise.\n\n")
	} else {
		analysis.WriteString("Below average compatibility. Significant challenges may be present that require careful consideration.\n\n")
	}

	analysis.WriteString("Detailed Analysis:\n")
	analysis.WriteString(fmt.Sprintf("- Varna (Spiritual): %d/1 - %s\n", varna, s.getScoreDescription(varna, 1)))
	analysis.WriteString(fmt.Sprintf("- Vashya (Dominance): %d/2 - %s\n", vashya, s.getScoreDescription(vashya, 2)))
	analysis.WriteString(fmt.Sprintf("- Tara (Destiny): %d/3 - %s\n", tara, s.getScoreDescription(tara, 3)))
	analysis.WriteString(fmt.Sprintf("- Yoni (Physical): %d/4 - %s\n", yoni, s.getScoreDescription(yoni, 4)))
	analysis.WriteString(fmt.Sprintf("- Graha Maitri (Mental): %d/5 - %s\n", grahaMaitri, s.getScoreDescription(grahaMaitri, 5)))
	analysis.WriteString(fmt.Sprintf("- Gana (Temperament): %d/6 - %s\n", gana, s.getScoreDescription(gana, 6)))
	analysis.WriteString(fmt.Sprintf("- Bhakut (Happiness): %d/7 - %s\n", bhakut, s.getScoreDescription(bhakut, 7)))
	analysis.WriteString(fmt.Sprintf("- Nadi (Health): %d/8 - %s\n", nadi, s.getScoreDescription(nadi, 8)))

	return analysis.String()
}

func (s *CompatibilityService) getScoreDescription(score, max int) string {
	percentage := float64(score) / float64(max) * 100
	switch {
	case percentage >= 80:
		return "Excellent"
	case percentage >= 60:
		return "Good"
	case percentage >= 40:
		return "Average"
	default:
		return "Needs attention"
	}
}
