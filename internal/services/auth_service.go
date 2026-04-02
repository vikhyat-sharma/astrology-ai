package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"github.com/vikhyat-sharma/astrology-ai/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo  *repositories.UserRepository
	jwtSecret string
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repositories.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// RegisterUser registers a new user
func (s *AuthService) RegisterUser(email, password, name string) (*database.User, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(email)
	if existingUser != nil {
		return nil, errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &database.User{
		Email:    email,
		Password: string(hashedPassword),
		Name:     name,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// AuthenticateUser authenticates a user and returns a JWT token
func (s *AuthService) AuthenticateUser(email, password string) (string, *database.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// GetUserByID gets a user by ID
func (s *AuthService) GetUserByID(id uuid.UUID) (*database.User, error) {
	return s.userRepo.GetByID(id)
}

// UpdateUser updates user information
func (s *AuthService) UpdateUser(user *database.User) error {
	return s.userRepo.Update(user)
}

// generateToken generates a JWT token for the user
func (s *AuthService) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ValidateToken validates a JWT token and returns the user ID
func (s *AuthService) ValidateToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return uuid.Nil, errors.New("invalid token claims")
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return uuid.Nil, errors.New("invalid user ID in token")
		}

		return userID, nil
	}

	return uuid.Nil, errors.New("invalid token")
}