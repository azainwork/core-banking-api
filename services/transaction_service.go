package services

import (
	"errors"
	"fmt"

	"github.com/azainwork/core-banking-api/models"
	"github.com/azainwork/core-banking-api/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionService struct {
	db *gorm.DB
}

func NewTransactionService(db *gorm.DB) *TransactionService {
	return &TransactionService{db: db}
}

func (s *TransactionService) CreateTransaction(transaction *models.Transaction) error {
	transaction.TransactionID = utils.GenerateTransactionID()
	
	if transaction.Status == "" {
		transaction.Status = models.TransactionStatusPending
	}

	if err := s.db.Create(transaction).Error; err != nil {
		return fmt.Errorf("failed to create transaction: %v", err)
	}

	return nil
}

func (s *TransactionService) ProcessDeposit(accountID string, amount float64, description string) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}

	account, err := s.GetAccountByID(accountID)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	transaction := &models.Transaction{
		Type:          models.TransactionTypeDeposit,
		Amount:        amount,
		Currency:      account.Currency,
		Status:        models.TransactionStatusPending,
		Description:   description,
		AccountID:     account.ID,
		BalanceBefore: account.Balance,
		BalanceAfter:  account.Balance + amount,
	}

	if err := tx.Create(transaction).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	newBalance := account.Balance + amount
	if err := tx.Model(&models.Account{}).Where("id = ?", account.ID).Update("balance", newBalance).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update account balance: %v", err)
	}

	if err := tx.Model(transaction).Update("status", models.TransactionStatusCompleted).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update transaction status: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return transaction, nil
}

func (s *TransactionService) ProcessWithdrawal(accountID string, amount float64, description string) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}

	account, err := s.GetAccountByID(accountID)
	if err != nil {
		return nil, err
	}

	if account.Balance < amount {
		return nil, errors.New("insufficient balance")
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	transaction := &models.Transaction{
		Type:          models.TransactionTypeWithdraw,
		Amount:        amount,
		Currency:      account.Currency,
		Status:        models.TransactionStatusPending,
		Description:   description,
		AccountID:     account.ID,
		BalanceBefore: account.Balance,
		BalanceAfter:  account.Balance - amount,
	}

	if err := tx.Create(transaction).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	newBalance := account.Balance - amount
	if err := tx.Model(&models.Account{}).Where("id = ?", account.ID).Update("balance", newBalance).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update account balance: %v", err)
	}

	if err := tx.Model(transaction).Update("status", models.TransactionStatusCompleted).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update transaction status: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return transaction, nil
}

func (s *TransactionService) ProcessTransfer(fromAccountID, toAccountID string, amount float64, description string) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}

	fromAccount, err := s.GetAccountByID(fromAccountID)
	if err != nil {
		return nil, err
	}

	toAccount, err := s.GetAccountByID(toAccountID)
	if err != nil {
		return nil, err
	}

	if fromAccount.ID == toAccount.ID {
		return nil, errors.New("cannot transfer to the same account")
	}

	if fromAccount.Balance < amount {
		return nil, errors.New("insufficient balance")
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	transaction := &models.Transaction{
		Type:          models.TransactionTypeTransfer,
		Amount:        amount,
		Currency:      fromAccount.Currency,
		Status:        models.TransactionStatusPending,
		Description:   description,
		AccountID:     fromAccount.ID,
		ToAccountID:   &toAccount.ID,
		BalanceBefore: fromAccount.Balance,
		BalanceAfter:  fromAccount.Balance - amount,
	}

	if err := tx.Create(transaction).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	newFromBalance := fromAccount.Balance - amount
	if err := tx.Model(&models.Account{}).Where("id = ?", fromAccount.ID).Update("balance", newFromBalance).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update source account balance: %v", err)
	}

	newToBalance := toAccount.Balance + amount
	if err := tx.Model(&models.Account{}).Where("id = ?", toAccount.ID).Update("balance", newToBalance).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update destination account balance: %v", err)
	}

	if err := tx.Model(transaction).Update("status", models.TransactionStatusCompleted).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update transaction status: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return transaction, nil
}

func (s *TransactionService) GetTransactionByID(transactionID string) (*models.Transaction, error) {
	var transaction models.Transaction
	
	id, err := uuid.Parse(transactionID)
	if err != nil {
		return nil, errors.New("invalid transaction ID")
	}

	if err := s.db.Preload("Account").Preload("ToAccount").Where("id = ?", id).First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, fmt.Errorf("failed to find transaction: %v", err)
	}

	return &transaction, nil
}

func (s *TransactionService) GetTransactionsByAccountID(accountID string, limit, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	
	accountUUID, err := uuid.Parse(accountID)
	if err != nil {
		return nil, errors.New("invalid account ID")
	}

	if limit <= 0 {
		limit = 50
	}

	if err := s.db.Preload("Account").Preload("ToAccount").
		Where("account_id = ? OR to_account_id = ?", accountUUID, accountUUID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("failed to find transactions: %v", err)
	}

	return transactions, nil
}

func (s *TransactionService) GetAccountByID(accountID string) (*models.Account, error) {
	var account models.Account
	
	id, err := uuid.Parse(accountID)
	if err != nil {
		return nil, errors.New("invalid account ID")
	}

	if err := s.db.Where("id = ? AND is_active = ?", id, true).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, fmt.Errorf("failed to find account: %v", err)
	}

	return &account, nil
} 