package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"website-eval-system/config"
	"website-eval-system/database"
	"website-eval-system/models"

	"github.com/gin-gonic/gin"
)

type SubmissionHandler struct {
	Config *config.Config
}

func NewSubmissionHandler(cfg *config.Config) *SubmissionHandler {
	return &SubmissionHandler{Config: cfg}
}

func (h *SubmissionHandler) List(c *gin.Context) {
	var submissions []models.Submission
	query := database.DB.Preload("University").Preload("AcademicYear")

	role, _ := c.Get("role")
	user, _ := c.Get("user")
	u := user.(models.User)

	// University users can only see their own submissions
	if role == "university" && u.UniversityID != nil {
		query = query.Where("university_id = ?", *u.UniversityID)
	}

	// Filters
	if uid := c.Query("university_id"); uid != "" {
		query = query.Where("university_id = ?", uid)
	}
	if yearID := c.Query("academic_year_id"); yearID != "" {
		query = query.Where("academic_year_id = ?", yearID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&submissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions"})
		return
	}

	// Flatten for frontend
	type FlatSubmission struct {
		models.Submission
		UniversityName   string `json:"university_name"`
		UniversityType   string `json:"university_type"`
		AcademicYearName string `json:"academic_year_name"`
	}

	var flat []FlatSubmission
	for _, s := range submissions {
		fs := FlatSubmission{Submission: s}
		if s.University != nil {
			fs.UniversityName = s.University.Name
			fs.UniversityType = s.University.Type
		}
		if s.AcademicYear != nil {
			fs.AcademicYearName = s.AcademicYear.Name
		}
		flat = append(flat, fs)
	}

	c.JSON(http.StatusOK, gin.H{"submissions": flat})
}

func (h *SubmissionHandler) Get(c *gin.Context) {
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

	// Check access
	role, _ := c.Get("role")
	user, _ := c.Get("user")
	u := user.(models.User)

	if role == "university" && u.UniversityID != nil && *u.UniversityID != submission.UniversityID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"submission": submission})
}

func (h *SubmissionHandler) Create(c *gin.Context) {
	// Check if submissions are open
	if !GetSubmissionsOpen() {
		c.JSON(http.StatusForbidden, gin.H{"error": "عملية التقديم مغلقة حالياً"})
		return
	}

	user, _ := c.Get("user")
	u := user.(models.User)

	if u.UniversityID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not associated with a university"})
		return
	}

	var req models.CreateSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get the latest version for this university + academic year
	var latestVersion int
	database.DB.Model(&models.Submission{}).
		Where("university_id = ? AND academic_year_id = ?", *u.UniversityID, req.AcademicYearID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&latestVersion)

	submission := models.Submission{
		UniversityID:     *u.UniversityID,
		AcademicYearID:   req.AcademicYearID,
		Version:          latestVersion + 1,
		Status:           "draft",
		AuthorizedPerson: req.AuthorizedPerson,
		AuthorizedPhone:  req.AuthorizedPhone,
		AuthorizedEmail:  req.AuthorizedEmail,
	}

	if err := database.DB.Create(&submission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create submission"})
		return
	}

	// Pre-populate submission items for all criteria
	var allCriteria []models.Criteria
	database.DB.Find(&allCriteria)

	for _, criteria := range allCriteria {
		item := models.SubmissionItem{
			SubmissionID: submission.ID,
			CriteriaID:   criteria.ID,
			Status:       "pending",
		}
		database.DB.Create(&item)
	}

	// Reload with associations
	database.DB.Preload("University").Preload("AcademicYear").Preload("Items.Criteria").First(&submission, submission.ID)

	c.JSON(http.StatusCreated, gin.H{"submission": submission})
}

func (h *SubmissionHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var submission models.Submission
	if err := database.DB.First(&submission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	if submission.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only update draft submissions"})
		return
	}

	// Check ownership
	user, _ := c.Get("user")
	u := user.(models.User)
	if u.UniversityID == nil || *u.UniversityID != submission.UniversityID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req models.UpdateSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Update submission fields
	updates := map[string]interface{}{}
	if req.AuthorizedPerson != "" {
		updates["authorized_person"] = req.AuthorizedPerson
	}
	if req.AuthorizedPhone != "" {
		updates["authorized_phone"] = req.AuthorizedPhone
	}
	if req.AuthorizedEmail != "" {
		updates["authorized_email"] = req.AuthorizedEmail
	}

	if len(updates) > 0 {
		database.DB.Model(&submission).Updates(updates)
	}

	// Update items
	for _, item := range req.Items {
		database.DB.Model(&models.SubmissionItem{}).
			Where("submission_id = ? AND criteria_id = ?", submission.ID, item.CriteriaID).
			Updates(map[string]interface{}{
				"evidence":      item.Evidence,
				"evidence_file": item.EvidenceFile,
			})
	}

	// Reload
	database.DB.Preload("University").Preload("AcademicYear").Preload("Items.Criteria").First(&submission, submission.ID)

	c.JSON(http.StatusOK, gin.H{"submission": submission})
}

func (h *SubmissionHandler) Submit(c *gin.Context) {
	// Check if submissions are open
	if !GetSubmissionsOpen() {
		c.JSON(http.StatusForbidden, gin.H{"error": "عملية التقديم مغلقة حالياً"})
		return
	}

	id := c.Param("id")

	var submission models.Submission
	if err := database.DB.First(&submission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	if submission.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only submit draft submissions"})
		return
	}

	// Check ownership
	user, _ := c.Get("user")
	u := user.(models.User)
	if u.UniversityID == nil || *u.UniversityID != submission.UniversityID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	now := time.Now()
	submission.Status = "submitted"
	submission.SubmittedAt = &now

	database.DB.Save(&submission)

	c.JSON(http.StatusOK, gin.H{"submission": submission, "message": "Submission submitted successfully"})
}

func (h *SubmissionHandler) Diff(c *gin.Context) {
	id := c.Param("id")
	versionStr := c.Param("version")

	compareVersion, err := strconv.Atoi(versionStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version number"})
		return
	}

	// Get the current submission
	var current models.Submission
	if err := database.DB.Preload("Items.Criteria").First(&current, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	// Find the other version for comparison
	var other models.Submission
	if err := database.DB.Preload("Items.Criteria").
		Where("university_id = ? AND academic_year_id = ? AND version = ?",
			current.UniversityID, current.AcademicYearID, compareVersion).
		First(&other).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Version %d not found", compareVersion)})
		return
	}

	// Build diff
	type DiffItem struct {
		CriteriaID   uint    `json:"criteria_id"`
		CriteriaName string  `json:"criteria_name"`
		OldEvidence  string  `json:"old_evidence"`
		NewEvidence  string  `json:"new_evidence"`
		OldScore     float64 `json:"old_score"`
		NewScore     float64 `json:"new_score"`
		Changed      bool    `json:"changed"`
	}

	// Map other items by criteria_id
	otherMap := map[uint]models.SubmissionItem{}
	for _, item := range other.Items {
		otherMap[item.CriteriaID] = item
	}

	var diffs []DiffItem
	for _, item := range current.Items {
		criteriaName := ""
		if item.Criteria != nil {
			criteriaName = item.Criteria.NameAr
		}

		diff := DiffItem{
			CriteriaID:   item.CriteriaID,
			CriteriaName: criteriaName,
			NewEvidence:  item.Evidence,
			NewScore:     item.Score,
		}

		if otherItem, ok := otherMap[item.CriteriaID]; ok {
			diff.OldEvidence = otherItem.Evidence
			diff.OldScore = otherItem.Score
			diff.Changed = item.Evidence != otherItem.Evidence || item.Score != otherItem.Score
		} else {
			diff.Changed = true
		}

		diffs = append(diffs, diff)
	}

	c.JSON(http.StatusOK, gin.H{
		"current_version": current.Version,
		"compare_version": compareVersion,
		"diffs":           diffs,
	})
}

func (h *SubmissionHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Limit file size from settings
	maxMB, _ := strconv.Atoi(getSetting("max_file_size_mb", "10"))
	if maxMB < 1 {
		maxMB = 10
	}
	maxSize := int64(maxMB) << 20
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("حجم الملف يتجاوز الحد المسموح (%d ميغابايت)", maxMB)})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	savePath := filepath.Join(h.Config.UploadDir, filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"filename": filename,
		"path":     "/uploads/" + filename,
	})
}
