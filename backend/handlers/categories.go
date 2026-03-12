package handlers

import (
	"net/http"

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"category": category})
}
