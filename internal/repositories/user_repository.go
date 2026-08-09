package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/database"
	"gorm.io/gorm"
)

// UserRepository handles database operations for users.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create persists a new user.
func (r *UserRepository) Create(ctx context.Context, user *database.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID retrieves a user by UUID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*database.User, error) {
	var user database.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email address.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*database.User, error) {
	var user database.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Update saves changes to an existing user.
func (r *UserRepository) Update(ctx context.Context, user *database.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete soft-deletes a user by ID.
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&database.User{}, "id = ?", id).Error
}
