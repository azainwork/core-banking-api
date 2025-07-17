package controllers

import (
	"net/http"
	"strconv"

	"github.com/azainwork/core-banking-api/services"
	"github.com/azainwork/core-banking-api/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TransactionController struct {
	transactionService *services.TransactionService
	accountService     *services.AccountService
}

func NewTransactionController(db *gorm.DB) *TransactionController {
	return &TransactionController{
		transactionService: services.NewTransactionService(db),
		accountService:     services.NewAccountService(db),
	}
}

type TransactionRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

type TransferRequest struct {
	ToAccountID string  `json:"to_account_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

func (c *TransactionController) Deposit(ctx *gin.Context) {
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

	var req TransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, err.Error())
		return
	}

	transaction, err := c.transactionService.ProcessDeposit(accountID, req.Amount, req.Description)
	if err != nil {
		utils.InternalServerError(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Deposit processed successfully", gin.H{
		"transaction": gin.H{
			"id":              transaction.ID,
			"transaction_id":  transaction.TransactionID,
			"type":            transaction.Type,
			"amount":          transaction.Amount,
			"currency":        transaction.Currency,
			"status":          transaction.Status,
			"description":     transaction.Description,
			"balance_before":  transaction.BalanceBefore,
			"balance_after":   transaction.BalanceAfter,
			"created_at":      transaction.CreatedAt,
		},
	})
}

func (c *TransactionController) Withdraw(ctx *gin.Context) {
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

	var req TransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, err.Error())
		return
	}

	transaction, err := c.transactionService.ProcessWithdrawal(accountID, req.Amount, req.Description)
	if err != nil {
		utils.InternalServerError(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Withdrawal processed successfully", gin.H{
		"transaction": gin.H{
			"id":              transaction.ID,
			"transaction_id":  transaction.TransactionID,
			"type":            transaction.Type,
			"amount":          transaction.Amount,
			"currency":        transaction.Currency,
			"status":          transaction.Status,
			"description":     transaction.Description,
			"balance_before":  transaction.BalanceBefore,
			"balance_after":   transaction.BalanceAfter,
			"created_at":      transaction.CreatedAt,
		},
	})
}

func (c *TransactionController) Transfer(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.UnauthorizedError(ctx, "User not authenticated")
		return
	}

	fromAccountID := ctx.Param("id")
	if fromAccountID == "" {
		utils.ValidationError(ctx, "Account ID is required")
		return
	}

	if err := c.accountService.ValidateAccountOwnership(fromAccountID, userID.(string)); err != nil {
		utils.NotFoundError(ctx, err.Error())
		return
	}

	var req TransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, err.Error())
		return
	}

	transaction, err := c.transactionService.ProcessTransfer(fromAccountID, req.ToAccountID, req.Amount, req.Description)
	if err != nil {
		utils.InternalServerError(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Transfer processed successfully", gin.H{
		"transaction": gin.H{
			"id":              transaction.ID,
			"transaction_id":  transaction.TransactionID,
			"type":            transaction.Type,
			"amount":          transaction.Amount,
			"currency":        transaction.Currency,
			"status":          transaction.Status,
			"description":     transaction.Description,
			"from_account_id": transaction.AccountID,
			"to_account_id":   transaction.ToAccountID,
			"balance_before":  transaction.BalanceBefore,
			"balance_after":   transaction.BalanceAfter,
			"created_at":      transaction.CreatedAt,
		},
	})
}

func (c *TransactionController) GetTransactions(ctx *gin.Context) {
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

	limitStr := ctx.DefaultQuery("limit", "50")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		utils.ValidationError(ctx, "Invalid limit parameter")
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		utils.ValidationError(ctx, "Invalid offset parameter")
		return
	}

	transactions, err := c.transactionService.GetTransactionsByAccountID(accountID, limit, offset)
	if err != nil {
		utils.InternalServerError(ctx, err.Error())
		return
	}

	var transactionList []gin.H
	for _, transaction := range transactions {
		transactionData := gin.H{
			"id":              transaction.ID,
			"transaction_id":  transaction.TransactionID,
			"type":            transaction.Type,
			"amount":          transaction.Amount,
			"currency":        transaction.Currency,
			"status":          transaction.Status,
			"description":     transaction.Description,
			"balance_before":  transaction.BalanceBefore,
			"balance_after":   transaction.BalanceAfter,
			"created_at":      transaction.CreatedAt,
		}

		if transaction.ToAccountID != nil {
			transactionData["to_account_id"] = transaction.ToAccountID
		}

		transactionList = append(transactionList, transactionData)
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Transactions retrieved successfully", gin.H{
		"transactions": transactionList,
		"count":        len(transactionList),
		"limit":        limit,
		"offset":       offset,
	})
}

func (c *TransactionController) GetTransaction(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.UnauthorizedError(ctx, "User not authenticated")
		return
	}

	transactionID := ctx.Param("id")
	if transactionID == "" {
		utils.ValidationError(ctx, "Transaction ID is required")
		return
	}

	transaction, err := c.transactionService.GetTransactionByID(transactionID)
	if err != nil {
		utils.NotFoundError(ctx, err.Error())
		return
	}

	if err := c.accountService.ValidateAccountOwnership(transaction.AccountID.String(), userID.(string)); err != nil {
		utils.NotFoundError(ctx, "Access denied")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Transaction retrieved successfully", gin.H{
		"transaction": gin.H{
			"id":              transaction.ID,
			"transaction_id":  transaction.TransactionID,
			"type":            transaction.Type,
			"amount":          transaction.Amount,
			"currency":        transaction.Currency,
			"status":          transaction.Status,
			"description":     transaction.Description,
			"account_id":      transaction.AccountID,
			"to_account_id":   transaction.ToAccountID,
			"balance_before":  transaction.BalanceBefore,
			"balance_after":   transaction.BalanceAfter,
			"created_at":      transaction.CreatedAt,
		},
	})
} 