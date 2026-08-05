package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a registered user.
type User struct {
	ID         uuid.UUID `json:"id"          gorm:"type:uuid;primaryKey"`
	Email      string    `json:"email"       gorm:"uniqueIndex;not null"`
	Password   string    `json:"-"           gorm:"not null"` // never serialised
	Name       string    `json:"name"`
	BirthDate  time.Time `json:"birth_date"`
	BirthTime  string    `json:"birth_time"`  // HH:MM
	BirthPlace string    `json:"birth_place"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Timezone   string    `json:"timezone"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// BirthChart represents an astrological birth chart.
// The User association field is intentionally absent — loading it via Preload
// would expose the bcrypt password hash. Fetch the user separately when needed.
type BirthChart struct {
	ID                uuid.UUID `json:"id"                 gorm:"type:uuid;primaryKey"`
	UserID            uuid.UUID `json:"user_id"            gorm:"not null;index"`
	SunSign           string    `json:"sun_sign"`
	MoonSign          string    `json:"moon_sign"`
	RisingSign        string    `json:"rising_sign"`
	Nakshatra         string    `json:"nakshatra"`
	NakshatraPad      int       `json:"nakshatra_pad"`
	Ayanamsha         float64   `json:"ayanamsha"`
	Ascendant         float64   `json:"ascendant"`
	Midheaven         float64   `json:"midheaven"`
	Planets           string    `json:"planets"            gorm:"type:jsonb"`
	Houses            string    `json:"houses"             gorm:"type:jsonb"`
	Aspects           string    `json:"aspects"            gorm:"type:jsonb"`
	Yogas             string    `json:"yogas"              gorm:"type:jsonb"`
	DashaData         string    `json:"dasha_data"         gorm:"type:jsonb"`
	CalculationMethod string    `json:"calculation_method"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Horoscope represents a generated horoscope for a sign, type, and date.
// The unique index on (sign, type, date) backs the atomic GetOrCreateHoroscope upsert
// and prevents duplicate rows under concurrent load.
type Horoscope struct {
	ID           uuid.UUID `json:"id"            gorm:"type:uuid;primaryKey"`
	Sign         string    `json:"sign"          gorm:"not null;uniqueIndex:idx_horoscope_sign_type_date"`
	Type         string    `json:"type"          gorm:"not null;uniqueIndex:idx_horoscope_sign_type_date"`
	Date         time.Time `json:"date"          gorm:"not null;uniqueIndex:idx_horoscope_sign_type_date"`
	Content      string    `json:"content"       gorm:"type:text"`
	LoveRating   int       `json:"love_rating"`
	MoneyRating  int       `json:"money_rating"`
	HealthRating int       `json:"health_rating"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Dasha represents a Vimshottari dasha period for a birth chart.
type Dasha struct {
	ID              uuid.UUID `json:"id"               gorm:"type:uuid;primaryKey"`
	ChartID         uuid.UUID `json:"chart_id"         gorm:"not null;index"`
	Type            string    `json:"type"             gorm:"not null"` // vimshottari
	Mahadasha       string    `json:"mahadasha"`
	MahadashaStart  time.Time `json:"mahadasha_start"`
	MahadashaEnd    time.Time `json:"mahadasha_end"`
	Antardasha      string    `json:"antardasha"`
	AntardashaStart time.Time `json:"antardasha_start"`
	AntardashaEnd   time.Time `json:"antardasha_end"`
	PratyantarDasha string    `json:"pratyantar_dasha"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Compatibility represents a Guna Milan compatibility result between two charts.
type Compatibility struct {
	ID               uuid.UUID `json:"id"               gorm:"type:uuid;primaryKey"`
	ChartID1         uuid.UUID `json:"chart_id_1"       gorm:"not null;index"`
	ChartID2         uuid.UUID `json:"chart_id_2"       gorm:"not null;index"`
	OverallScore     int       `json:"overall_score"`
	VarnaScore       int       `json:"varna_score"`
	VashyaScore      int       `json:"vashya_score"`
	TaraScore        int       `json:"tara_score"`
	YoniScore        int       `json:"yoni_score"`
	GrahaMaitriScore int       `json:"graha_maitri_score"`
	GanaScore        int       `json:"gana_score"`
	BhakutScore      int       `json:"bhakut_score"`
	NadiScore        int       `json:"nadi_score"`
	Analysis         string    `json:"analysis"         gorm:"type:text"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Panchang represents Hindu calendar (Panchanga) data for a date and location.
type Panchang struct {
	ID        uuid.UUID `json:"id"        gorm:"type:uuid;primaryKey"`
	Date      time.Time `json:"date"      gorm:"not null"`
	Location  string    `json:"location"`
	Tithi     string    `json:"tithi"`
	Nakshatra string    `json:"nakshatra"`
	Yoga      string    `json:"yoga"`
	Karan     string    `json:"karan"`
	RahuKaal  string    `json:"rahu_kaal"`
	Sunrise   string    `json:"sunrise"`
	Sunset    string    `json:"sunset"`
	Moonrise  string    `json:"moonrise"`
	Moonset   string    `json:"moonset"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Transit represents a recorded planetary transit position.
type Transit struct {
	ID        uuid.UUID `json:"id"      gorm:"type:uuid;primaryKey"`
	Date      time.Time `json:"date"    gorm:"not null;index"`
	Planet    string    `json:"planet"`
	Sign      string    `json:"sign"`
	Degree    float64   `json:"degree"`
	House     int       `json:"house"`
	Aspects   string    `json:"aspects" gorm:"type:jsonb"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate hooks assign a UUID primary key if one has not been set.

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (b *BirthChart) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

func (h *Horoscope) BeforeCreate(_ *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}

func (d *Dasha) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

func (c *Compatibility) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (p *Panchang) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (t *Transit) BeforeCreate(_ *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
