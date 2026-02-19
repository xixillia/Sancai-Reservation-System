package controllers

import (
	"database/sql"
	"net/http"

	"Sancai/auth"
	"Sancai/structs"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context, db *sql.DB) {
	var user structs.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//email format
	if !isValidEmail(user.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	//unique email
	var existingID int
	err := db.QueryRow("SELECT id FROM users WHERE email = $1", user.Email).Scan(&existingID)
	if err != sql.ErrNoRows {
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// hash password BEFORE save (bcrypt)
	hashed, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	user.Password = string(hashed)

	query := "INSERT INTO users (name, email, password, role, is_active) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err = db.QueryRow(query, user.Name, user.Email, user.Password, user.Role, user.IsActive).Scan(&user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user registered"})
}

// Login user
// @Summary Login user
// @Description Login pakai email dan password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body structs.LoginInput true "Login input"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /login [post]
func Login(c *gin.Context, db *sql.DB) {
	var user structs.User
	var input structs.LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	query := "SELECT id, email, password, role, is_active FROM users WHERE email = $1"
	err := db.QueryRow(query, input.Email).Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.IsActive)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.Role, user.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}
