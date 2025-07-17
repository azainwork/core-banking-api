package controllers

import (
	"net/http"
	_"strconv"

	"github.com/azainwork/core-banking-api/models"
	"github.com/azainwork/core-banking-api/services"
	"github.com/azainwork/core-banking-api/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AccountController struct {
	accountService *services.AccountService
}

func NewAccountController(db *gorm.DB) *AccountController {
	return &AccountController{
		accountService: services.NewAccountService(db),
	}
}

type CreateAccountRequest struct {
	Type           string  `json:"type" binding:"required,oneof=checking saving"`
	InitialBalance float64 `json:"initial_balance" binding:"gte=0"`
}

func (c *AccountController) CreateAccount(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.UnauthorizedError(ctx, "User not authenticated")
		return
	}

	var req CreateAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, err.Error())
		return
	}

	accountType := models.AccountType(req.Type)

	account, err := c.accountService.CreateAccount(userID.(string), accountType, req.InitialBalance)
	if err != nil {
		utils.InternalServerError(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, http.StatusCreated, "Account created successfully", gin.H{
		"account": gin.H{
			"id":             account.ID,
			"account_number": account.AccountNumber,
			"type":           account.Type,
			"balance":        account.Balance,
			"currency":       account.Currency,
			"created_at":     account.CreatedAt,
		},
	})
}

func (c *AccountController) GetAccounts(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.UnauthorizedError(ctx, "User not authenticated")
		return
	}

	accounts, err := c.accountService.GetAccountsByUserID(userID.(string))
	if err != nil {
		utils.InternalServerError(ctx, err.Error())
		return
	}

	var accountList []gin.H
	for _, account := range accounts {
		accountList = append(accountList, gin.H{
			"id":             account.ID,
			"account_number": account.AccountNumber,
			"type":           account.Type,
			"balance":        account.Balance,
			"currency":       account.Currency,
			"is_active":      account.IsActive,
			"created_at":     account.CreatedAt,
		})
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Accounts retrieved successfully", gin.H{
		"accounts": accountList,
		"count":    len(accountList),
	})
}

func (c *AccountController) GetAccount(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.UnauthorizedError(ctx, "User not authenticated")
		return
	}

	accountID := ctx.Param("id")
	if accountID == "" {
		utils.ValidationError(ctx, "Account ID is required")
		return
	}

	if err := c.accountService.ValidateAccountOwnership(accountID, userID.(string)); err != nil {
		utils.NotFoundError(ctx, err.Error())
		return
	}

	account, err := c.accountService.GetAccountByID(accountID)
	if err != nil {
		utils.NotFoundError(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Account retrieved successfully", gin.H{
		"account": gin.H{
			"id":             account.ID,
			"account_number": account.AccountNumber,
			"type":           account.Type,
			"balance":        account.Balance,
			"currency":       account.Currency,
			"is_active":      account.IsActive,
			"created_at":     account.CreatedAt,
			"updated_at":     account.UpdatedAt,
		},
	})
}

func (c *AccountController) GetAccountBalance(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.UnauthorizedError(ctx, "User not authenticated")
		return
	}

	accountID := ctx.Param("id")
	if accountID == "" {
		utils.ValidationError(ctx, "Account ID is required")
		return
	}

	if err := c.accountService.ValidateAccountOwnership(accountID, userID.(string)); err != nil {
		utils.NotFoundError(ctx, err.Error())
		return
	}

	account, err := c.accountService.GetAccountByID(accountID)
	if err != nil {
		utils.NotFoundError(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Account balance retrieved successfully", gin.H{
		"account_id": account.ID,
		"balance":    account.Balance,
		"currency":   account.Currency,
	})
} 