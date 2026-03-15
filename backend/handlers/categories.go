package handlers

import (
	"net/http"
	"strconv"

	"website-eval-system/database"
	"website-eval-system/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CategoryHandler struct{}

func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{}
}

func (h *CategoryHandler) List(c *gin.Context) {
	var categories []models.Category
	query := database.DB.Preload("Criteria", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).Order("sort_order ASC")

	if err := query.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في جلب الفئات"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

func (h *CategoryHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var category models.Category
	if err := database.DB.Preload("Criteria", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الفئة غير موجودة"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"category": category})
}

// Create creates a new category (super_admin only)
func (h *CategoryHandler) Create(c *gin.Context) {
	var req struct {
		Number  int     `json:"number"`
		NameAr  string  `json:"name_ar" binding:"required"`
		Weight  float64 `json:"weight"`
		IsBonus bool    `json:"is_bonus"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	// Get max sort order
	var maxSort int
	database.DB.Model(&models.Category{}).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxSort)

	// Auto-assign number if not provided
	number := req.Number
	if number == 0 {
		var maxNum int
		database.DB.Model(&models.Category{}).Select("COALESCE(MAX(number), 0)").Scan(&maxNum)
		number = maxNum + 1
	}

	category := models.Category{
		Number:    number,
		NameAr:    req.NameAr,
		Weight:    req.Weight,
		SortOrder: maxSort + 1,
		IsBonus:   req.IsBonus,
	}

	if err := database.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في إنشاء الفئة"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"category": category})
}

// Update updates a category
func (h *CategoryHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var category models.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الفئة غير موجودة"})
		return
	}

	var req struct {
		Number  *int     `json:"number"`
		NameAr  *string  `json:"name_ar"`
		Weight  *float64 `json:"weight"`
		IsBonus *bool    `json:"is_bonus"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	if req.Number != nil {
		category.Number = *req.Number
	}
	if req.NameAr != nil {
		category.NameAr = *req.NameAr
	}
	if req.Weight != nil {
		category.Weight = *req.Weight
	}
	if req.IsBonus != nil {
		category.IsBonus = *req.IsBonus
	}

	if err := database.DB.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في تحديث الفئة"})
		return
	}

	// Reload with criteria
	database.DB.Preload("Criteria", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).First(&category, category.ID)

	c.JSON(http.StatusOK, gin.H{"category": category})
}

// Delete deletes a category (only if no submissions reference it)
func (h *CategoryHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	var category models.Category
	if err := database.DB.Preload("Criteria").First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الفئة غير موجودة"})
		return
	}

	// Check if any submission items reference criteria in this category
	var criteriaIDs []uint
	for _, cr := range category.Criteria {
		criteriaIDs = append(criteriaIDs, cr.ID)
	}

	if len(criteriaIDs) > 0 {
		var count int64
		database.DB.Model(&models.SubmissionItem{}).Where("criteria_id IN ?", criteriaIDs).Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "لا يمكن حذف الفئة لوجود تقديمات مرتبطة بها"})
			return
		}
	}

	// Delete criteria first, then category
	if len(criteriaIDs) > 0 {
		database.DB.Where("category_id = ?", category.ID).Delete(&models.Criteria{})
	}
	database.DB.Delete(&category)

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الفئة بنجاح"})
}

// CreateCriteria adds a new criteria to a category
func (h *CategoryHandler) CreateCriteria(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرّف الفئة غير صالح"})
		return
	}

	// Verify category exists
	var category models.Category
	if err := database.DB.First(&category, categoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الفئة غير موجودة"})
		return
	}

	var req struct {
		NameAr      string  `json:"name_ar" binding:"required"`
		Description string  `json:"description"`
		MaxScore    float64 `json:"max_score"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	// Get max sort order within category
	var maxSort int
	database.DB.Model(&models.Criteria{}).Where("category_id = ?", categoryID).
		Select("COALESCE(MAX(sort_order), 0)").Scan(&maxSort)

	criteria := models.Criteria{
		CategoryID:  uint(categoryID),
		NameAr:      req.NameAr,
		Description: req.Description,
		MaxScore:    req.MaxScore,
		SortOrder:   maxSort + 1,
	}

	if err := database.DB.Create(&criteria).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في إنشاء الفقرة"})
		return
	}

	// Update category weight = sum of criteria max_score
	h.updateCategoryWeight(uint(categoryID))

	c.JSON(http.StatusCreated, gin.H{"criteria": criteria})
}

// UpdateCriteria updates a criteria
func (h *CategoryHandler) UpdateCriteria(c *gin.Context) {
	id := c.Param("id")

	var criteria models.Criteria
	if err := database.DB.First(&criteria, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الفقرة غير موجودة"})
		return
	}

	var req struct {
		NameAr      *string  `json:"name_ar"`
		Description *string  `json:"description"`
		MaxScore    *float64 `json:"max_score"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	if req.NameAr != nil {
		criteria.NameAr = *req.NameAr
	}
	if req.Description != nil {
		criteria.Description = *req.Description
	}
	if req.MaxScore != nil {
		criteria.MaxScore = *req.MaxScore
	}

	if err := database.DB.Save(&criteria).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في تحديث الفقرة"})
		return
	}

	// Update category weight
	h.updateCategoryWeight(criteria.CategoryID)

	c.JSON(http.StatusOK, gin.H{"criteria": criteria})
}

// DeleteCriteria deletes a criteria
func (h *CategoryHandler) DeleteCriteria(c *gin.Context) {
	id := c.Param("id")

	var criteria models.Criteria
	if err := database.DB.First(&criteria, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الفقرة غير موجودة"})
		return
	}

	// Check if submission items reference this criteria
	var count int64
	database.DB.Model(&models.SubmissionItem{}).Where("criteria_id = ?", criteria.ID).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "لا يمكن حذف الفقرة لوجود تقديمات مرتبطة بها"})
		return
	}

	categoryID := criteria.CategoryID
	database.DB.Delete(&criteria)

	// Update category weight
	h.updateCategoryWeight(categoryID)

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الفقرة بنجاح"})
}

// updateCategoryWeight recalculates category weight from sum of criteria max_score
func (h *CategoryHandler) updateCategoryWeight(categoryID uint) {
	var total float64
	database.DB.Model(&models.Criteria{}).Where("category_id = ?", categoryID).
		Select("COALESCE(SUM(max_score), 0)").Scan(&total)
	database.DB.Model(&models.Category{}).Where("id = ?", categoryID).Update("weight", total)
}
