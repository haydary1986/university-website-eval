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
	database.DB.Model(&models.Submission{}).Where("status IN ?", []string{"submitted", "under_review"}).Count(&pendingSubmissions)

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

	// Status counts
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var statusCounts []StatusCount
	database.DB.Model(&models.Submission{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts)

	statusMap := map[string]int64{}
	for _, sc := range statusCounts {
		statusMap[sc.Status] = sc.Count
	}

	// Type comparison (avg scores)
	type TypeAvg struct {
		Type     string  `json:"type"`
		AvgScore float64 `json:"avg_score"`
	}
	var typeAvgs []TypeAvg
	database.DB.Table("submissions").
		Select("universities.type, AVG(submissions.total_score) as avg_score").
		Joins("JOIN universities ON universities.id = submissions.university_id").
		Where("submissions.status = ? AND submissions.total_score > 0", "approved").
		Group("universities.type").
		Scan(&typeAvgs)

	typeComparison := map[string]float64{}
	for _, ta := range typeAvgs {
		typeComparison[ta.Type] = ta.AvgScore
	}

	c.JSON(http.StatusOK, gin.H{
		"total_universities":   totalUniversities,
		"government_count":     govCount,
		"gov_count":            govCount,
		"private_count":        privateCount,
		"total_submissions":    totalSubmissions,
		"approved_submissions": approvedSubmissions,
		"pending_reviews":      pendingSubmissions,
		"pending_submissions":  pendingSubmissions,
		"submitted_count":      statusMap["submitted"],
		"under_review_count":   statusMap["under_review"],
		"rejected_count":       statusMap["rejected"],
		"average_score":        avgScore,
		"max_score":            maxScore,
		"status_counts":        statusMap,
		"type_comparison":      typeComparison,
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

	query := database.DB.Table("submission_items").
		Select("criteria.category_id, categories.name_ar as category_name, categories.weight as max_possible, "+
			"AVG(submission_items.score) as avg_score, MIN(submission_items.score) as min_score, MAX(submission_items.score) as max_score").
		Joins("JOIN criteria ON criteria.id = submission_items.criteria_id").
		Joins("JOIN categories ON categories.id = criteria.category_id").
		Joins("JOIN submissions ON submissions.id = submission_items.submission_id").
		Where("submissions.status = ?", "approved")

	if yearID := c.Query("academic_year_id"); yearID != "" {
		query = query.Where("submissions.academic_year_id = ?", yearID)
	}

	query.Group("criteria.category_id, categories.name_ar, categories.weight").
		Order("criteria.category_id ASC").
		Scan(&stats)

	c.JSON(http.StatusOK, gin.H{"category_stats": stats})
}

// CategoryRankings returns university rankings per category
func (h *StatsHandler) CategoryRankings(c *gin.Context) {
	type UniCategoryScore struct {
		CategoryID     uint    `json:"category_id"`
		CategoryName   string  `json:"category_name"`
		CategoryWeight float64 `json:"category_weight"`
		UniversityID   uint    `json:"university_id"`
		UniversityName string  `json:"university_name"`
		UniversityType string  `json:"university_type"`
		Score          float64 `json:"score"`
		Percentage     float64 `json:"percentage"`
	}

	var results []UniCategoryScore

	query := database.DB.Table("submission_items").
		Select(`criteria.category_id,
			categories.name_ar as category_name,
			categories.weight as category_weight,
			submissions.university_id,
			universities.name as university_name,
			universities.type as university_type,
			SUM(submission_items.score) as score,
			CASE WHEN categories.weight > 0 THEN (SUM(submission_items.score) * 100.0 / categories.weight) ELSE 0 END as percentage`).
		Joins("JOIN criteria ON criteria.id = submission_items.criteria_id").
		Joins("JOIN categories ON categories.id = criteria.category_id").
		Joins("JOIN submissions ON submissions.id = submission_items.submission_id").
		Joins("JOIN universities ON universities.id = submissions.university_id").
		Where("submissions.status = ?", "approved")

	if yearID := c.Query("academic_year_id"); yearID != "" {
		query = query.Where("submissions.academic_year_id = ?", yearID)
	}
	if uniType := c.Query("type"); uniType != "" {
		query = query.Where("universities.type = ?", uniType)
	}

	query.Group("criteria.category_id, categories.name_ar, categories.weight, submissions.university_id, universities.name, universities.type").
		Order("criteria.category_id ASC, score DESC").
		Scan(&results)

	// Organize by category with rankings
	type RankedUni struct {
		Rank           int     `json:"rank"`
		UniversityID   uint    `json:"university_id"`
		UniversityName string  `json:"university_name"`
		UniversityType string  `json:"university_type"`
		Score          float64 `json:"score"`
		Percentage     float64 `json:"percentage"`
	}

	type CategoryRanking struct {
		CategoryID     uint        `json:"category_id"`
		CategoryName   string      `json:"category_name"`
		CategoryWeight float64     `json:"category_weight"`
		Universities   []RankedUni `json:"universities"`
	}

	categoryMap := map[uint]*CategoryRanking{}
	var categoryOrder []uint

	for _, r := range results {
		cat, exists := categoryMap[r.CategoryID]
		if !exists {
			cat = &CategoryRanking{
				CategoryID:     r.CategoryID,
				CategoryName:   r.CategoryName,
				CategoryWeight: r.CategoryWeight,
				Universities:   []RankedUni{},
			}
			categoryMap[r.CategoryID] = cat
			categoryOrder = append(categoryOrder, r.CategoryID)
		}

		rank := len(cat.Universities) + 1
		cat.Universities = append(cat.Universities, RankedUni{
			Rank:           rank,
			UniversityID:   r.UniversityID,
			UniversityName: r.UniversityName,
			UniversityType: r.UniversityType,
			Score:          r.Score,
			Percentage:     r.Percentage,
		})
	}

	var rankings []CategoryRanking
	for _, catID := range categoryOrder {
		rankings = append(rankings, *categoryMap[catID])
	}

	c.JSON(http.StatusOK, gin.H{"category_rankings": rankings})
}

// UniversityProfile returns a full breakdown for a single university
func (h *StatsHandler) UniversityProfile(c *gin.Context) {
	universityID := c.Param("universityId")

	// University info
	var uni models.University
	if err := database.DB.First(&uni, universityID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الجامعة غير موجودة"})
		return
	}

	// All approved submissions
	type SubmissionSummary struct {
		ID             uint    `json:"id"`
		AcademicYear   string  `json:"academic_year"`
		AcademicYearID uint    `json:"academic_year_id"`
		Version        int     `json:"version"`
		TotalScore     float64 `json:"total_score"`
		Status         string  `json:"status"`
	}
	var submissions []SubmissionSummary
	database.DB.Table("submissions").
		Select("submissions.id, academic_years.name as academic_year, submissions.academic_year_id, submissions.version, submissions.total_score, submissions.status").
		Joins("JOIN academic_years ON academic_years.id = submissions.academic_year_id").
		Where("submissions.university_id = ? AND submissions.status = ?", universityID, "approved").
		Order("academic_years.start_date DESC, submissions.version DESC").
		Scan(&submissions)

	// Category scores for latest submission
	type CategoryScore struct {
		CategoryID   uint    `json:"category_id"`
		CategoryName string  `json:"category_name"`
		Weight       float64 `json:"weight"`
		Score        float64 `json:"score"`
		Percentage   float64 `json:"percentage"`
	}

	var categoryScores []CategoryScore
	if len(submissions) > 0 {
		latestID := submissions[0].ID
		database.DB.Table("submission_items").
			Select(`criteria.category_id,
				categories.name_ar as category_name,
				categories.weight as weight,
				SUM(submission_items.score) as score,
				CASE WHEN categories.weight > 0 THEN (SUM(submission_items.score) * 100.0 / categories.weight) ELSE 0 END as percentage`).
			Joins("JOIN criteria ON criteria.id = submission_items.criteria_id").
			Joins("JOIN categories ON categories.id = criteria.category_id").
			Where("submission_items.submission_id = ?", latestID).
			Group("criteria.category_id, categories.name_ar, categories.weight").
			Order("categories.sort_order ASC").
			Scan(&categoryScores)
	}

	// Rankings per category (what's the university's rank in each category)
	type CategoryRankInfo struct {
		CategoryID   uint    `json:"category_id"`
		CategoryName string  `json:"category_name"`
		Rank         int     `json:"rank"`
		TotalInRank  int     `json:"total_in_rank"`
		Score        float64 `json:"score"`
		Weight       float64 `json:"weight"`
	}

	var categoryRanks []CategoryRankInfo

	// Get latest academic year
	yearFilter := c.Query("academic_year_id")
	if yearFilter == "" && len(submissions) > 0 {
		// Use latest submission's year
		for _, s := range submissions {
			yearFilter = string(rune(s.AcademicYearID))
			break
		}
	}

	// Get all university scores per category for ranking calculation
	type UniCatScore struct {
		CategoryID     uint    `json:"category_id"`
		CategoryName   string  `json:"category_name"`
		CategoryWeight float64 `json:"category_weight"`
		UniversityID   uint    `json:"university_id"`
		Score          float64 `json:"score"`
	}

	var allScores []UniCatScore
	rankQuery := database.DB.Table("submission_items").
		Select(`criteria.category_id, categories.name_ar as category_name, categories.weight as category_weight,
			submissions.university_id, SUM(submission_items.score) as score`).
		Joins("JOIN criteria ON criteria.id = submission_items.criteria_id").
		Joins("JOIN categories ON categories.id = criteria.category_id").
		Joins("JOIN submissions ON submissions.id = submission_items.submission_id").
		Where("submissions.status = ?", "approved")

	if yearFilter != "" {
		rankQuery = rankQuery.Where("submissions.academic_year_id = ?", yearFilter)
	}

	rankQuery.Group("criteria.category_id, categories.name_ar, categories.weight, submissions.university_id").
		Order("criteria.category_id ASC, score DESC").
		Scan(&allScores)

	// Calculate ranks
	currentCat := uint(0)
	rankCounter := 0
	catTotals := map[uint]int{}

	// First count totals per category
	for _, s := range allScores {
		catTotals[s.CategoryID]++
	}

	for _, s := range allScores {
		if s.CategoryID != currentCat {
			currentCat = s.CategoryID
			rankCounter = 0
		}
		rankCounter++
		if s.UniversityID == uni.ID {
			categoryRanks = append(categoryRanks, CategoryRankInfo{
				CategoryID:   s.CategoryID,
				CategoryName: s.CategoryName,
				Rank:         rankCounter,
				TotalInRank:  catTotals[s.CategoryID],
				Score:        s.Score,
				Weight:       s.CategoryWeight,
			})
		}
	}

	// Year over year scores
	type YearScore struct {
		AcademicYear string  `json:"academic_year"`
		TotalScore   float64 `json:"total_score"`
		Version      int     `json:"version"`
	}
	var yearlyScores []YearScore
	database.DB.Table("submissions").
		Select("academic_years.name as academic_year, submissions.total_score, submissions.version").
		Joins("JOIN academic_years ON academic_years.id = submissions.academic_year_id").
		Where("submissions.university_id = ? AND submissions.status = ?", universityID, "approved").
		Order("academic_years.start_date ASC").
		Scan(&yearlyScores)

	// Overall rank
	type OverallRank struct {
		UniversityID uint    `json:"university_id"`
		TotalScore   float64 `json:"total_score"`
	}
	var overallRankings []OverallRank
	overallQuery := database.DB.Table("submissions").
		Select("university_id, total_score").
		Where("status = ?", "approved")
	if yearFilter != "" {
		overallQuery = overallQuery.Where("academic_year_id = ?", yearFilter)
	}
	overallQuery.Order("total_score DESC").Scan(&overallRankings)

	overallRank := 0
	for i, r := range overallRankings {
		if r.UniversityID == uni.ID {
			overallRank = i + 1
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"university":      uni,
		"submissions":     submissions,
		"category_scores": categoryScores,
		"category_ranks":  categoryRanks,
		"yearly_scores":   yearlyScores,
		"overall_rank":    overallRank,
		"total_ranked":    len(overallRankings),
	})
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
