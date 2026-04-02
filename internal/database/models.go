package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"` // Don't serialize password
	Name      string    `json:"name"`
	BirthDate time.Time `json:"birth_date"`
	BirthTime string    `json:"birth_time"` // HH:MM format
	BirthPlace string   `json:"birth_place"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timezone  string    `json:"timezone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BirthChart represents an astrological birth chart
type BirthChart struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"not null"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	SunSign   string    `json:"sun_sign"`
	MoonSign  string    `json:"moon_sign"`
	RisingSign string   `json:"rising_sign"`
	Planets   string    `json:"planets" gorm:"type:jsonb"` // JSON string of planet positions
	Houses    string    `json:"houses" gorm:"type:jsonb"`  // JSON string of house cusps
	Aspects   string    `json:"aspects" gorm:"type:jsonb"` // JSON string of aspects
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Horoscope represents a daily/weekly horoscope
type Horoscope struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Sign        string    `json:"sign" gorm:"not null"`
	Type        string    `json:"type" gorm:"not null"` // daily, weekly, monthly
	Date        time.Time `json:"date" gorm:"not null"`
	Content     string    `json:"content" gorm:"type:text"`
	LoveRating  int       `json:"love_rating"`  // 1-10
	MoneyRating int       `json:"money_rating"` // 1-10
	HealthRating int      `json:"health_rating"` // 1-10
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (b *BirthChart) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

func (h *Horoscope) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}