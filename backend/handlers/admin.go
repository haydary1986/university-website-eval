package handlers

import (
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
