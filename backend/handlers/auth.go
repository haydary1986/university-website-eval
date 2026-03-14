package handlers

import (
	"net/http"
	"time"

	"website-eval-system/config"
	"website-eval-system/database"
	"website-eval-system/middleware"
	"website-eval-system/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Config *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{Config: cfg}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Preload("University").Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Log login action
	database.DB.Create(&models.AuditLog{
		UserID:    user.ID,
		Action:    "login",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Details:   "تسجيل دخول ناجح",
	})

	// Generate JWT
	claims := &middleware.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(h.Config.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":                tokenStr,
		"user":                 user,
		"must_change_password": user.MustChangePassword,
	})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "كلمة المرور الجديدة يجب أن تكون 8 أحرف على الأقل"})
		return
	}

	userID, _ := c.Get("user_id")
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "كلمة المرور الحالية غير صحيحة"})
		return
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password and clear must_change_password flag
	database.DB.Model(&user).Updates(map[string]interface{}{
		"password":             string(hash),
		"must_change_password": false,
	})

	// Log password change
	database.DB.Create(&models.AuditLog{
		UserID:    user.ID,
		Action:    "password_change",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Details:   "تغيير كلمة المرور",
	})

	c.JSON(http.StatusOK, gin.H{"message": "تم تغيير كلمة المرور بنجاح"})
}

func (h *AuthHandler) Register(c *gin.Context) {
	// Only super_admin and admin can create users
	role, _ := c.Get("role")
	if role != "super_admin" && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Only super_admin can create admin or super_admin users
	if (req.Role == "admin" || req.Role == "super_admin") && role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only super admin can create admin users"})
		return
	}

	// Validate role
	if req.Role != "super_admin" && req.Role != "admin" && req.Role != "university" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	// Check if username exists
	var count int64
	database.DB.Model(&models.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Username:           req.Username,
		Password:           string(hash),
		FullName:           req.FullName,
		Email:              req.Email,
		Phone:              req.Phone,
		Role:               req.Role,
		UniversityID:       req.UniversityID,
		AssignedCategories: req.AssignedCategories,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Reload with university
	database.DB.Preload("University").First(&user, user.ID)

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, _ := c.Get("user")
	c.JSON(http.StatusOK, gin.H{"user": user})
}
