package repositories

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"gorm.io/gorm"
)

// AstrologyRepository handles database operations for astrology data
type AstrologyRepository struct {
	db *gorm.DB
}

// NewAstrologyRepository creates a new astrology repository
func NewAstrologyRepository(db *gorm.DB) *AstrologyRepository {
	return &AstrologyRepository{db: db}
}

// CreateBirthChart creates a new birth chart
func (r *AstrologyRepository) CreateBirthChart(chart *database.BirthChart) error {
	return r.db.Create(chart).Error
}

// GetBirthChart gets a birth chart by ID
func (r *AstrologyRepository) GetBirthChart(id uuid.UUID) (*database.BirthChart, error) {
	var chart database.BirthChart
	err := r.db.Preload("User").First(&chart, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("birth chart not found")
		}
		return nil, err
	}
	return &chart, nil
}

// GetBirthChartsByUserID gets all birth charts for a user
func (r *AstrologyRepository) GetBirthChartsByUserID(userID uuid.UUID) ([]*database.BirthChart, error) {
	var charts []*database.BirthChart
	err := r.db.Where("user_id = ?", userID).Find(&charts).Error
	return charts, err
}

// CreateHoroscope creates a new horoscope
func (r *AstrologyRepository) CreateHoroscope(horoscope *database.Horoscope) error {
	return r.db.Create(horoscope).Error
}

// GetHoroscope gets a horoscope by sign and type for today
func (r *AstrologyRepository) GetHoroscope(sign, horoscopeType string) (*database.Horoscope, error) {
	var horoscope database.Horoscope
	today := time.Now().Truncate(24 * time.Hour)
	err := r.db.Where("sign = ? AND type = ? AND date = ?", sign, horoscopeType, today).First(&horoscope).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("horoscope not found")
		}
		return nil, err
	}
	return &horoscope, nil
}

// GetHoroscopesBySign gets all horoscopes for a sign
func (r *AstrologyRepository) GetHoroscopesBySign(sign string) ([]*database.Horoscope, error) {
	var horoscopes []*database.Horoscope
	err := r.db.Where("sign = ?", sign).Order("date DESC").Find(&horoscopes).Error
	return horoscopes, err
}

// UpdateHoroscope updates a horoscope
func (r *AstrologyRepository) UpdateHoroscope(horoscope *database.Horoscope) error {
	return r.db.Save(horoscope).Error
}