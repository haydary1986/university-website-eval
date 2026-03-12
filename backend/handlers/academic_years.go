package handlers

import (
	"net/http"
	"time"

	"website-eval-system/database"
	"website-eval-system/models"

	"github.com/gin-gonic/gin"
)

type AcademicYearHandler struct{}

func NewAcademicYearHandler() *AcademicYearHandler {
	return &AcademicYearHandler{}
}

func (h *AcademicYearHandler) List(c *gin.Context) {
	var years []models.AcademicYear
	if err := database.DB.Order("start_date DESC").Find(&years).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch academic years"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"academic_years": years})
}

func (h *AcademicYearHandler) Create(c *gin.Context) {
	var req models.AcademicYearRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format (use YYYY-MM-DD)"})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format (use YYYY-MM-DD)"})
		return
	}

	year := models.AcademicYear{
		Name:      req.Name,
		StartDate: startDate,
		EndDate:   endDate,
		IsActive:  req.IsActive,
	}

	// If setting as active, deactivate others
	if req.IsActive {
		database.DB.Model(&models.AcademicYear{}).Where("is_active = ?", true).Update("is_active", false)
	}

	if err := database.DB.Create(&year).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create academic year"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"academic_year": year})
}

func (h *AcademicYearHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var year models.AcademicYear
	if err := database.DB.First(&year, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Academic year not found"})
		return
	}

	var req models.AcademicYearRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
		return
	}

	// If setting as active, deactivate others
	if req.IsActive {
		database.DB.Model(&models.AcademicYear{}).Where("is_active = ? AND id != ?", true, year.ID).Update("is_active", false)
	}

	year.Name = req.Name
	year.StartDate = startDate
	year.EndDate = endDate
	year.IsActive = req.IsActive

	database.DB.Save(&year)

	c.JSON(http.StatusOK, gin.H{"academic_year": year})
}
