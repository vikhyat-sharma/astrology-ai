package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"github.com/vikhyat-sharma/astrology-ai/internal/repositories"
)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo  *repositories.UserRepository
	jwtSecret []byte
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo *repositories.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

// claims is the typed JWT payload. Using typed claims avoids the fragile
// map[string]interface{} type assertion pattern.
type claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"uid"`
}

// RegisterUser registers a new user.
func (s *AuthService) RegisterUser(ctx context.Context, email, password, name string) (*database.User, error) {
	existing, _ := s.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, errors.New("user already exists")
	}

	hashed, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &database.User{
		Email:    email,
		Password: hashed,
		Name:     name,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// AuthenticateUser validates credentials and returns a signed JWT.
func (s *AuthService) AuthenticateUser(ctx context.Context, email, password string) (string, *database.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Return a generic message to prevent user enumeration.
		return "", nil, errors.New("invalid credentials")
	}

	if err := comparePassword(user.Password, password); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

// GetUserByID retrieves a user by their UUID.
func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*database.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// UpdateUser persists changes to a user record.
func (s *AuthService) UpdateUser(ctx context.Context, user *database.User) error {
	return s.userRepo.Update(ctx, user)
}

// ValidateToken parses and validates a JWT, returning the embedded user ID.
// This method is intentionally context-free because it is called from middleware
// before the request context is fully established.
func (s *AuthService) ValidateToken(tokenString string) (uuid.UUID, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	c, ok := parsed.Claims.(*claims)
	if !ok || !parsed.Valid {
		return uuid.Nil, errors.New("invalid token claims")
	}
	if c.UserID == uuid.Nil {
		return uuid.Nil, errors.New("token missing user ID")
	}
	return c.UserID, nil
}

// generateToken creates a signed JWT with a 15-minute expiry.
func (s *AuthService) generateToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "astrology-ai",
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			ID:        uuid.NewString(), // jti — unique per token, enables future revocation
		},
		UserID: userID,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(s.jwtSecret)
}
