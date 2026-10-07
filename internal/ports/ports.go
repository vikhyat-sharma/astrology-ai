// Package ports defines all service and repository interfaces (contracts).
// Handlers, middleware, and services depend only on these interfaces — never on concrete types.
// This makes it easy to swap implementations or write tests with mocks.
package ports

import (
	"context"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
)

// --- Inbound ports (consumed by handlers / middleware) ---

// AuthService is the contract for authentication logic.
type AuthService interface {
	RegisterUser(ctx context.Context, email, password, name string) (*database.User, error)
	AuthenticateUser(ctx context.Context, email, password string) (string, *database.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*database.User, error)
	UpdateUser(ctx context.Context, user *database.User) error
	ValidateToken(token string) (uuid.UUID, error)
}

// AstrologyService is the contract for astrology domain logic.
type AstrologyService interface {
	CreateBirthChart(ctx context.Context, data BirthChartData) (*database.BirthChart, error)
	GetBirthChart(ctx context.Context, id uuid.UUID) (*database.BirthChart, error)
	GetUserBirthCharts(ctx context.Context, userID uuid.UUID) ([]*database.BirthChart, error)
	GetHoroscope(ctx context.Context, sign, horoscopeType string) (*database.Horoscope, error)
	CheckCompatibility(ctx context.Context, chartID1, chartID2 uuid.UUID) (map[string]interface{}, error)
	GetRemedies(ctx context.Context, chart *database.BirthChart) (map[string]interface{}, error)
	GeneratePersonalizedHoroscope(ctx context.Context, chart *database.BirthChart, prefs PersonalizationPreferences) (map[string]interface{}, error)
	GetCurrentTransits(ctx context.Context, lat, lng float64) ([]map[string]interface{}, error)
}

// --- Outbound ports (consumed by services) ---

// AstrologyRepository is the contract for astrology data access.
type AstrologyRepository interface {
	CreateBirthChart(chart *database.BirthChart) error
	GetBirthChart(id uuid.UUID) (*database.BirthChart, error)
	GetBirthChartsByUserID(userID uuid.UUID) ([]*database.BirthChart, error)
	CreateHoroscope(horoscope *database.Horoscope) error
	GetHoroscope(sign, horoscopeType string) (*database.Horoscope, error)
	GetOrCreateHoroscope(h *database.Horoscope) (*database.Horoscope, error)
	GetHoroscopesBySign(sign string) ([]*database.Horoscope, error)
	UpdateHoroscope(horoscope *database.Horoscope) error
}

// HTTPClient is the contract for making HTTP requests (used by the Ollama client).
type HTTPClient interface {
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

// --- Input / output data types ---

// BirthChartData carries the inputs required to create a birth chart.
type BirthChartData struct {
	UserID     uuid.UUID
	BirthDate  string // YYYY-MM-DD
	BirthTime  string // HH:MM
	BirthPlace string
	Latitude   float64
	Longitude  float64
	Timezone   string
}

// PersonalizationPreferences carries optional inputs for personalized horoscopes.
type PersonalizationPreferences struct {
	Goals      string
	FocusAreas []string
	Tone       string
}
