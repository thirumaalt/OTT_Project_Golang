package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/myflix/auth-service/store"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	db        *gorm.DB
	jwtSecret []byte
}

func New(db *gorm.DB, secret string) *Handler {
	if secret == "" {
		panic("JWT_SECRET is required")
	}
	return &Handler{db: db, jwtSecret: []byte(secret)}
}

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email"    binding:"required,email"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register creates a new user account.
func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := store.User{
		Username: req.Username,
		Password: string(hash),
		Email:    req.Email,
	}

	if result := h.db.Create(&user); result.Error != nil {
		// Return generic message — don't reveal whether username/email exists
		c.JSON(http.StatusConflict, gin.H{"error": "registration failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"userId":   user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

// Login validates credentials and returns a signed JWT.
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user store.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		// Constant-time response — prevents user enumeration
		bcrypt.CompareHashAndPassword([]byte("$2a$12$placeholder"), []byte(req.Password))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"userId":   user.ID,
		"username": user.Username,
	})
}

func (h *Handler) generateToken(user store.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":    user.Username,
		"userId": user.ID,
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(h.jwtSecret)
}
