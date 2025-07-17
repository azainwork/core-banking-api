package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountType string

const (
	AccountTypeChecking AccountType = "checking"
	AccountTypeSaving   AccountType = "saving"
)

type Account struct {
	ID          uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AccountNumber string    `json:"account_number" gorm:"uniqueIndex;not null"`
	Type        AccountType `json:"type" gorm:"not null"`
	Balance     float64     `json:"balance" gorm:"not null;default:0"`
	Currency    string      `json:"currency" gorm:"default:'USD'"`
	IsActive    bool        `json:"is_active" gorm:"default:true"`
	UserID      uuid.UUID   `json:"user_id" gorm:"type:uuid;not null"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	User         User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Transactions []Transaction  `json:"transactions,omitempty" gorm:"foreignKey:AccountID"`
}

func (a *Account) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
} 