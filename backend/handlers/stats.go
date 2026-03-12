package handlers

import (
	"net/http"

	"website-eval-system/database"
	"website-eval-system/models"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct{}

func NewStatsHandler() *StatsHandler {
	return &StatsHandler{}
}

func (h *StatsHandler) Overview(c *gin.Context) {
	var totalUniversities int64
	database.DB.Model(&models.University{}).Count(&totalUniversities)

	var totalSubmissions int64
	database.DB.Model(&models.Submission{}).Count(&totalSubmissions)

	var approvedSubmissions int64
	database.DB.Model(&models.Submission{}).Where("status = ?", "approved").Count(&approvedSubmissions)

	var pendingSubmissions int64
	database.DB.Model(&models.Submission{}).Where("status = ?", "submitted").Count(&pendingSubmissions)

	var avgScore float64
	database.DB.Model(&models.Submission{}).
		Where("status = ? AND total_score > 0", "approved").
		Select("COALESCE(AVG(total_score), 0)").
		Scan(&avgScore)

	var maxScore float64
	database.DB.Model(&models.Submission{}).
		Where("status = ?", "approved").
		Select("COALESCE(MAX(total_score), 0)").
		Scan(&maxScore)

	var govCount, privateCount int64
	database.DB.Model(&models.University{}).Where("type = ?", "government").Count(&govCount)
	database.DB.Model(&models.University{}).Where("type = ?", "private").Count(&privateCount)

	c.JSON(http.StatusOK, gin.H{
		"total_universities":   totalUniversities,
		"government_count":     govCount,
		"private_count":        privateCount,
		"total_submissions":    totalSubmissions,
		"approved_submissions": approvedSubmissions,
		"pending_submissions":  pendingSubmissions,
		"average_score":        avgScore,
		"max_score":            maxScore,
	})
}

func (h *StatsHandler) Universities(c *gin.Context) {
	type UniversityRanking struct {
		UniversityID   uint    `json:"university_id"`
		UniversityName string  `json:"university_name"`
		UniversityType string  `json:"university_type"`
		TotalScore     float64 `json:"total_score"`
		AcademicYear   string  `json:"academic_year"`
		Version        int     `json:"version"`
	}

	var rankings []UniversityRanking

	query := database.DB.Table("submissions").
		Select("submissions.university_id, universities.name as university_name, universities.type as university_type, submissions.total_score, academic_years.name as academic_year, submissions.version").
		Joins("JOIN universities ON universities.id = submissions.university_id").
		Joins("JOIN academic_years ON academic_years.id = submissions.academic_year_id").
		Where("submissions.status = ?", "approved")

	if yearID := c.Query("academic_year_id"); yearID != "" {
		query = query.Where("submissions.academic_year_id = ?", yearID)
	}

	if uniType := c.Query("type"); uniType != "" {
		query = query.Where("universities.type = ?", uniType)
	}

	query.Order("submissions.total_score DESC").Scan(&rankings)

	c.JSON(http.StatusOK, gin.H{"rankings": rankings})
}

func (h *StatsHandler) Categories(c *gin.Context) {
	type CategoryAvg struct {
		CategoryID   uint    `json:"category_id"`
		CategoryName string  `json:"category_name"`
		MaxPossible  float64 `json:"max_possible"`
		AvgScore     float64 `json:"avg_score"`
		MinScore     float64 `json:"min_score"`
		MaxScore     float64 `json:"max_score"`
	}

	var stats []CategoryAvg

	database.DB.Table("submission_items").
		Select("criteria.category_id, categories.name_ar as category_name, categories.weight as max_possible, "+
			"AVG(submission_items.score) as avg_score, MIN(submission_items.score) as min_score, MAX(submission_items.score) as max_score").
		Joins("JOIN criteria ON criteria.id = submission_items.criteria_id").
		Joins("JOIN categories ON categories.id = criteria.category_id").
		Joins("JOIN submissions ON submissions.id = submission_items.submission_id").
		Where("submissions.status = ?", "approved").
		Group("criteria.category_id, categories.name_ar, categories.weight").
		Order("criteria.category_id ASC").
		Scan(&stats)

	c.JSON(http.StatusOK, gin.H{"category_stats": stats})
}

func (h *StatsHandler) Comparison(c *gin.Context) {
	universityID := c.Param("universityId")

	type YearScore struct {
		AcademicYearID   uint    `json:"academic_year_id"`
		AcademicYearName string  `json:"academic_year_name"`
		TotalScore       float64 `json:"total_score"`
		Version          int     `json:"version"`
	}

	var scores []YearScore

	database.DB.Table("submissions").
		Select("submissions.academic_year_id, academic_years.name as academic_year_name, submissions.total_score, submissions.version").
		Joins("JOIN academic_years ON academic_years.id = submissions.academic_year_id").
		Where("submissions.university_id = ? AND submissions.status = ?", universityID, "approved").
		Order("academic_years.start_date ASC").
		Scan(&scores)

	// Category breakdown per year
	type CategoryBreakdown struct {
		AcademicYearName string  `json:"academic_year_name"`
		CategoryName     string  `json:"category_name"`
		Score            float64 `json:"score"`
		MaxPossible      float64 `json:"max_possible"`
	}

	var breakdown []CategoryBreakdown

	database.DB.Table("submission_items").
		Select("academic_years.name as academic_year_name, categories.name_ar as category_name, "+
			"SUM(submission_items.score) as score, categories.weight as max_possible").
		Joins("JOIN submissions ON submissions.id = submission_items.submission_id").
		Joins("JOIN criteria ON criteria.id = submission_items.criteria_id").
		Joins("JOIN categories ON categories.id = criteria.category_id").
		Joins("JOIN academic_years ON academic_years.id = submissions.academic_year_id").
		Where("submissions.university_id = ? AND submissions.status = ?", universityID, "approved").
		Group("academic_years.name, categories.name_ar, categories.weight").
		Order("academic_years.name ASC, categories.sort_order ASC").
		Scan(&breakdown)

	c.JSON(http.StatusOK, gin.H{
		"university_id": universityID,
		"yearly_scores": scores,
		"breakdown":     breakdown,
	})
}
