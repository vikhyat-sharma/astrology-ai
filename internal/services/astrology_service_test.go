package services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/constants"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"github.com/vikhyat-sharma/astrology-ai/internal/mocks"
	"github.com/vikhyat-sharma/astrology-ai/internal/ports"
)

var ctx = context.Background()

// TestSunSignCalculation validates the VSOP87-simplified sun longitude against
// known reference values. Tolerance is ±1 sign boundary (the simplified series
// is accurate to ~1 arcminute, well within a single sign).
func TestSunSignCalculation(t *testing.T) {
	tests := []struct {
		name string
		date string
		want string
	}{
		// Reference: https://www.astro.com — tropical sun signs for noon UTC.
		{"Aries equinox", "2023-03-21", "Aries"},
		{"Taurus mid", "2023-05-05", "Taurus"},
		{"Leo mid", "2023-08-05", "Leo"},
		{"Capricorn mid", "2023-01-10", "Capricorn"},
		{"Scorpio mid", "2023-11-05", "Scorpio"},
	}

	calc := NewCalculationService()
	noon := time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := time.Parse(constants.DateFormat, tt.date)
			if err != nil {
				t.Fatal(err)
			}
			chart, err := calc.CalculateBirthChart(d, noon, 0, 0)
			if err != nil {
				t.Fatalf("CalculateBirthChart error: %v", err)
			}
			if chart.SunSign != tt.want {
				t.Errorf("SunSign for %s = %q, want %q", tt.date, chart.SunSign, tt.want)
			}
		})
	}
}

func TestGetSignElement(t *testing.T) {
	svc := &AstrologyService{}
	cases := map[string]string{
		constants.Aries:  constants.ElementFiery,
		constants.Virgo:  constants.ElementEarthy,
		constants.Gemini: constants.ElementAiry,
		constants.Pisces: constants.ElementWatery,
	}
	for sign, want := range cases {
		if got := svc.getSignElement(sign); got != want {
			t.Errorf("getSignElement(%q) = %q, want %q", sign, got, want)
		}
	}
}

func TestParseBirthTimeValid(t *testing.T) {
	bt, err := parseBirthTime("10:30")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bt.Hour() != 10 || bt.Minute() != 30 {
		t.Errorf("expected 10:30, got %02d:%02d", bt.Hour(), bt.Minute())
	}
}

func TestParseBirthTimeInvalid(t *testing.T) {
	_, err := parseBirthTime("25:99")
	if err == nil {
		t.Fatal("expected error for invalid time, got nil")
	}
}

func TestParseBirthTimeEmpty(t *testing.T) {
	_, err := parseBirthTime("")
	if err == nil {
		t.Fatal("expected error for empty time, got nil")
	}
}

func TestGenerateDailyHoroscopeWithMock(t *testing.T) {
	mockHTTP := &mocks.MockHTTPClient{
		PostFunc: func(url, contentType string, body io.Reader) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"output": "AI generated horoscope content"}`))),
			}, nil
		},
	}
	svc := NewAstrologyServiceWithClient(nil, "http://localhost:11434", "llama2", mockHTTP)
	out := svc.generateHoroscope(constants.Leo, constants.HoroscopeTypeDaily)
	if out != "AI generated horoscope content" {
		t.Fatalf("expected AI content, got %q", out)
	}
}

func TestGenerateDailyHoroscopeFallback(t *testing.T) {
	mockHTTP := &mocks.MockHTTPClient{
		PostFunc: func(url, contentType string, body io.Reader) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}
	svc := NewAstrologyServiceWithClient(nil, "", "", mockHTTP)
	out := svc.generateHoroscope(constants.Leo, constants.HoroscopeTypeDaily)
	if out == "" {
		t.Fatal("expected non-empty fallback horoscope")
	}
}

func TestGetCurrentTransits(t *testing.T) {
	svc := NewAstrologyServiceWithClient(nil, "", "", nil)
	transits, err := svc.GetCurrentTransits(ctx, 12.9716, 77.5946)
	if err != nil {
		t.Fatalf("GetCurrentTransits failed: %v", err)
	}
	if len(transits) == 0 {
		t.Fatal("expected at least one transit")
	}
	first := transits[0]
	if first["planet"] == "" {
		t.Fatal("expected planet name in transit")
	}
	if first["sign"] == "" {
		t.Fatal("expected sign in transit")
	}
	if _, ok := first["house"]; !ok {
		t.Fatal("expected house in transit")
	}
}

func TestCalculateCompatibility(t *testing.T) {
	mockRepo := &mocks.MockAstrologyRepository{
		GetBirthChartFunc: func(id uuid.UUID) (*database.BirthChart, error) {
			return &database.BirthChart{
				ID:         id,
				SunSign:    constants.Aries,
				RisingSign: constants.Aries,
				MoonSign:   constants.Aries,
				Ascendant:  0,
			}, nil
		},
	}
	svc := NewAstrologyServiceWithClient(mockRepo, "", "", nil)
	result, err := svc.CheckCompatibility(ctx, uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("CheckCompatibility failed: %v", err)
	}
	score := result["overall_score"].(int)
	if score < 0 || score > 36 {
		t.Fatalf("overall_score %d out of range 0–36", score)
	}
	if result["analysis"].(string) == "" {
		t.Fatal("expected non-empty analysis")
	}
}

func TestCreateBirthChart(t *testing.T) {
	mockRepo := &mocks.MockAstrologyRepository{
		CreateBirthChartFunc: func(chart *database.BirthChart) error {
			chart.ID = uuid.New()
			return nil
		},
	}
	svc := NewAstrologyServiceWithClient(mockRepo, "", "", nil)

	userID := uuid.New()
	chart, err := svc.CreateBirthChart(ctx, ports.BirthChartData{
		UserID:     userID,
		BirthDate:  "1995-12-15",
		BirthTime:  "10:30",
		BirthPlace: "Bangalore",
		Latitude:   12.9716,
		Longitude:  77.5946,
		Timezone:   "Asia/Kolkata",
	})
	if err != nil {
		t.Fatalf("CreateBirthChart failed: %v", err)
	}
	if chart.UserID != userID {
		t.Fatalf("chart.UserID = %v, want %v", chart.UserID, userID)
	}
	// Dec 15 1995 — Sun should be in Sagittarius.
	if chart.SunSign != "Sagittarius" {
		t.Fatalf("expected Sagittarius for Dec 15, got %s", chart.SunSign)
	}
}

func TestCreateBirthChartInvalidDate(t *testing.T) {
	svc := NewAstrologyServiceWithClient(nil, "", "", nil)
	_, err := svc.CreateBirthChart(ctx, ports.BirthChartData{
		UserID:    uuid.New(),
		BirthDate: "not-a-date",
		BirthTime: "10:30",
	})
	if err == nil {
		t.Fatal("expected error for invalid birth_date")
	}
}

func TestCreateBirthChartInvalidTime(t *testing.T) {
	svc := NewAstrologyServiceWithClient(nil, "", "", nil)
	_, err := svc.CreateBirthChart(ctx, ports.BirthChartData{
		UserID:    uuid.New(),
		BirthDate: "1995-12-15",
		BirthTime: "99:99",
	})
	if err == nil {
		t.Fatal("expected error for invalid birth_time")
	}
}

func TestGetBirthChart(t *testing.T) {
	expected := &database.BirthChart{ID: uuid.New(), UserID: uuid.New(), SunSign: constants.Leo}
	mockRepo := &mocks.MockAstrologyRepository{
		GetBirthChartFunc: func(id uuid.UUID) (*database.BirthChart, error) {
			if id == expected.ID {
				return expected, nil
			}
			return nil, errors.New("not found")
		},
	}
	svc := NewAstrologyServiceWithClient(mockRepo, "", "", nil)
	chart, err := svc.GetBirthChart(ctx, expected.ID)
	if err != nil {
		t.Fatalf("GetBirthChart failed: %v", err)
	}
	if chart.ID != expected.ID {
		t.Fatalf("chart.ID = %v, want %v", chart.ID, expected.ID)
	}
}

func TestGetDailyHoroscope(t *testing.T) {
	mockRepo := &mocks.MockAstrologyRepository{
		GetOrCreateHoroscopeFunc: func(h *database.Horoscope) (*database.Horoscope, error) {
			h.ID = uuid.New()
			return h, nil
		},
	}
	mockHTTP := &mocks.MockHTTPClient{
		PostFunc: func(url, contentType string, body io.Reader) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"output": "Mock AI horoscope"}`))),
			}, nil
		},
	}
	svc := NewAstrologyServiceWithClient(mockRepo, "http://localhost:11434", "llama2", mockHTTP)

	horoscope, err := svc.GetHoroscope(ctx, constants.Gemini, constants.HoroscopeTypeDaily)
	if err != nil {
		t.Fatalf("GetHoroscope failed: %v", err)
	}
	if horoscope.Sign != constants.Gemini {
		t.Fatalf("Sign = %s, want Gemini", horoscope.Sign)
	}
	if horoscope.Type != constants.HoroscopeTypeDaily {
		t.Fatalf("Type = %s, want daily", horoscope.Type)
	}
}

func TestCheckCompatibility(t *testing.T) {
	chart1ID, chart2ID := uuid.New(), uuid.New()
	chart1 := &database.BirthChart{ID: chart1ID, UserID: uuid.New(), SunSign: constants.Aries}
	chart2 := &database.BirthChart{ID: chart2ID, UserID: uuid.New(), SunSign: constants.Libra}

	mockRepo := &mocks.MockAstrologyRepository{
		GetBirthChartFunc: func(id uuid.UUID) (*database.BirthChart, error) {
			if id == chart1ID {
				return chart1, nil
			}
			if id == chart2ID {
				return chart2, nil
			}
			return nil, errors.New("not found")
		},
	}
	svc := NewAstrologyServiceWithClient(mockRepo, "", "", nil)
	result, err := svc.CheckCompatibility(ctx, chart1ID, chart2ID)
	if err != nil {
		t.Fatalf("CheckCompatibility failed: %v", err)
	}
	score := result["overall_score"].(int)
	if score < 0 || score > 36 {
		t.Fatalf("overall_score %d out of range", score)
	}
}

func TestFetchOllamaPredictionSuccess(t *testing.T) {
	mockHTTP := &mocks.MockHTTPClient{
		PostFunc: func(url, contentType string, body io.Reader) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"output": "Test prediction"}`))),
			}, nil
		},
	}
	svc := NewAstrologyServiceWithClient(nil, "http://localhost:11434", "llama2", mockHTTP)
	result, err := svc.fetchOllamaPrediction("test prompt")
	if err != nil {
		t.Fatalf("fetchOllamaPrediction failed: %v", err)
	}
	if result != "Test prediction" {
		t.Fatalf("expected 'Test prediction', got %q", result)
	}
}

func TestFetchOllamaPredictionNetworkError(t *testing.T) {
	mockHTTP := &mocks.MockHTTPClient{
		PostFunc: func(url, contentType string, body io.Reader) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}
	svc := NewAstrologyServiceWithClient(nil, "http://localhost:11434", "llama2", mockHTTP)
	_, err := svc.fetchOllamaPrediction("test prompt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFetchOllamaPredictionNotConfigured(t *testing.T) {
	svc := NewAstrologyServiceWithClient(nil, "", "", nil)
	_, err := svc.fetchOllamaPrediction("test prompt")
	if err == nil {
		t.Fatal("expected error when ollama not configured")
	}
}

func TestGetRemedies(t *testing.T) {
	chart := &database.BirthChart{
		ID: uuid.New(), UserID: uuid.New(),
		SunSign: constants.Aries, MoonSign: constants.Leo, RisingSign: constants.Sagittarius,
		Planets: `[]`, Houses: `[]`, Aspects: `[]`,
	}
	mockHTTP := &mocks.MockHTTPClient{
		PostFunc: func(url, contentType string, body io.Reader) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"output": "Personalized remedies for Aries..."}`))),
			}, nil
		},
	}
	svc := NewAstrologyServiceWithClient(nil, "http://localhost:11434", "llama2", mockHTTP)
	remedies, err := svc.GetRemedies(ctx, chart)
	if err != nil {
		t.Fatalf("GetRemedies failed: %v", err)
	}
	if remedies["chart_id"] != chart.ID {
		t.Fatalf("chart_id mismatch")
	}
	if remedies["sun_sign"] != constants.Aries {
		t.Fatalf("sun_sign mismatch")
	}
	if remedies["remedies"].(string) == "" {
		t.Fatal("expected non-empty remedies")
	}
}

func TestGetRemediesFallback(t *testing.T) {
	chart := &database.BirthChart{
		ID: uuid.New(), UserID: uuid.New(),
		SunSign: constants.Aries, MoonSign: constants.Leo, RisingSign: constants.Sagittarius,
	}
	mockHTTP := &mocks.MockHTTPClient{
		PostFunc: func(url, contentType string, body io.Reader) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}
	svc := NewAstrologyServiceWithClient(nil, "", "", mockHTTP)
	remedies, err := svc.GetRemedies(ctx, chart)
	if err != nil {
		t.Fatalf("GetRemedies failed: %v", err)
	}
	text := remedies["remedies"].(string)
	if text == "" {
		t.Fatal("expected non-empty fallback remedies")
	}
	if !strings.Contains(text, "Aries") {
		t.Fatal("expected fallback to mention sun sign")
	}
}

func TestGeneratePersonalizedHoroscopeWithMock(t *testing.T) {
	chart := &database.BirthChart{
		ID: uuid.New(), UserID: uuid.New(),
		SunSign: constants.Gemini, MoonSign: constants.Aries, RisingSign: constants.Leo,
		Planets: `[]`,
	}
	mockHTTP := &mocks.MockHTTPClient{
		PostFunc: func(url, contentType string, body io.Reader) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"output": "Personalized horoscope focused on growth."}`))),
			}, nil
		},
	}
	svc := NewAstrologyServiceWithClient(nil, "http://localhost:11434", "llama2", mockHTTP)
	result, err := svc.GeneratePersonalizedHoroscope(ctx, chart, ports.PersonalizationPreferences{
		Goals:      "Build confidence",
		FocusAreas: []string{"career", "relationships"},
		Tone:       "motivating",
	})
	if err != nil {
		t.Fatalf("GeneratePersonalizedHoroscope failed: %v", err)
	}
	if result["chart_id"] != chart.ID {
		t.Fatalf("chart_id mismatch")
	}
	if result["personalized_horoscope"].(string) != "Personalized horoscope focused on growth." {
		t.Fatalf("unexpected horoscope text: %v", result["personalized_horoscope"])
	}
}

func TestGeneratePersonalizedHoroscopeFallback(t *testing.T) {
	chart := &database.BirthChart{
		ID: uuid.New(), UserID: uuid.New(),
		SunSign: constants.Pisces, MoonSign: constants.Cancer, RisingSign: constants.Scorpio,
	}
	mockHTTP := &mocks.MockHTTPClient{
		PostFunc: func(url, contentType string, body io.Reader) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}
	svc := NewAstrologyServiceWithClient(nil, "", "", mockHTTP)
	result, err := svc.GeneratePersonalizedHoroscope(ctx, chart, ports.PersonalizationPreferences{
		Goals: "Find balance",
		Tone:  "gentle",
	})
	if err != nil {
		t.Fatalf("GeneratePersonalizedHoroscope failed: %v", err)
	}
	if result["personalized_horoscope"].(string) == "" {
		t.Fatal("expected non-empty fallback horoscope")
	}
}

func TestJulianDayKnownValue(t *testing.T) {
	// J2000.0 epoch: 2000-01-01 12:00:00 UTC = JD 2451545.0
	j2000 := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	jd := julianDay(j2000)
	if jd < 2451544.9 || jd > 2451545.1 {
		t.Fatalf("julianDay(J2000) = %.4f, want ~2451545.0", jd)
	}
}

func TestNakshatraCalculation(t *testing.T) {
	// Moon at 0° = Ashwini pada 1
	name, pada := calculateNakshatra(0)
	if name != "Ashwini" {
		t.Fatalf("expected Ashwini at 0°, got %s", name)
	}
	if pada != 1 {
		t.Fatalf("expected pada 1, got %d", pada)
	}

	// Moon at 360° - epsilon = Revati
	name, _ = calculateNakshatra(359.9)
	if name != "Revati" {
		t.Fatalf("expected Revati at 359.9°, got %s", name)
	}
}
