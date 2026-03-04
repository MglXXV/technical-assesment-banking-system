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
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
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
	// En una implementación básica con JWT, el logout lo maneja el cliente
	// borrando el token. Aquí podemos registrar el evento o limpiar cookies.

	c.JSON(http.StatusOK, gin.H{
		"message": "Sesión cerrada exitosamente",
	})
}

// ListUsers devuelve todos los usuarios registrados en PostgreSQL
func (ctrl *AuthController) ListUsers(c *gin.Context) {
	var users []models.Users

	// Consultamos todos los registros de la tabla users
	if err := ctrl.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch users"})
		return
	}

	// Si la lista está vacía, enviamos un mensaje informativo
	if len(users) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No users found in database", "count": 0})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(users),
		"users": users,
	})
}
