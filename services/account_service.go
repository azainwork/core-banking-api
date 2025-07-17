package services

import (
	"errors"
	"fmt"

	"github.com/azainwork/core-banking-api/models"
	"github.com/azainwork/core-banking-api/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountService struct {
	db *gorm.DB
}

func NewAccountService(db *gorm.DB) *AccountService {
	return &AccountService{db: db}
}

func (s *AccountService) CreateAccount(userID string, accountType models.AccountType, initialBalance float64) (*models.Account, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	if accountType != models.AccountTypeChecking && accountType != models.AccountTypeSaving {
		return nil, errors.New("invalid account type")
	}

	var user models.User
	if err := s.db.Where("id = ? AND is_active = ?", userUUID, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	accountNumber := utils.GenerateAccountNumber()

	account := &models.Account{
		AccountNumber: accountNumber,
		Type:         accountType,
		Balance:      initialBalance,
		Currency:     "USD",
		IsActive:     true,
		UserID:       userUUID,
	}

	if err := s.db.Create(account).Error; err != nil {
		return nil, fmt.Errorf("failed to create account: %v", err)
	}

	return account, nil
}

func (s *AccountService) GetAccountByID(accountID string) (*models.Account, error) {
	var account models.Account
	
	id, err := uuid.Parse(accountID)
	if err != nil {
		return nil, errors.New("invalid account ID")
	}

	if err := s.db.Preload("User").Where("id = ? AND is_active = ?", id, true).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, fmt.Errorf("failed to find account: %v", err)
	}

	return &account, nil
}

func (s *AccountService) GetAccountsByUserID(userID string) ([]models.Account, error) {
	var accounts []models.Account
	
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	if err := s.db.Where("user_id = ? AND is_active = ?", userUUID, true).Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("failed to find accounts: %v", err)
	}

	return accounts, nil
}

func (s *AccountService) GetAccountByNumber(accountNumber string) (*models.Account, error) {
	var account models.Account
	
	if err := s.db.Preload("User").Where("account_number = ? AND is_active = ?", accountNumber, true).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, fmt.Errorf("failed to find account: %v", err)
	}

	return &account, nil
}

func (s *AccountService) UpdateAccountBalance(accountID string, newBalance float64) error {
	id, err := uuid.Parse(accountID)
	if err != nil {
		return errors.New("invalid account ID")
	}

	if err := s.db.Model(&models.Account{}).Where("id = ?", id).Update("balance", newBalance).Error; err != nil {
		return fmt.Errorf("failed to update account balance: %v", err)
	}

	return nil
}

func (s *AccountService) ValidateAccountOwnership(accountID, userID string) error {
	accountUUID, err := uuid.Parse(accountID)
	if err != nil {
		return errors.New("invalid account ID")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	var count int64
	if err := s.db.Model(&models.Account{}).Where("id = ? AND user_id = ? AND is_active = ?", accountUUID, userUUID, true).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to validate account ownership: %v", err)
	}

	if count == 0 {
		return errors.New("account not found or access denied")
	}

	return nil
} 