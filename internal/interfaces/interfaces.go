package interfaces

import (
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
)

// HTTPClientInterface defines the interface for HTTP client operations.
type HTTPClientInterface interface {
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

// AstrologyRepositoryInterface defines the data access contract for astrology data.
type AstrologyRepositoryInterface interface {
	CreateBirthChart(chart *database.BirthChart) error
	GetBirthChart(id uuid.UUID) (*database.BirthChart, error)
	GetBirthChartsByUserID(userID uuid.UUID) ([]*database.BirthChart, error)
	CreateHoroscope(horoscope *database.Horoscope) error
	GetHoroscope(sign, horoscopeType string) (*database.Horoscope, error)
	GetOrCreateHoroscope(h *database.Horoscope) (*database.Horoscope, error)
	GetHoroscopesBySign(sign string) ([]*database.Horoscope, error)
	UpdateHoroscope(horoscope *database.Horoscope) error
}
