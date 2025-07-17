package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionType string

const (
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeWithdraw TransactionType = "withdraw"
	TransactionTypeTransfer TransactionType = "transfer"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

type Transaction struct {
	ID              uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TransactionID   string            `json:"transaction_id" gorm:"uniqueIndex;not null"`
	Type            TransactionType   `json:"type" gorm:"not null"`
	Amount          float64           `json:"amount" gorm:"not null"`
	Currency        string            `json:"currency" gorm:"default:'USD'"`
	Status          TransactionStatus `json:"status" gorm:"default:'pending'"`
	Description     string            `json:"description"`
	
	AccountID       uuid.UUID         `json:"account_id" gorm:"type:uuid;not null"`
	
	ToAccountID     *uuid.UUID        `json:"to_account_id,omitempty" gorm:"type:uuid"`
	
	BalanceBefore   float64           `json:"balance_before"`
	BalanceAfter    float64           `json:"balance_after"`
	
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	DeletedAt       gorm.DeletedAt    `json:"-" gorm:"index"`

	Account         Account           `json:"account,omitempty" gorm:"foreignKey:AccountID"`
	ToAccount       *Account          `json:"to_account,omitempty" gorm:"foreignKey:ToAccountID"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
} 