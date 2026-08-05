package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"gorm.io/gorm"
)

// AstrologyRepository handles database operations for astrology data.
type AstrologyRepository struct {
	db *gorm.DB
}

// NewAstrologyRepository creates a new AstrologyRepository.
func NewAstrologyRepository(db *gorm.DB) *AstrologyRepository {
	return &AstrologyRepository{db: db}
}

// CreateBirthChart persists a new birth chart.
func (r *AstrologyRepository) CreateBirthChart(chart *database.BirthChart) error {
	return r.db.Create(chart).Error
}

// GetBirthChart retrieves a birth chart by ID.
// The User association is intentionally NOT preloaded to prevent password hash exposure.
func (r *AstrologyRepository) GetBirthChart(id uuid.UUID) (*database.BirthChart, error) {
	var chart database.BirthChart
	err := r.db.First(&chart, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("birth chart not found")
		}
		return nil, err
	}
	return &chart, nil
}

// GetBirthChartsByUserID retrieves all birth charts for a user.
func (r *AstrologyRepository) GetBirthChartsByUserID(userID uuid.UUID) ([]*database.BirthChart, error) {
	var charts []*database.BirthChart
	err := r.db.Where("user_id = ?", userID).Find(&charts).Error
	return charts, err
}

// CreateHoroscope persists a new horoscope.
func (r *AstrologyRepository) CreateHoroscope(horoscope *database.Horoscope) error {
	return r.db.Create(horoscope).Error
}

// GetHoroscope retrieves today's horoscope for a sign and type.
func (r *AstrologyRepository) GetHoroscope(sign, horoscopeType string) (*database.Horoscope, error) {
	var h database.Horoscope
	today := time.Now().UTC().Truncate(24 * time.Hour)
	err := r.db.Where("sign = ? AND type = ? AND date = ?", sign, horoscopeType, today).First(&h).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("horoscope not found")
		}
		return nil, err
	}
	return &h, nil
}

// GetOrCreateHoroscope atomically returns an existing horoscope or creates a new one.
// This eliminates the check-then-insert race condition under concurrent load.
// A UNIQUE constraint on (sign, type, date) must exist in the schema.
func (r *AstrologyRepository) GetOrCreateHoroscope(h *database.Horoscope) (*database.Horoscope, error) {
	result := r.db.
		Where(database.Horoscope{Sign: h.Sign, Type: h.Type, Date: h.Date}).
		Attrs(database.Horoscope{
			Content:      h.Content,
			LoveRating:   h.LoveRating,
			MoneyRating:  h.MoneyRating,
			HealthRating: h.HealthRating,
		}).
		FirstOrCreate(h)
	return h, result.Error
}

// GetHoroscopesBySign retrieves all horoscopes for a sign, newest first.
func (r *AstrologyRepository) GetHoroscopesBySign(sign string) ([]*database.Horoscope, error) {
	var horoscopes []*database.Horoscope
	err := r.db.Where("sign = ?", sign).Order("date DESC").Find(&horoscopes).Error
	return horoscopes, err
}

// UpdateHoroscope saves changes to an existing horoscope.
func (r *AstrologyRepository) UpdateHoroscope(horoscope *database.Horoscope) error {
	return r.db.Save(horoscope).Error
}

// WithContext returns a repository scoped to the given context (for cancellation/tracing).
func (r *AstrologyRepository) WithContext(ctx context.Context) *AstrologyRepository {
	return &AstrologyRepository{db: r.db.WithContext(ctx)}
}

