// Package ports defines the inbound service interfaces (ports in Hexagonal Architecture).
// Handlers and middleware depend only on these interfaces, never on concrete service types.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
)

// AuthService defines the authentication contract consumed by handlers and middleware.
type AuthService interface {
	RegisterUser(ctx context.Context, email, password, name string) (*database.User, error)
	AuthenticateUser(ctx context.Context, email, password string) (string, *database.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*database.User, error)
	UpdateUser(ctx context.Context, user *database.User) error
	ValidateToken(token string) (uuid.UUID, error)
}

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

// PersonalizationPreferences carries optional personalization inputs.
type PersonalizationPreferences struct {
	Goals      string
	FocusAreas []string
	Tone       string
}

// AstrologyService defines the astrology domain contract consumed by handlers.
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
