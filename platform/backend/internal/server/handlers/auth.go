package handlers

import (
	"net/http"

	"poker-platform/backend/internal/auth"
	"poker-platform/backend/internal/db"
	"poker-platform/backend/internal/models"

	"github.com/gin-gonic/gin"
)

// HandleRegister handles user registration
func HandleRegister(c *gin.Context, database *db.DB, authService *auth.Service) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	hash, err := authService.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	userID := auth.GenerateID()
	user := models.User{
		ID:           userID,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		Chips:        10000,
	}

	if err := database.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
		return
	}

	token, _ := authService.GenerateToken(userID)
	user.PasswordHash = ""

	c.JSON(http.StatusCreated, models.AuthResponse{Token: token, User: user})
}

// HandleLogin handles user login
func HandleLogin(c *gin.Context, database *db.DB, authService *auth.Service) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	if err := database.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !authService.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, _ := authService.GenerateToken(user.ID)
	user.PasswordHash = ""

	c.JSON(http.StatusOK, models.AuthResponse{Token: token, User: user})
}

// HandleGetCurrentUser returns the current authenticated user
func HandleGetCurrentUser(c *gin.Context, database *db.DB) {
	userID := c.GetString("user_id")

	var user models.User
	if err := database.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.PasswordHash = ""
	c.JSON(http.StatusOK, user)
}

// AuthMiddleware validates JWT tokens and sets user_id in context
func AuthMiddleware(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) < 8 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		token := authHeader[7:]
		userID, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
