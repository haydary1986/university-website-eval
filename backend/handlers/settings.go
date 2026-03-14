package handlers

import (
	"net/http"
	"strings"

	"website-eval-system/database"
	"website-eval-system/models"
	"website-eval-system/services"

	"github.com/gin-gonic/gin"
)

// maskKey hides API key, showing only last 4 chars
func maskKey(key string) string {
	if len(key) <= 4 {
		return strings.Repeat("*", len(key))
	}
	return strings.Repeat("*", len(key)-4) + key[len(key)-4:]
}

type SettingsHandler struct{}

func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

func getSetting(key, fallback string) string {
	var s models.SystemSetting
	if err := database.DB.Where("`key` = ?", key).First(&s).Error; err != nil {
		return fallback
	}
	return s.Value
}

func setSetting(key, value string) {
	database.DB.Where("`key` = ?", key).Assign(models.SystemSetting{Value: value}).FirstOrCreate(&models.SystemSetting{Key: key})
}

// GetSettings returns all system settings (super_admin only)
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	dsKey := getSetting("deepseek_api_key", "")
	gKey := getSetting("gemini_api_key", "")

	resp := models.SystemSettingsResponse{
		SiteTitle:       getSetting("site_title", "نظام تقييم جودة المواقع الالكترونية الجامعية"),
		SiteDescription: getSetting("site_description", "نظام تقييم جودة المواقع الالكترونية للجامعات العراقية - وزارة التعليم العالي والبحث العلمي"),
		SubmissionsOpen: getSetting("submissions_open", "true") == "true",
		DeepSeekAPIKey:  maskKey(dsKey),
		DeepSeekURL:     getSetting("deepseek_url", "https://api.deepseek.com/v1/chat/completions"),
		GeminiAPIKey:    maskKey(gKey),
		GeminiURL:       getSetting("gemini_url", "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"),
		HasDeepSeekKey:  dsKey != "",
		HasGeminiKey:    gKey != "",
	}
	c.JSON(http.StatusOK, gin.H{"settings": resp})
}

// GetPublicSettings returns only public settings (title, description, submissions_open)
func (h *SettingsHandler) GetPublicSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"site_title":       getSetting("site_title", "نظام تقييم جودة المواقع الالكترونية الجامعية"),
		"site_description": getSetting("site_description", "نظام تقييم جودة المواقع الالكترونية للجامعات العراقية - وزارة التعليم العالي والبحث العلمي"),
		"submissions_open": getSetting("submissions_open", "true") == "true",
	})
}

// UpdateSettings updates system settings
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	var req models.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "طلب غير صالح: " + err.Error()})
		return
	}

	if req.SiteTitle != nil {
		setSetting("site_title", *req.SiteTitle)
	}
	if req.SiteDescription != nil {
		setSetting("site_description", *req.SiteDescription)
	}
	if req.SubmissionsOpen != nil {
		val := "false"
		if *req.SubmissionsOpen {
			val = "true"
		}
		setSetting("submissions_open", val)
	}
	if req.DeepSeekAPIKey != nil {
		setSetting("deepseek_api_key", *req.DeepSeekAPIKey)
	}
	if req.DeepSeekURL != nil {
		setSetting("deepseek_url", *req.DeepSeekURL)
	}
	if req.GeminiAPIKey != nil {
		setSetting("gemini_api_key", *req.GeminiAPIKey)
	}
	if req.GeminiURL != nil {
		setSetting("gemini_url", *req.GeminiURL)
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث الإعدادات بنجاح"})
}

// TestAI tests an AI provider connection
func (h *SettingsHandler) TestAI(c *gin.Context) {
	var req models.TestAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "طلب غير صالح: " + err.Error()})
		return
	}

	var provider services.AIProvider

	// If key is __use_saved__, use the saved key from DB
	apiKey := req.APIKey
	if apiKey == "__use_saved__" {
		dsKey, _, gKey, _ := GetAISettings()
		if req.Provider == "deepseek" {
			apiKey = dsKey
		} else {
			apiKey = gKey
		}
		if apiKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "لا يوجد مفتاح محفوظ", "success": false})
			return
		}
	}

	switch req.Provider {
	case "deepseek":
		baseURL := req.BaseURL
		if baseURL == "" {
			baseURL = "https://api.deepseek.com/v1/chat/completions"
		}
		provider = &services.DeepSeekClient{APIKey: apiKey, BaseURL: baseURL}
	case "gemini":
		baseURL := req.BaseURL
		if baseURL == "" {
			baseURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"
		}
		provider = &services.GeminiClient{APIKey: apiKey, BaseURL: baseURL}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "مزود غير معروف"})
		return
	}

	response, err := provider.Chat("Say hello in Arabic in one sentence.")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فشل الاتصال: " + err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "response": response})
}

// GetSubmissionsOpen is a helper used by other handlers
func GetSubmissionsOpen() bool {
	return getSetting("submissions_open", "true") == "true"
}

// GetAISettings returns AI keys from DB settings (with env var fallback via config)
func GetAISettings() (deepseekKey, deepseekURL, geminiKey, geminiURL string) {
	deepseekKey = getSetting("deepseek_api_key", "")
	deepseekURL = getSetting("deepseek_url", "https://api.deepseek.com/v1/chat/completions")
	geminiKey = getSetting("gemini_api_key", "")
	geminiURL = getSetting("gemini_url", "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent")
	return
}
