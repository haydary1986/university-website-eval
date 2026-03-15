package handlers

import (
	"encoding/csv"
	"fmt"
	"strconv"

	"website-eval-system/database"
	"website-eval-system/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ExportHandler struct{}

func NewExportHandler() *ExportHandler {
	return &ExportHandler{}
}

// ExportRankings exports university rankings as CSV
func (h *ExportHandler) ExportRankings(c *gin.Context) {
	yearID := c.Query("academic_year_id")

	var submissions []models.Submission
	query := database.DB.Preload("University").Preload("AcademicYear").
		Where("status = ?", "approved").
		Order("total_score DESC")

	if yearID != "" {
		query = query.Where("academic_year_id = ?", yearID)
	}

	query.Find(&submissions)

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=rankings.csv")
	// BOM for Excel Arabic support
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"الترتيب", "الجامعة", "النوع", "السنة الدراسية", "الدرجة", "النسبة المئوية"})

	// Get total possible score
	var maxPossible float64
	database.DB.Model(&models.Criteria{}).
		Joins("JOIN categories ON criteria.category_id = categories.id AND categories.is_bonus = ?", false).
		Select("COALESCE(SUM(max_score), 0)").Scan(&maxPossible)

	for i, s := range submissions {
		uniType := "حكومية"
		if s.University != nil && s.University.Type == "private" {
			uniType = "أهلية"
		}
		uniName := ""
		if s.University != nil {
			uniName = s.University.Name
		}
		yearName := ""
		if s.AcademicYear != nil {
			yearName = s.AcademicYear.Name
		}
		pct := float64(0)
		if maxPossible > 0 {
			pct = (s.TotalScore / maxPossible) * 100
		}
		w.Write([]string{
			strconv.Itoa(i + 1),
			uniName,
			uniType,
			yearName,
			fmt.Sprintf("%.1f", s.TotalScore),
			fmt.Sprintf("%.1f%%", pct),
		})
	}
	w.Flush()
}

// ExportCategoryRankings exports per-category rankings as CSV
func (h *ExportHandler) ExportCategoryRankings(c *gin.Context) {
	yearID := c.Query("academic_year_id")

	var categories []models.Category
	database.DB.Preload("Criteria").Where("is_bonus = ?", false).Order("sort_order ASC").Find(&categories)

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=category_rankings.csv")
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"الفئة", "الوزن", "الجامعة", "الدرجة", "النسبة المئوية"})

	for _, cat := range categories {
		var criteriaIDs []uint
		for _, cr := range cat.Criteria {
			criteriaIDs = append(criteriaIDs, cr.ID)
		}
		if len(criteriaIDs) == 0 {
			continue
		}

		type uniScore struct {
			UniversityID uint
			Score        float64
		}

		query := database.DB.Model(&models.SubmissionItem{}).
			Select("submissions.university_id, SUM(submission_items.score) as score").
			Joins("JOIN submissions ON submission_items.submission_id = submissions.id").
			Where("submission_items.criteria_id IN ? AND submissions.status = ?", criteriaIDs, "approved").
			Group("submissions.university_id").
			Order("score DESC")

		if yearID != "" {
			query = query.Where("submissions.academic_year_id = ?", yearID)
		}

		var scores []uniScore
		query.Find(&scores)

		for _, s := range scores {
			var uni models.University
			database.DB.First(&uni, s.UniversityID)
			pct := float64(0)
			if cat.Weight > 0 {
				pct = (s.Score / cat.Weight) * 100
			}
			w.Write([]string{
				cat.NameAr,
				fmt.Sprintf("%.0f", cat.Weight),
				uni.Name,
				fmt.Sprintf("%.1f", s.Score),
				fmt.Sprintf("%.1f%%", pct),
			})
		}
	}
	w.Flush()
}

// ExportSubmissions exports all submissions as CSV
func (h *ExportHandler) ExportSubmissions(c *gin.Context) {
	var submissions []models.Submission
	query := database.DB.Preload("University").Preload("AcademicYear").
		Preload("Items.Criteria", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Category")
		}).
		Where("status != ?", "draft").
		Order("submitted_at DESC")

	if yearID := c.Query("academic_year_id"); yearID != "" {
		query = query.Where("academic_year_id = ?", yearID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	query.Find(&submissions)

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=submissions.csv")
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"الجامعة", "السنة", "النسخة", "الحالة", "الدرجة الكلية", "تاريخ التقديم", "الشخص المخول"})

	statusMap := map[string]string{
		"submitted":    "مقدم",
		"under_review": "قيد المراجعة",
		"approved":     "معتمد",
		"rejected":     "مرفوض",
	}

	for _, s := range submissions {
		uniName := ""
		if s.University != nil {
			uniName = s.University.Name
		}
		yearName := ""
		if s.AcademicYear != nil {
			yearName = s.AcademicYear.Name
		}
		submittedAt := ""
		if s.SubmittedAt != nil {
			submittedAt = s.SubmittedAt.Format("2006-01-02")
		}
		status := statusMap[s.Status]
		if status == "" {
			status = s.Status
		}

		w.Write([]string{
			uniName,
			yearName,
			strconv.Itoa(s.Version),
			status,
			fmt.Sprintf("%.1f", s.TotalScore),
			submittedAt,
			s.AuthorizedPerson,
		})
	}
	w.Flush()
}
