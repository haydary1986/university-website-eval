package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"website-eval-system/database"
	"website-eval-system/models"
	"website-eval-system/services"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	AIService *services.AIService
}

func NewAIHandler(aiService *services.AIService) *AIHandler {
	return &AIHandler{AIService: aiService}
}

// getAIProvider returns an AI provider, preferring DB settings over env vars
func getAIProvider(svc *services.AIService, name string) (services.AIProvider, error) {
	dsKey, dsURL, gKey, gURL := GetAISettings()

	switch name {
	case "deepseek":
		key := dsKey
		if key == "" {
			key = svc.Config.DeepSeekKey
		}
		if key == "" {
			return nil, fmt.Errorf("مفتاح DeepSeek API غير مُعدّ")
		}
		url := dsURL
		if url == "" {
			url = svc.Config.DeepSeekURL
		}
		return &services.DeepSeekClient{APIKey: key, BaseURL: url}, nil
	case "gemini":
		key := gKey
		if key == "" {
			key = svc.Config.GeminiKey
		}
		if key == "" {
			return nil, fmt.Errorf("مفتاح Gemini API غير مُعدّ")
		}
		url := gURL
		if url == "" {
			url = svc.Config.GeminiURL
		}
		return &services.GeminiClient{APIKey: key, BaseURL: url}, nil
	default:
		// Try deepseek first, then gemini
		if dsKey != "" || svc.Config.DeepSeekKey != "" {
			key := dsKey
			if key == "" {
				key = svc.Config.DeepSeekKey
			}
			url := dsURL
			if url == "" {
				url = svc.Config.DeepSeekURL
			}
			return &services.DeepSeekClient{APIKey: key, BaseURL: url}, nil
		}
		if gKey != "" || svc.Config.GeminiKey != "" {
			key := gKey
			if key == "" {
				key = svc.Config.GeminiKey
			}
			url := gURL
			if url == "" {
				url = svc.Config.GeminiURL
			}
			return &services.GeminiClient{APIKey: key, BaseURL: url}, nil
		}
		return nil, fmt.Errorf("لم يتم تكوين أي مزود ذكاء اصطناعي")
	}
}

func (h *AIHandler) AnalyzeSubmission(c *gin.Context) {
	id := c.Param("id")

	var submission models.Submission
	if err := database.DB.
		Preload("University").
		Preload("AcademicYear").
		Preload("Items.Criteria").
		Preload("Items.Criteria.Category").
		First(&submission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	var req models.AIAnalysisRequest
	c.ShouldBindJSON(&req)

	provider, err := getAIProvider(h.AIService, req.Provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build prompt
	prompt := buildAnalysisPrompt(submission)

	result, err := provider.Chat(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI analysis failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"submission_id": submission.ID,
		"university":    submission.University.Name,
		"analysis":      result,
	})
}

func (h *AIHandler) SuggestImprovements(c *gin.Context) {
	id := c.Param("id")

	var submission models.Submission
	if err := database.DB.
		Preload("University").
		Preload("AcademicYear").
		Preload("Items.Criteria").
		Preload("Items.Criteria.Category").
		First(&submission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	var req models.AIAnalysisRequest
	c.ShouldBindJSON(&req)

	provider, err := getAIProvider(h.AIService, req.Provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prompt := buildImprovementPrompt(submission)

	result, err := provider.Chat(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI suggestion failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"submission_id": submission.ID,
		"university":    submission.University.Name,
		"suggestions":   result,
	})
}

func (h *AIHandler) CompareUniversities(c *gin.Context) {
	var req models.AICompareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if len(req.UniversityIDs) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least 2 universities required for comparison"})
		return
	}

	provider, err := getAIProvider(h.AIService, req.Provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch submissions for each university
	var submissions []models.Submission
	for _, uid := range req.UniversityIDs {
		var sub models.Submission
		if err := database.DB.
			Preload("University").
			Preload("Items.Criteria").
			Preload("Items.Criteria.Category").
			Where("university_id = ? AND academic_year_id = ? AND status = ?", uid, req.AcademicYearID, "approved").
			Order("version DESC").
			First(&sub).Error; err != nil {
			continue
		}
		submissions = append(submissions, sub)
	}

	if len(submissions) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough approved submissions found for comparison"})
		return
	}

	prompt := buildComparisonPrompt(submissions)

	result, err := provider.Chat(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI comparison failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comparison": result,
	})
}

func buildAnalysisPrompt(sub models.Submission) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("قم بتحليل نتائج تقييم موقع جامعة %s الالكتروني.\n\n", sub.University.Name))
	sb.WriteString(fmt.Sprintf("المجموع الكلي: %.1f\n\n", sub.TotalScore))
	sb.WriteString("تفاصيل التقييم حسب المعايير:\n\n")

	// Group items by category
	categoryItems := map[string][]models.SubmissionItem{}
	categoryMax := map[string]float64{}
	for _, item := range sub.Items {
		catName := "غير مصنف"
		if item.Criteria != nil && item.Criteria.Category != nil {
			catName = item.Criteria.Category.NameAr
			categoryMax[catName] = item.Criteria.Category.Weight
		}
		categoryItems[catName] = append(categoryItems[catName], item)
	}

	for catName, items := range categoryItems {
		var catScore float64
		for _, item := range items {
			catScore += item.Score
		}
		sb.WriteString(fmt.Sprintf("- %s: %.1f / %.1f\n", catName, catScore, categoryMax[catName]))
		for _, item := range items {
			criteriaName := ""
			if item.Criteria != nil {
				criteriaName = item.Criteria.NameAr
			}
			sb.WriteString(fmt.Sprintf("  * %s: %.1f / %.1f", criteriaName, item.Score, item.Criteria.MaxScore))
			if item.Evidence != "" {
				sb.WriteString(fmt.Sprintf(" (الدليل: %s)", item.Evidence))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\nقدم تحليلاً شاملاً يتضمن:\n")
	sb.WriteString("1. نقاط القوة الرئيسية\n")
	sb.WriteString("2. نقاط الضعف الرئيسية\n")
	sb.WriteString("3. التقييم العام للموقع\n")
	sb.WriteString("4. مقارنة مع المعايير المطلوبة\n")

	return sb.String()
}

func buildImprovementPrompt(sub models.Submission) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("بناءً على تقييم موقع جامعة %s الالكتروني، قدم اقتراحات للتحسين.\n\n", sub.University.Name))

	// Find low-scoring categories
	sb.WriteString("المعايير التي تحتاج تحسين (حصلت على أقل من 50%% من الدرجة):\n\n")

	for _, item := range sub.Items {
		if item.Criteria != nil && item.Criteria.MaxScore > 0 {
			percentage := (item.Score / item.Criteria.MaxScore) * 100
			if percentage < 50 {
				sb.WriteString(fmt.Sprintf("- %s: %.1f / %.1f (%.0f%%)\n", item.Criteria.NameAr, item.Score, item.Criteria.MaxScore, percentage))
			}
		}
	}

	sb.WriteString("\nقدم اقتراحات محددة وعملية لتحسين كل معيار مذكور أعلاه، مع:\n")
	sb.WriteString("1. خطوات عملية للتنفيذ\n")
	sb.WriteString("2. الأولوية (عالية، متوسطة، منخفضة)\n")
	sb.WriteString("3. الأثر المتوقع على الدرجة\n")
	sb.WriteString("4. أمثلة من جامعات عراقية أو عربية ناجحة\n")

	return sb.String()
}

func buildComparisonPrompt(subs []models.Submission) string {
	var sb strings.Builder
	sb.WriteString("قم بمقارنة مواقع الجامعات التالية:\n\n")

	for _, sub := range subs {
		sb.WriteString(fmt.Sprintf("## %s (المجموع: %.1f)\n", sub.University.Name, sub.TotalScore))

		categoryScores := map[string]float64{}
		categoryMax := map[string]float64{}
		for _, item := range sub.Items {
			if item.Criteria != nil && item.Criteria.Category != nil {
				catName := item.Criteria.Category.NameAr
				categoryScores[catName] += item.Score
				categoryMax[catName] = item.Criteria.Category.Weight
			}
		}

		for catName, score := range categoryScores {
			sb.WriteString(fmt.Sprintf("- %s: %.1f / %.1f\n", catName, score, categoryMax[catName]))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("قدم مقارنة شاملة تتضمن:\n")
	sb.WriteString("1. ترتيب الجامعات حسب الأداء\n")
	sb.WriteString("2. نقاط القوة لكل جامعة\n")
	sb.WriteString("3. المجالات التي تتفوق فيها كل جامعة\n")
	sb.WriteString("4. توصيات لكل جامعة\n")

	return sb.String()
}
