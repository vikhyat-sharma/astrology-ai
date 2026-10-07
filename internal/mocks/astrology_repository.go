package mocks

import (
	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
)

// MockAstrologyRepository is a test double for ports.AstrologyRepository.
// Set any Func field to override the default no-op behaviour for that method.
type MockAstrologyRepository struct {
	CreateBirthChartFunc       func(chart *database.BirthChart) error
	GetBirthChartFunc          func(id uuid.UUID) (*database.BirthChart, error)
	GetBirthChartsByUserIDFunc func(userID uuid.UUID) ([]*database.BirthChart, error)
	CreateHoroscopeFunc        func(horoscope *database.Horoscope) error
	GetHoroscopeFunc           func(sign, horoscopeType string) (*database.Horoscope, error)
	GetOrCreateHoroscopeFunc   func(h *database.Horoscope) (*database.Horoscope, error)
	GetHoroscopesBySignFunc    func(sign string) ([]*database.Horoscope, error)
	UpdateHoroscopeFunc        func(horoscope *database.Horoscope) error
}

func (m *MockAstrologyRepository) CreateBirthChart(chart *database.BirthChart) error {
	if m.CreateBirthChartFunc != nil {
		return m.CreateBirthChartFunc(chart)
	}
	return nil
}

func (m *MockAstrologyRepository) GetBirthChart(id uuid.UUID) (*database.BirthChart, error) {
	if m.GetBirthChartFunc != nil {
		return m.GetBirthChartFunc(id)
	}
	return &database.BirthChart{}, nil
}

func (m *MockAstrologyRepository) GetBirthChartsByUserID(userID uuid.UUID) ([]*database.BirthChart, error) {
	if m.GetBirthChartsByUserIDFunc != nil {
		return m.GetBirthChartsByUserIDFunc(userID)
	}
	return []*database.BirthChart{}, nil
}

func (m *MockAstrologyRepository) CreateHoroscope(horoscope *database.Horoscope) error {
	if m.CreateHoroscopeFunc != nil {
		return m.CreateHoroscopeFunc(horoscope)
	}
	return nil
}

func (m *MockAstrologyRepository) GetHoroscope(sign, horoscopeType string) (*database.Horoscope, error) {
	if m.GetHoroscopeFunc != nil {
		return m.GetHoroscopeFunc(sign, horoscopeType)
	}
	return nil, nil
}

func (m *MockAstrologyRepository) GetOrCreateHoroscope(h *database.Horoscope) (*database.Horoscope, error) {
	if m.GetOrCreateHoroscopeFunc != nil {
		return m.GetOrCreateHoroscopeFunc(h)
	}
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return h, nil
}

func (m *MockAstrologyRepository) GetHoroscopesBySign(sign string) ([]*database.Horoscope, error) {
	if m.GetHoroscopesBySignFunc != nil {
		return m.GetHoroscopesBySignFunc(sign)
	}
	return []*database.Horoscope{}, nil
}

func (m *MockAstrologyRepository) UpdateHoroscope(horoscope *database.Horoscope) error {
	if m.UpdateHoroscopeFunc != nil {
		return m.UpdateHoroscopeFunc(horoscope)
	}
	return nil
}
