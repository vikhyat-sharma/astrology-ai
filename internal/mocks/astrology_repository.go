package mocks

import (
	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
)

// MockAstrologyRepository is a mock implementation of interfaces.AstrologyRepositoryInterface
type MockAstrologyRepository struct {
	CreateBirthChartFunc       func(chart *database.BirthChart) error
	GetBirthChartFunc          func(id uuid.UUID) (*database.BirthChart, error)
	GetBirthChartsByUserIDFunc func(userID uuid.UUID) ([]*database.BirthChart, error)
	CreateHoroscopeFunc        func(horoscope *database.Horoscope) error
	GetHoroscopeFunc           func(sign, horoscopeType string) (*database.Horoscope, error)
	GetHoroscopesBySignFunc    func(sign string) ([]*database.Horoscope, error)
	UpdateHoroscopeFunc        func(horoscope *database.Horoscope) error
}

// CreateBirthChart mocks the CreateBirthChart method
func (m *MockAstrologyRepository) CreateBirthChart(chart *database.BirthChart) error {
	if m.CreateBirthChartFunc != nil {
		return m.CreateBirthChartFunc(chart)
	}
	return nil
}

// GetBirthChart mocks the GetBirthChart method
func (m *MockAstrologyRepository) GetBirthChart(id uuid.UUID) (*database.BirthChart, error) {
	if m.GetBirthChartFunc != nil {
		return m.GetBirthChartFunc(id)
	}
	return &database.BirthChart{}, nil
}

// GetBirthChartsByUserID mocks the GetBirthChartsByUserID method
func (m *MockAstrologyRepository) GetBirthChartsByUserID(userID uuid.UUID) ([]*database.BirthChart, error) {
	if m.GetBirthChartsByUserIDFunc != nil {
		return m.GetBirthChartsByUserIDFunc(userID)
	}
	return []*database.BirthChart{}, nil
}

// CreateHoroscope mocks the CreateHoroscope method
func (m *MockAstrologyRepository) CreateHoroscope(horoscope *database.Horoscope) error {
	if m.CreateHoroscopeFunc != nil {
		return m.CreateHoroscopeFunc(horoscope)
	}
	return nil
}

// GetHoroscope mocks the GetHoroscope method
func (m *MockAstrologyRepository) GetHoroscope(sign, horoscopeType string) (*database.Horoscope, error) {
	if m.GetHoroscopeFunc != nil {
		return m.GetHoroscopeFunc(sign, horoscopeType)
	}
	return nil, nil
}

// GetHoroscopesBySign mocks the GetHoroscopesBySign method
func (m *MockAstrologyRepository) GetHoroscopesBySign(sign string) ([]*database.Horoscope, error) {
	if m.GetHoroscopesBySignFunc != nil {
		return m.GetHoroscopesBySignFunc(sign)
	}
	return []*database.Horoscope{}, nil
}

// UpdateHoroscope mocks the UpdateHoroscope method
func (m *MockAstrologyRepository) UpdateHoroscope(horoscope *database.Horoscope) error {
	if m.UpdateHoroscopeFunc != nil {
		return m.UpdateHoroscopeFunc(horoscope)
	}
	return nil
}
