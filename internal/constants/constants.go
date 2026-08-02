package constants

// Zodiac signs
const (
	Aries       = "Aries"
	Taurus      = "Taurus"
	Gemini      = "Gemini"
	Cancer      = "Cancer"
	Leo         = "Leo"
	Virgo       = "Virgo"
	Libra       = "Libra"
	Scorpio     = "Scorpio"
	Sagittarius = "Sagittarius"
	Capricorn   = "Capricorn"
	Aquarius    = "Aquarius"
	Pisces      = "Pisces"
)

// Astrological elements
const (
	ElementFiery  = "fiery"
	ElementEarthy = "earthy"
	ElementAiry   = "airy"
	ElementWatery = "watery"
)

// Horoscope types
const (
	HoroscopeTypeDaily   = "daily"
	HoroscopeTypeWeekly  = "weekly"
	HoroscopeTypeMonthly = "monthly"
	HoroscopeTypeYearly  = "yearly"
	HoroscopeTypeLove    = "love"
)

// API paths
const (
	APIV1Prefix                   = "/api/v1"
	AuthPrefix                    = "/auth"
	UserPrefix                    = "/user"
	AstrologyPrefix               = "/astrology"
	HealthEndpoint                = "/health"
	RegisterEndpoint              = "/register"
	LoginEndpoint                 = "/login"
	ProfileEndpoint               = "/profile"
	BirthInfoEndpoint             = "/birth-info"
	BirthChartEndpoint            = "/birth-chart"
	DailyHoroscopeEndpoint        = "/horoscope/daily"
	HoroscopeEndpoint             = "/horoscope"
	CompatibilityEndpoint         = "/compatibility"
	RemediesEndpoint              = "/remedies"
	PersonalizedHoroscopeEndpoint = "/personalized-horoscope"
)

// Content types
const (
	ContentTypeJSON = "application/json"
)

// Time formats
const (
	DateFormat = "2006-01-02"
	TimeFormat = "15:04"
)

// Default ratings for horoscopes
const (
	DefaultLoveRating   = 7
	DefaultMoneyRating  = 6
	DefaultHealthRating = 8
)

// Ollama API
const (
	OllamaPredictionsEndpoint = "/api/predictions"
	OllamaTimeoutSeconds      = 60
)

// Server defaults (non-secret)
const (
	DefaultPort        = "8080"
	DefaultEnvironment = "production"
	DefaultOllamaURL   = "http://127.0.0.1:11434"
	DefaultOllamaModel = "llama3"
)
