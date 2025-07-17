package routes

import (
	"github.com/azainwork/core-banking-api/controllers"
	"github.com/azainwork/core-banking-api/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	authController := controllers.NewAuthController(db)
	accountController := controllers.NewAccountController(db)
	transactionController := controllers.NewTransactionController(db)

	api := router.Group("/api/v1")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Core Banking API is running",
		})
	})

	public := api.Group("/auth")
	{
		public.POST("/register", authController.Register)
		public.POST("/login", authController.Login)
	}

	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", authController.GetProfile)

		accounts := protected.Group("/accounts")
		{
			accounts.POST("/", accountController.CreateAccount)
			accounts.GET("/", accountController.GetAccounts)
			accounts.GET("/:id", accountController.GetAccount)
			accounts.GET("/:id/balance", accountController.GetAccountBalance)
		}

		transactions := protected.Group("/accounts/:id/transactions")
		{
			transactions.POST("/deposit", transactionController.Deposit)
			transactions.POST("/withdraw", transactionController.Withdraw)
			transactions.POST("/transfer", transactionController.Transfer)
			transactions.GET("/", transactionController.GetTransactions)
		}

		protected.GET("/transactions/:id", transactionController.GetTransaction)
	}
} 