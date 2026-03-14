package handlers

import (
	"fmt"
	"net/http"
	"time"

	"website-eval-system/database"
	"website-eval-system/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

// === User Management ===

func (h *AdminHandler) ListUsers(c *gin.Context) {
	var users []models.User
	query := database.DB.Preload("University")

	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}

	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if req.Role != "super_admin" && req.Role != "admin" && req.Role != "university" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

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

	database.DB.Preload("University").First(&user, user.ID)
	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.FullName != "" {
		updates["full_name"] = req.FullName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.UniversityID != nil {
		updates["university_id"] = req.UniversityID
	}
	if req.AssignedCategories != nil {
		updates["assigned_categories"] = req.AssignedCategories
	}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		updates["password"] = string(hash)
	}

	database.DB.Model(&user).Updates(updates)
	database.DB.Preload("University").First(&user, user.ID)

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Prevent deleting yourself
	currentUserID, _ := c.Get("user_id")
	if user.ID == currentUserID.(uint) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete yourself"})
		return
	}

	database.DB.Delete(&user)
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *AdminHandler) AssignCategories(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only assign categories to admin users"})
		return
	}

	var req models.AssignCategoriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	user.AssignedCategories = req.CategoryIDs
	database.DB.Save(&user)

	database.DB.Preload("University").First(&user, user.ID)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// === Audit Logs ===

func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	var logs []models.AuditLog
	query := database.DB.Preload("User").Order("created_at DESC")

	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
	}
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	page := 1
	pageSize := 50
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	var total int64
	query.Model(&models.AuditLog{}).Count(&total)

	offset := (page - 1) * pageSize
	query.Offset(offset).Limit(pageSize).Find(&logs)

	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// === Security: Login Attempts ===

func (h *AdminHandler) ListLoginAttempts(c *gin.Context) {
	var attempts []models.LoginAttempt
	query := database.DB.Preload("User").Order("created_at DESC")

	if success := c.Query("success"); success != "" {
		query = query.Where("success = ?", success == "true")
	}
	if ip := c.Query("ip"); ip != "" {
		query = query.Where("ip_address = ?", ip)
	}
	if username := c.Query("username"); username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	page := 1
	pageSize := 50
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}

	var total int64
	query.Model(&models.LoginAttempt{}).Count(&total)

	offset := (page - 1) * pageSize
	query.Offset(offset).Limit(pageSize).Find(&attempts)

	c.JSON(http.StatusOK, gin.H{
		"attempts": attempts,
		"total":    total,
		"page":     page,
	})
}

// === Security: Block/Unblock Users ===

func (h *AdminHandler) UnblockUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "المستخدم غير موجود"})
		return
	}

	database.DB.Model(&user).Updates(map[string]interface{}{
		"is_blocked":      false,
		"blocked_until":   nil,
		"failed_attempts": 0,
		"last_failed_at":  nil,
	})

	adminID, _ := c.Get("user_id")
	database.DB.Create(&models.AuditLog{
		UserID:    adminID.(uint),
		Action:    "account_unblocked",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Details:   fmt.Sprintf("إلغاء حظر المستخدم: %s (ID: %s)", user.Username, id),
	})

	c.JSON(http.StatusOK, gin.H{"message": "تم إلغاء حظر المستخدم بنجاح", "user": user})
}

func (h *AdminHandler) BlockUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "المستخدم غير موجود"})
		return
	}

	// Prevent blocking yourself
	currentUserID, _ := c.Get("user_id")
	if user.ID == currentUserID.(uint) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "لا يمكنك حظر حسابك"})
		return
	}

	var body struct {
		Duration int    `json:"duration"` // minutes, 0 = permanent
		Reason   string `json:"reason"`
	}
	c.ShouldBindJSON(&body)

	updates := map[string]interface{}{
		"is_blocked":      true,
		"failed_attempts": 0,
	}
	if body.Duration > 0 {
		blockUntil := time.Now().Add(time.Duration(body.Duration) * time.Minute)
		updates["blocked_until"] = blockUntil
	}

	database.DB.Model(&user).Updates(updates)

	adminID, _ := c.Get("user_id")
	database.DB.Create(&models.AuditLog{
		UserID:    adminID.(uint),
		Action:    "account_blocked_manual",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Details:   fmt.Sprintf("حظر المستخدم يدوياً: %s - السبب: %s", user.Username, body.Reason),
	})

	c.JSON(http.StatusOK, gin.H{"message": "تم حظر المستخدم بنجاح"})
}

// === Security: IP Management ===

func (h *AdminHandler) ListBlockedIPs(c *gin.Context) {
	var ips []models.BlockedIP
	database.DB.Preload("Admin").Order("created_at DESC").Find(&ips)
	c.JSON(http.StatusOK, gin.H{"blocked_ips": ips})
}

func (h *AdminHandler) BlockIP(c *gin.Context) {
	var body struct {
		IPAddress string `json:"ip_address" binding:"required"`
		Reason    string `json:"reason"`
		Duration  int    `json:"duration"` // hours, 0 = permanent
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "عنوان IP مطلوب"})
		return
	}

	adminID, _ := c.Get("user_id")
	blockedIP := models.BlockedIP{
		IPAddress: body.IPAddress,
		Reason:    body.Reason,
		BlockedBy: adminID.(uint),
	}
	if body.Duration > 0 {
		expires := time.Now().Add(time.Duration(body.Duration) * time.Hour)
		blockedIP.ExpiresAt = &expires
	}

	database.DB.Where("ip_address = ?", body.IPAddress).Delete(&models.BlockedIP{})
	database.DB.Create(&blockedIP)

	database.DB.Create(&models.AuditLog{
		UserID:    adminID.(uint),
		Action:    "ip_blocked",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Details:   fmt.Sprintf("حظر عنوان IP: %s - السبب: %s", body.IPAddress, body.Reason),
	})

	c.JSON(http.StatusOK, gin.H{"message": "تم حظر العنوان بنجاح"})
}

func (h *AdminHandler) UnblockIP(c *gin.Context) {
	ip := c.Param("ip")
	database.DB.Where("ip_address = ?", ip).Delete(&models.BlockedIP{})

	adminID, _ := c.Get("user_id")
	database.DB.Create(&models.AuditLog{
		UserID:    adminID.(uint),
		Action:    "ip_unblocked",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Details:   fmt.Sprintf("إلغاء حظر عنوان IP: %s", ip),
	})

	c.JSON(http.StatusOK, gin.H{"message": "تم إلغاء حظر العنوان بنجاح"})
}

// === Security: Active Sessions ===

func (h *AdminHandler) ListAllSessions(c *gin.Context) {
	var sessions []models.ActiveSession
	database.DB.Preload("User").Where("expires_at > ?", time.Now()).
		Order("created_at DESC").Find(&sessions)
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func (h *AdminHandler) TerminateSession(c *gin.Context) {
	id := c.Param("id")
	database.DB.Delete(&models.ActiveSession{}, id)

	adminID, _ := c.Get("user_id")
	database.DB.Create(&models.AuditLog{
		UserID:    adminID.(uint),
		Action:    "session_terminated",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Details:   fmt.Sprintf("إنهاء جلسة رقم: %s", id),
	})

	c.JSON(http.StatusOK, gin.H{"message": "تم إنهاء الجلسة بنجاح"})
}

func (h *AdminHandler) TerminateUserSessions(c *gin.Context) {
	id := c.Param("id")
	database.DB.Where("user_id = ?", id).Delete(&models.ActiveSession{})

	adminID, _ := c.Get("user_id")
	database.DB.Create(&models.AuditLog{
		UserID:    adminID.(uint),
		Action:    "all_sessions_terminated",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Details:   fmt.Sprintf("إنهاء جميع جلسات المستخدم ID: %s", id),
	})

	c.JSON(http.StatusOK, gin.H{"message": "تم إنهاء جميع الجلسات بنجاح"})
}

// === Security: Dashboard Stats ===

func (h *AdminHandler) SecurityOverview(c *gin.Context) {
	now := time.Now()
	last24h := now.Add(-24 * time.Hour)
	lastWeek := now.Add(-7 * 24 * time.Hour)

	var totalUsers int64
	database.DB.Model(&models.User{}).Count(&totalUsers)

	var blockedUsers int64
	database.DB.Model(&models.User{}).Where("is_blocked = ?", true).Count(&blockedUsers)

	var blockedIPs int64
	database.DB.Model(&models.BlockedIP{}).Where("expires_at IS NULL OR expires_at > ?", now).Count(&blockedIPs)

	var activeSessions int64
	database.DB.Model(&models.ActiveSession{}).Where("expires_at > ?", now).Count(&activeSessions)

	var loginSuccess24h int64
	database.DB.Model(&models.LoginAttempt{}).Where("success = ? AND created_at > ?", true, last24h).Count(&loginSuccess24h)

	var loginFailed24h int64
	database.DB.Model(&models.LoginAttempt{}).Where("success = ? AND created_at > ?", false, last24h).Count(&loginFailed24h)

	var loginFailedWeek int64
	database.DB.Model(&models.LoginAttempt{}).Where("success = ? AND created_at > ?", false, lastWeek).Count(&loginFailedWeek)

	var passwordChanges24h int64
	database.DB.Model(&models.AuditLog{}).Where("action = ? AND created_at > ?", "password_change", last24h).Count(&passwordChanges24h)

	// Top failed IPs
	type IPStat struct {
		IPAddress string `json:"ip_address"`
		Count     int64  `json:"count"`
	}
	var topFailedIPs []IPStat
	database.DB.Model(&models.LoginAttempt{}).
		Select("ip_address, COUNT(*) as count").
		Where("success = ? AND created_at > ?", false, lastWeek).
		Group("ip_address").
		Order("count DESC").
		Limit(10).
		Find(&topFailedIPs)

	// Top failed usernames
	type UsernameStat struct {
		Username string `json:"username"`
		Count    int64  `json:"count"`
	}
	var topFailedUsers []UsernameStat
	database.DB.Model(&models.LoginAttempt{}).
		Select("username, COUNT(*) as count").
		Where("success = ? AND created_at > ?", false, lastWeek).
		Group("username").
		Order("count DESC").
		Limit(10).
		Find(&topFailedUsers)

	// Recent blocked accounts
	var recentBlocked []models.User
	database.DB.Where("is_blocked = ?", true).
		Order("updated_at DESC").
		Limit(10).
		Find(&recentBlocked)

	c.JSON(http.StatusOK, gin.H{
		"total_users":         totalUsers,
		"blocked_users":       blockedUsers,
		"blocked_ips":         blockedIPs,
		"active_sessions":     activeSessions,
		"login_success_24h":   loginSuccess24h,
		"login_failed_24h":    loginFailed24h,
		"login_failed_week":   loginFailedWeek,
		"password_changes_24h": passwordChanges24h,
		"top_failed_ips":      topFailedIPs,
		"top_failed_users":    topFailedUsers,
		"recent_blocked":      recentBlocked,
	})
}

// === Submission Review ===

func (h *AdminHandler) ListSubmissions(c *gin.Context) {
	var submissions []models.Submission
	query := database.DB.Preload("University").Preload("AcademicYear")

	// Admin users see submissions based on status (not draft)
	query = query.Where("status != ?", "draft")

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if uid := c.Query("university_id"); uid != "" {
		query = query.Where("university_id = ?", uid)
	}
	if yearID := c.Query("academic_year_id"); yearID != "" {
		query = query.Where("academic_year_id = ?", yearID)
	}

	if err := query.Order("submitted_at DESC").Find(&submissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"submissions": submissions})
}

func (h *AdminHandler) GetSubmission(c *gin.Context) {
	id := c.Param("id")

	var submission models.Submission
	if err := database.DB.
		Preload("University").
		Preload("AcademicYear").
		Preload("Items.Criteria").
		Preload("Items.Criteria.Category").
		Preload("Reviews.Admin").
		Preload("Reviews.Category").
		First(&submission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"submission": submission})
}

func (h *AdminHandler) ReviewSubmission(c *gin.Context) {
	id := c.Param("id")

	var submission models.Submission
	if err := database.DB.First(&submission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	if submission.Status != "submitted" && submission.Status != "under_review" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Submission is not available for review"})
		return
	}

	var req models.ReviewInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	adminID, _ := c.Get("user_id")

	// Update submission items with scores
	for _, item := range req.Items {
		updates := map[string]interface{}{
			"score":         item.Score,
			"admin_comment": item.AdminComment,
		}
		if item.Status != "" {
			updates["status"] = item.Status
		}
		database.DB.Model(&models.SubmissionItem{}).Where("id = ?", item.SubmissionItemID).Updates(updates)
	}

	// Create or update review record
	var review models.Review
	result := database.DB.Where("submission_id = ? AND admin_id = ? AND category_id = ?",
		submission.ID, adminID, req.CategoryID).First(&review)

	if result.Error != nil {
		review = models.Review{
			SubmissionID: submission.ID,
			AdminID:      adminID.(uint),
			CategoryID:   req.CategoryID,
			Status:       "reviewed",
			Comments:     req.Comments,
		}
		database.DB.Create(&review)
	} else {
		review.Status = "reviewed"
		review.Comments = req.Comments
		database.DB.Save(&review)
	}

	// Update submission status to under_review
	if submission.Status == "submitted" {
		database.DB.Model(&submission).Update("status", "under_review")
	}

	// Recalculate total score
	var totalScore float64
	database.DB.Model(&models.SubmissionItem{}).
		Where("submission_id = ?", submission.ID).
		Select("COALESCE(SUM(score), 0)").
		Scan(&totalScore)
	database.DB.Model(&submission).Update("total_score", totalScore)

	c.JSON(http.StatusOK, gin.H{"message": "Review submitted successfully", "total_score": totalScore})
}

func (h *AdminHandler) ApproveSubmission(c *gin.Context) {
	id := c.Param("id")

	var submission models.Submission
	if err := database.DB.First(&submission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	now := time.Now()
	submission.Status = "approved"
	submission.ReviewedAt = &now

	// Recalculate total score
	var totalScore float64
	database.DB.Model(&models.SubmissionItem{}).
		Where("submission_id = ?", submission.ID).
		Select("COALESCE(SUM(score), 0)").
		Scan(&totalScore)
	submission.TotalScore = totalScore

	database.DB.Save(&submission)

	c.JSON(http.StatusOK, gin.H{"submission": submission, "message": "Submission approved"})
}

func (h *AdminHandler) RejectSubmission(c *gin.Context) {
	id := c.Param("id")

	var submission models.Submission
	if err := database.DB.First(&submission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	var body struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&body)

	now := time.Now()
	submission.Status = "rejected"
	submission.ReviewedAt = &now

	database.DB.Save(&submission)

	c.JSON(http.StatusOK, gin.H{"submission": submission, "message": "Submission rejected", "reason": body.Reason})
}
