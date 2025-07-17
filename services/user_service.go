package services

import (
	"errors"
	"fmt"

	"github.com/azainwork/core-banking-api/models"
	"github.com/azainwork/core-banking-api/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) RegisterUser(user *models.User) error {
	var existingUser models.User
	if err := s.db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		return errors.New("user with this email already exists")
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}
	user.Password = hashedPassword

	if err := s.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

func (s *UserService) LoginUser(email, password string) (string, *models.User, error) {
	var user models.User
	
	if err := s.db.Where("email = ? AND is_active = ?", email, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("invalid email or password")
		}
		return "", nil, fmt.Errorf("failed to find user: %v", err)
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return "", nil, errors.New("invalid email or password")
	}

	token, err := utils.GenerateJWTToken(user.ID.String(), user.Email)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %v", err)
	}

	return token, &user, nil
}

func (s *UserService) GetUserByID(userID string) (*models.User, error) {
	var user models.User
	
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	if err := s.db.Preload("Accounts").Where("id = ? AND is_active = ?", id, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	return &user, nil
}

func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	
	if err := s.db.Where("email = ? AND is_active = ?", email, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	return &user, nil
}

func (s *UserService) UpdateUser(userID string, updates map[string]interface{}) (*models.User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	if err := s.db.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	return s.GetUserByID(userID)
} 