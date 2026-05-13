package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	Email      string    `json:"email" gorm:"uniqueIndex;not null"`
	Password   string    `json:"-" gorm:"not null"` // Don't serialize password
	Name       string    `json:"name"`
	BirthDate  time.Time `json:"birth_date"`
	BirthTime  string    `json:"birth_time"` // HH:MM format
	BirthPlace string    `json:"birth_place"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Timezone   string    `json:"timezone"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// BirthChart represents an astrological birth chart
type BirthChart struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	UserID            uuid.UUID `json:"user_id" gorm:"not null"`
	User              User      `json:"user" gorm:"foreignKey:UserID"`
	SunSign           string    `json:"sun_sign"`
	MoonSign          string    `json:"moon_sign"`
	RisingSign        string    `json:"rising_sign"`
	Nakshatra         string    `json:"nakshatra"`
	NakshatraPad      int       `json:"nakshatra_pad"`
	Ayanamsha         float64   `json:"ayanamsha"`                    // Precession correction
	Ascendant         float64   `json:"ascendant"`                    // Exact ascendant degree
	Midheaven         float64   `json:"midheaven"`                    // MC degree
	Planets           string    `json:"planets" gorm:"type:jsonb"`    // JSON string of planet positions
	Houses            string    `json:"houses" gorm:"type:jsonb"`     // JSON string of house cusps
	Aspects           string    `json:"aspects" gorm:"type:jsonb"`    // JSON string of aspects
	Yogas             string    `json:"yogas" gorm:"type:jsonb"`      // JSON string of detected yogas
	DashaData         string    `json:"dasha_data" gorm:"type:jsonb"` // JSON string of current dasha
	CalculationMethod string    `json:"calculation_method"`           // "Placidus", "Koch", "Equal"
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Horoscope represents a daily/weekly horoscope
type Horoscope struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	Sign         string    `json:"sign" gorm:"not null"`
	Type         string    `json:"type" gorm:"not null"` // daily, weekly, monthly, yearly, love
	Date         time.Time `json:"date" gorm:"not null"`
	Content      string    `json:"content" gorm:"type:text"`
	LoveRating   int       `json:"love_rating"`   // 1-10
	MoneyRating  int       `json:"money_rating"`  // 1-10
	HealthRating int       `json:"health_rating"` // 1-10
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Dasha represents planetary periods (dashas)
type Dasha struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key"`
	ChartID         uuid.UUID  `json:"chart_id" gorm:"not null"`
	Chart           BirthChart `json:"chart" gorm:"foreignKey:ChartID"`
	Type            string     `json:"type" gorm:"not null"` // vimshottari, ashtottari, etc.
	Mahadasha       string     `json:"mahadasha"`            // Planet name
	MahadashaStart  time.Time  `json:"mahadasha_start"`
	MahadashaEnd    time.Time  `json:"mahadasha_end"`
	Antardasha      string     `json:"antardasha"` // Sub-period planet
	AntardashaStart time.Time  `json:"antardasha_start"`
	AntardashaEnd   time.Time  `json:"antardasha_end"`
	PratyantarDasha string     `json:"pratyantar_dasha"` // Sub-sub-period
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Compatibility represents compatibility analysis between two charts
type Compatibility struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	ChartID1         uuid.UUID `json:"chart_id_1" gorm:"not null"`
	ChartID2         uuid.UUID `json:"chart_id_2" gorm:"not null"`
	OverallScore     int       `json:"overall_score"`      // 1-36 (Guna Milan)
	VarnaScore       int       `json:"varna_score"`        // 1-1
	VashyaScore      int       `json:"vashya_score"`       // 1-2
	TaraScore        int       `json:"tara_score"`         // 1-3
	YoniScore        int       `json:"yoni_score"`         // 1-4
	GrahaMaitriScore int       `json:"graha_maitri_score"` // 1-5
	GanaScore        int       `json:"gana_score"`         // 1-6
	BhakutScore      int       `json:"bhakut_score"`       // 1-7
	NadiScore        int       `json:"nadi_score"`         // 1-8
	Analysis         string    `json:"analysis" gorm:"type:text"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Panchang represents Hindu calendar data
type Panchang struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	Date      time.Time `json:"date" gorm:"not null"`
	Location  string    `json:"location"`
	Tithi     string    `json:"tithi"`     // Lunar day
	Nakshatra string    `json:"nakshatra"` // Lunar constellation
	Yoga      string    `json:"yoga"`      // Lunar day combination
	Karan     string    `json:"karan"`     // Half-tithi
	RahuKaal  string    `json:"rahu_kaal"` // Inauspicious timing
	Sunrise   string    `json:"sunrise"`
	Sunset    string    `json:"sunset"`
	Moonrise  string    `json:"moonrise"`
	Moonset   string    `json:"moonset"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Transit represents planetary transits
type Transit struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	Date      time.Time `json:"date" gorm:"not null"`
	Planet    string    `json:"planet"`
	Sign      string    `json:"sign"`
	Degree    float64   `json:"degree"`
	House     int       `json:"house"`
	Aspects   string    `json:"aspects" gorm:"type:jsonb"` // JSON of aspects formed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

func (d *Dasha) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

func (c *Compatibility) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (p *Panchang) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (t *Transit) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
