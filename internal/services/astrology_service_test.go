package services

import (
	"bytes"
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
)

func TestCalculateSunSign(t *testing.T) {
	tests := []struct {
		name string
		date string
		want string
	}{
		{"AriesStart", "2023-03-21", constants.Aries},
		{"AriesEnd", "2023-04-19", constants.Aries},
		{"TaurusStart", "2023-04-20", constants.Taurus},
		{"CapricornEnd", "2023-01-19", constants.Capricorn},
	}

	service := &AstrologyService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := time.Parse(constants.DateFormat, tt.date)
			if err != nil {
				t.Fatal(err)
			}
			if got := service.calculateSunSign(parsed); got != tt.want {
				t.Fatalf("calculateSunSign() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetSignElement(t *testing.T) {
	service := &AstrologyService{}

	signTests := map[string]string{
		constants.Aries:  constants.ElementFiery,
		constants.Virgo:  constants.ElementEarthy,
		constants.Gemini: constants.ElementAiry,
		constants.Pisces: constants.ElementWatery,
	}

	for sign, want := range signTests {
		if got := service.getSignElement(sign); got != want {
			t.Fatalf("getSignElement(%q) = %q, want %q", sign, got, want)
		}
	}
}

func TestGenerateDailyHoroscopeWithMock(t *testing.T) {
	mockHTTPClient := &mocks.MockHTTPClient{
		PostFunc: func(url, contentType string, body io.Reader) (*http.Response, error) {
			return &http.Response{
				StatusCode: constants.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"output": "AI generated horoscope content"}`))),
			}, nil
		},
	}

	service := NewAstrologyServiceWithClient(nil, "http://localhost:11434", "llama2", mockHTTPClient)
	out := service.generateDailyHoroscope(constants.Leo)

	if out != "AI generated horoscope content" {
		t.Fatalf("expected AI content, got %q", out)
	}
}

func TestGenerateDailyHoroscopeFallback(t *testing.T) {
	mockHTTPClient := &mocks.MockHTTPClient{
		PostFunc: func(url, contentType string, body io.Reader) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}

	service := NewAstrologyServiceWithClient(nil, "", "", mockHTTPClient)
	out := service.generateDailyHoroscope(constants.Leo)
	if out == "" {
		t.Fatal("expected non-empty fallback horoscope")
	}
}

func TestCalculateCompatibility(t *testing.T) {
	service := &AstrologyService{}
	chart1 := &database.BirthChart{SunSign: constants.Aries}
	chart2 := &database.BirthChart{SunSign: constants.Libra}

	result := service.calculateCompatibility(chart1, chart2)
	if result["score"].(int) != 75 {
		t.Fatalf("expected score 75, got %v", result["score"])
	}
	if result["summary"].(string) == "" {
		t.Fatal("expected non-empty compatibility summary")
	}
	if len(result["sun_signs"].([]string)) != 2 {
		t.Fatalf("expected 2 sun signs, got %v", result["sun_signs"])
	}
}

func TestCreateBirthChart(t *testing.T) {
	mockRepo := &mocks.MockAstrologyRepository{
		CreateBirthChartFunc: func(chart *database.BirthChart) error {
			chart.ID = uuid.New()
			return nil
		},
	}

	service := NewAstrologyServiceWithClient(mockRepo, "", "", nil)

	input := BirthChartData{
		UserID:     uuid.New(),
		BirthDate:  time.Date(1995, time.December, 15, 0, 0, 0, 0, time.UTC),
		BirthTime:  "10:30",
		BirthPlace: "Bangalore",
		Latitude:   12.9716,
		Longitude:  77.5946,
		Timezone:   "Asia/Kolkata",
	}

	chart, err := service.CreateBirthChart(input)
	if err != nil {
		t.Fatalf("CreateBirthChart failed: %v", err)
	}

	if chart.UserID != input.UserID {
		t.Fatalf("chart.UserID = %v, want %v", chart.UserID, input.UserID)
	}

	if chart.SunSign != "Sagittarius" {
		t.Fatalf("expected Sagittarius, got %s", chart.SunSign)
	}
}

func TestGetBirthChart(t *testing.T) {
	expectedChart := &database.BirthChart{
		ID:      uuid.New(),
		UserID:  uuid.New(),
		SunSign: constants.Leo,
	}

	mockRepo := &mocks.MockAstrologyRepository{
		GetBirthChartFunc: func(id uuid.UUID) (*database.BirthChart, error) {
			if id == expectedChart.ID {
				return expectedChart, nil
			}
			return nil, errors.New("chart not found")
		},
	}

	service := NewAstrologyServiceWithClient(mockRepo, "", "", nil)

	chart, err := service.GetBirthChart(expectedChart.ID)
	if err != nil {
		t.Fatalf("GetBirthChart failed: %v", err)
	}

	if chart.ID != expectedChart.ID {
		t.Fatalf("chart.ID = %v, want %v", chart.ID, expectedChart.ID)
	}
}

func TestGetDailyHoroscope(t *testing.T) {
	mockRepo := &mocks.MockAstrologyRepository{
		GetHoroscopeFunc: func(sign, horoscopeType string) (*database.Horoscope, error) {
			return nil, errors.New("horoscope not found")
		},
		CreateHoroscopeFunc: func(horoscope *database.Horoscope) error {
			horoscope.ID = uuid.New()
			return nil
		},
	}

	mockHTTPClient := &mocks.MockHTTPClient{
		PostFunc: func(url, contentType string, body io.Reader) (*http.Response, error) {
			return &http.Response{
				StatusCode: constants.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"output": "Mock AI horoscope"}`))),
			}, nil
		},
	}

	service := NewAstrologyServiceWithClient(mockRepo, "http://localhost:11434", "llama2", mockHTTPClient)

	horoscope, err := service.GetDailyHoroscope(constants.Gemini)
	if err != nil {
		t.Fatalf("GetDailyHoroscope failed: %v", err)
	}

	if horoscope.Sign != constants.Gemini {
		t.Fatalf("horoscope.Sign = %s, want Gemini", horoscope.Sign)
	}

	if horoscope.Type != constants.HoroscopeTypeDaily {
		t.Fatalf("horoscope.Type = %s, want daily", horoscope.Type)
	}
}

func TestCheckCompatibility(t *testing.T) {
	chart1ID := uuid.New()
	chart2ID := uuid.New()

	chart1 := &database.BirthChart{
		ID:      chart1ID,
		UserID:  uuid.New(),
		SunSign: constants.Aries,
	}

	chart2 := &database.BirthChart{
		ID:      chart2ID,
		UserID:  uuid.New(),
		SunSign: constants.Libra,
	}

	mockRepo := &mocks.MockAstrologyRepository{
		GetBirthChartFunc: func(id uuid.UUID) (*database.BirthChart, error) {
			if id == chart1ID {
				return chart1, nil
			}
			if id == chart2ID {
				return chart2, nil
			}
			return nil, errors.New("chart not found")
		},
	}

	service := NewAstrologyServiceWithClient(mockRepo, "", "", nil)

	result, err := service.CheckCompatibility(chart1ID, chart2ID)
	if err != nil {
		t.Fatalf("CheckCompatibility failed: %v", err)
	}

	if result["compatibility"].(map[string]interface{})["score"].(int) != 75 {
		t.Fatalf("expected compatibility score 75")
	}
}

func TestFetchOllamaPredictionSuccess(t *testing.T) {
	mockHTTPClient := &mocks.MockHTTPClient{
		PostFunc: func(url, contentType string, body io.Reader) (*http.Response, error) {
			return &http.Response{
				StatusCode: constants.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"output": "Test prediction"}`))),
			}, nil
		},
	}

	service := NewAstrologyServiceWithClient(nil, "http://localhost:11434", "llama2", mockHTTPClient)

	result, err := service.fetchOllamaPrediction("test prompt")
	if err != nil {
		t.Fatalf("fetchOllamaPrediction failed: %v", err)
	}

	if result != "Test prediction" {
		t.Fatalf("expected 'Test prediction', got %q", result)
	}
}

func TestFetchOllamaPredictionError(t *testing.T) {
	mockHTTPClient := &mocks.MockHTTPClient{
		PostFunc: func(url, contentType string, body io.Reader) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}

	service := NewAstrologyServiceWithClient(nil, "", "", mockHTTPClient)

	_, err := service.fetchOllamaPrediction("test prompt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetRemedies(t *testing.T) {
	chart := &database.BirthChart{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		SunSign:    constants.Aries,
		MoonSign:   constants.Leo,
		RisingSign: constants.Sagittarius,
		Planets:    `{"sun": {"sign": "Aries", "degree": "15°23'"}}`,
		Houses:     `{"1": "0° Aries"}`,
		Aspects:    `[{"planet1": "sun", "planet2": "moon", "aspect": "trine"}]`,
	}

	mockHTTPClient := &mocks.MockHTTPClient{
		PostFunc: func(url, contentType string, body io.Reader) (*http.Response, error) {
			return &http.Response{
				StatusCode: constants.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"output": "Personalized remedies based on your Aries sun sign..."}`))),
			}, nil
		},
	}

	service := NewAstrologyServiceWithClient(nil, "http://localhost:11434", "llama2", mockHTTPClient)

	remedies, err := service.GetRemedies(chart)
	if err != nil {
		t.Fatalf("GetRemedies failed: %v", err)
	}

	if remedies["chart_id"] != chart.ID {
		t.Fatalf("expected chart_id %v, got %v", chart.ID, remedies["chart_id"])
	}

	if remedies["sun_sign"] != constants.Aries {
		t.Fatalf("expected sun_sign %s, got %v", constants.Aries, remedies["sun_sign"])
	}

	if remedies["remedies"].(string) == "" {
		t.Fatal("expected non-empty remedies")
	}
}

func TestGetRemediesFallback(t *testing.T) {
	chart := &database.BirthChart{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		SunSign:    constants.Aries,
		MoonSign:   constants.Leo,
		RisingSign: constants.Sagittarius,
		Planets:    `{"sun": {"sign": "Aries", "degree": "15°23'"}}`,
		Houses:     `{"1": "0° Aries"}`,
		Aspects:    `[{"planet1": "sun", "planet2": "moon", "aspect": "trine"}]`,
	}

	mockHTTPClient := &mocks.MockHTTPClient{
		PostFunc: func(url, contentType string, body io.Reader) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}

	service := NewAstrologyServiceWithClient(nil, "", "", mockHTTPClient)

	remedies, err := service.GetRemedies(chart)
	if err != nil {
		t.Fatalf("GetRemedies failed: %v", err)
	}

	if remedies["remedies"].(string) == "" {
		t.Fatal("expected non-empty fallback remedies")
	}

	// Check that fallback contains the sun sign
	remediesText := remedies["remedies"].(string)
	if !strings.Contains(remediesText, "Aries") {
		t.Fatal("expected fallback remedies to contain sun sign")
	}
}
