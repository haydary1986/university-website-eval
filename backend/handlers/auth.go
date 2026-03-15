package handlers

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"website-eval-system/config"
	"website-eval-system/database"
	"website-eval-system/middleware"
	"website-eval-system/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxFailedFromIP = 20 // block IP after 20 failed attempts in 1 hour
)

func getMaxFailedAttempts() int {
	v := getSetting("max_login_attempts", "5")
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return 5
	}
	return n
}

func getBlockDurationMinutes() int {
	v := getSetting("block_duration_minutes", "30")
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return 30
	}
	return n
}

type AuthHandler struct {
	Config *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{Config: cfg}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	// Check if IP is blocked
	var blockedIP models.BlockedIP
	if err := database.DB.Where("ip_address = ? AND (expires_at IS NULL OR expires_at > ?)", ip, time.Now()).First(&blockedIP).Error; err == nil {
		h.logAttempt(req.Username, nil, ip, ua, false, "ip_blocked")
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "عنوان IP محظور. تواصل مع مدير النظام",
			"blocked": true,
		})
		return
	}

	// Check IP-based rate limiting (20 failed in last hour)
	var ipFailCount int64
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	database.DB.Model(&models.LoginAttempt{}).
		Where("ip_address = ? AND success = ? AND created_at > ?", ip, false, oneHourAgo).
		Count(&ipFailCount)
	if ipFailCount >= maxFailedFromIP {
		// Auto-block IP for 1 hour
		expires := time.Now().Add(1 * time.Hour)
		database.DB.Create(&models.BlockedIP{
			IPAddress: ip,
			Reason:    fmt.Sprintf("تجاوز الحد الأقصى للمحاولات الفاشلة (%d محاولة من نفس العنوان)", ipFailCount),
			ExpiresAt: &expires,
		})
		h.logAttempt(req.Username, nil, ip, ua, false, "ip_rate_limited")
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":   "محاولات كثيرة جداً. تم حظر العنوان مؤقتاً",
			"blocked": true,
		})
		return
	}

	// Find user
	var user models.User
	if err := database.DB.Preload("University").Where("username = ?", req.Username).First(&user).Error; err != nil {
		h.logAttempt(req.Username, nil, ip, ua, false, "user_not_found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "اسم المستخدم أو كلمة المرور غير صحيحة"})
		return
	}

	// Check if account is blocked
	if user.IsBlocked {
		if user.BlockedUntil != nil && time.Now().After(*user.BlockedUntil) {
			// Block expired, auto-unblock
			database.DB.Model(&user).Updates(map[string]interface{}{
				"is_blocked":      false,
				"blocked_until":   nil,
				"failed_attempts": 0,
			})
		} else {
			h.logAttempt(req.Username, &user.ID, ip, ua, false, "account_blocked")
			remaining := ""
			if user.BlockedUntil != nil {
				mins := int(time.Until(*user.BlockedUntil).Minutes()) + 1
				remaining = fmt.Sprintf(" (يتم رفع الحظر بعد %d دقيقة)", mins)
			}
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "الحساب محظور بسبب محاولات دخول فاشلة متعددة" + remaining,
				"blocked": true,
			})
			return
		}
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		now := time.Now()
		newFailCount := user.FailedAttempts + 1

		updates := map[string]interface{}{
			"failed_attempts": newFailCount,
			"last_failed_at":  now,
		}

		// Block after max failed attempts
		maxAttempts := getMaxFailedAttempts()
		blockMins := getBlockDurationMinutes()
		if newFailCount >= maxAttempts {
			blockUntil := now.Add(time.Duration(blockMins) * time.Minute)
			updates["is_blocked"] = true
			updates["blocked_until"] = blockUntil

			h.logAttempt(req.Username, &user.ID, ip, ua, false, "account_auto_blocked")
			database.DB.Create(&models.AuditLog{
				UserID:    user.ID,
				Action:    "account_blocked",
				IPAddress: ip,
				UserAgent: ua,
				Details:   fmt.Sprintf("تم حظر الحساب تلقائياً بعد %d محاولة فاشلة. الحظر حتى %s", newFailCount, blockUntil.Format("15:04")),
			})

			database.DB.Model(&user).Updates(updates)
			c.JSON(http.StatusForbidden, gin.H{
				"error":              fmt.Sprintf("تم حظر الحساب لمدة %d دقيقة بسبب %d محاولات فاشلة", blockMins, maxAttempts),
				"blocked":            true,
				"remaining_attempts": 0,
			})
			return
		}

		database.DB.Model(&user).Updates(updates)
		h.logAttempt(req.Username, &user.ID, ip, ua, false, "invalid_password")

		database.DB.Create(&models.AuditLog{
			UserID:    user.ID,
			Action:    "login_failed",
			IPAddress: ip,
			UserAgent: ua,
			Details:   fmt.Sprintf("محاولة دخول فاشلة (%d من %d)", newFailCount, maxAttempts),
		})

		remaining := maxAttempts - newFailCount
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":              "اسم المستخدم أو كلمة المرور غير صحيحة",
			"remaining_attempts": remaining,
		})
		return
	}

	// Successful login - reset failed attempts
	now := time.Now()
	database.DB.Model(&user).Updates(map[string]interface{}{
		"failed_attempts": 0,
		"last_failed_at":  nil,
		"last_login_at":   now,
		"last_login_ip":   ip,
	})

	h.logAttempt(req.Username, &user.ID, ip, ua, true, "success")
	database.DB.Create(&models.AuditLog{
		UserID:    user.ID,
		Action:    "login",
		IPAddress: ip,
		UserAgent: ua,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في إنشاء رمز المصادقة"})
		return
	}

	// Track active session
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(tokenStr)))
	database.DB.Create(&models.ActiveSession{
		UserID:    user.ID,
		TokenHash: tokenHash,
		IPAddress: ip,
		UserAgent: ua,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

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
		c.JSON(http.StatusNotFound, gin.H{"error": "المستخدم غير موجود"})
		return
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "كلمة المرور الحالية غير صحيحة"})
		return
	}

	// Ensure new password != old password
	if req.NewPassword == req.OldPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "كلمة المرور الجديدة يجب أن تكون مختلفة عن الحالية"})
		return
	}

	// Password strength validation
	if len(req.NewPassword) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "كلمة المرور يجب أن تكون 8 أحرف على الأقل"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في تشفير كلمة المرور"})
		return
	}

	now := time.Now()
	database.DB.Model(&user).Updates(map[string]interface{}{
		"password":              string(hash),
		"must_change_password":  false,
		"password_changed_at":   now,
	})

	database.DB.Create(&models.AuditLog{
		UserID:    user.ID,
		Action:    "password_change",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Details:   "تغيير كلمة المرور بنجاح",
	})

	c.JSON(http.StatusOK, gin.H{"message": "تم تغيير كلمة المرور بنجاح"})
}

func (h *AuthHandler) Register(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "super_admin" && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "صلاحيات غير كافية"})
		return
	}

	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	if (req.Role == "admin" || req.Role == "super_admin") && role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "فقط المدير العام يمكنه إنشاء حسابات المراجعين"})
		return
	}

	if req.Role != "super_admin" && req.Role != "admin" && req.Role != "university" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "دور غير صالح"})
		return
	}

	var count int64
	database.DB.Model(&models.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "اسم المستخدم موجود مسبقاً"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في تشفير كلمة المرور"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في إنشاء المستخدم"})
		return
	}

	database.DB.Preload("University").First(&user, user.ID)

	adminID, _ := c.Get("user_id")
	database.DB.Create(&models.AuditLog{
		UserID:    adminID.(uint),
		Action:    "user_created",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Details:   fmt.Sprintf("إنشاء مستخدم جديد: %s (الدور: %s)", req.Username, req.Role),
	})

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, _ := c.Get("user")
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	if len(tokenStr) > 7 {
		tokenStr = tokenStr[7:] // Remove "Bearer "
		tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(tokenStr)))
		database.DB.Where("token_hash = ?", tokenHash).Delete(&models.ActiveSession{})
	}

	userID, _ := c.Get("user_id")
	database.DB.Create(&models.AuditLog{
		UserID:    userID.(uint),
		Action:    "logout",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Details:   "تسجيل خروج",
	})

	c.JSON(http.StatusOK, gin.H{"message": "تم تسجيل الخروج بنجاح"})
}

func (h *AuthHandler) GetActiveSessions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var sessions []models.ActiveSession
	database.DB.Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").Find(&sessions)
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func (h *AuthHandler) logAttempt(username string, userID *uint, ip, ua string, success bool, reason string) {
	database.DB.Create(&models.LoginAttempt{
		Username:  username,
		UserID:    userID,
		IPAddress: ip,
		UserAgent: ua,
		Success:   success,
		Reason:    reason,
	})
}
