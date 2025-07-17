package controllers

import (
	"net/http"

	"github.com/azainwork/core-banking-api/models"
	"github.com/azainwork/core-banking-api/services"
	"github.com/azainwork/core-banking-api/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	userService *services.UserService
}

func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{
		userService: services.NewUserService(db),
	}
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (c *AuthController) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, err.Error())
		return
	}

	user := &models.User{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		IsActive:  true,
	}

	if err := c.userService.RegisterUser(user); err != nil {
		utils.ConflictError(ctx, err.Error())
		return
	}

	token, err := utils.GenerateJWTToken(user.ID.String(), user.Email)
	if err != nil {
		utils.InternalServerError(ctx, "Failed to generate token")
		return
	}

	utils.SuccessResponse(ctx, http.StatusCreated, "User registered successfully", gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"phone":      user.Phone,
		},
		"token": token,
	})
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, err.Error())
		return
	}

	token, user, err := c.userService.LoginUser(req.Email, req.Password)
	if err != nil {
		utils.UnauthorizedError(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Login successful", gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"phone":      user.Phone,
		},
		"token": token,
	})
}

func (c *AuthController) GetProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		utils.UnauthorizedError(ctx, "User not authenticated")
		return
	}

	user, err := c.userService.GetUserByID(userID.(string))
	if err != nil {
		utils.NotFoundError(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Profile retrieved successfully", gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"phone":      user.Phone,
			"created_at": user.CreatedAt,
		},
	})
} 