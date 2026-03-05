package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	tigerbeetle "github.com/tigerbeetle/tigerbeetle-go"
	"gorm.io/gorm"

	"banking-system/src/models"
	"banking-system/src/utils"
)

type AuthController struct {
	DB *gorm.DB
	TB tigerbeetle.Client
}

type AuthInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name"`
}

// Register creates a user in Postgres and a ledger account in TigerBeetle
func (ctrl *AuthController) Register(c *gin.Context) {
	var input AuthInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, _ := utils.HashPassword(input.Password)

	user := models.Users{
		UserFullname: input.Name,
		UserEmail:    input.Email,
		UserPassword: hashedPassword,
		TBAccountID:  "[]",
	}

	if err := ctrl.DB.Create(&user).Error; err != nil {
		// This will tell you if it failed because of tb_account_id, email, or connection
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Could not create user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Registration successful. Now login to open an account."})
}

// Login verifies credentials and returns a JWT
func (ctrl *AuthController) Login(c *gin.Context) {
	var input AuthInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.Users
	if err := ctrl.DB.Where("user_email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !utils.CheckPasswordHash(input.Password, user.UserPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, _ := utils.GenerateToken(user.UUIDUser.String())
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user.UserFullname})
}

func (ctrl *AuthController) Logout(c *gin.Context) {
	// In a basic implementation with JWT, the logout is handled by the client
	// by deleting the token. Here we can register the event or clear cookies.

	c.JSON(http.StatusOK, gin.H{
		"message": "Session closed successfully",
	})
}

// ListUsers returns all registered users in PostgreSQL
func (ctrl *AuthController) ListUsers(c *gin.Context) {
	var users []models.Users

	// We query all records from the users table
	if err := ctrl.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch users"})
		return
	}

	// If the list is empty, we send an informational message
	if len(users) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No users found in database", "count": 0})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(users),
		"users": users,
	})
}
