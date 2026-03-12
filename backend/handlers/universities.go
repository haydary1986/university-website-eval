package handlers

import (
	"net/http"

	"website-eval-system/database"
	"website-eval-system/models"

	"github.com/gin-gonic/gin"
)

type UniversityHandler struct{}

func NewUniversityHandler() *UniversityHandler {
	return &UniversityHandler{}
}

func (h *UniversityHandler) List(c *gin.Context) {
	var universities []models.University

	query := database.DB

	// Filter by type
	if t := c.Query("type"); t != "" {
		query = query.Where("type = ?", t)
	}

	// Search by name
	if search := c.Query("search"); search != "" {
		query = query.Where("name LIKE ? OR name_en LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Order("name ASC").Find(&universities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch universities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"universities": universities})
}

func (h *UniversityHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var university models.University
	if err := database.DB.First(&university, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "University not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"university": university})
}

func (h *UniversityHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var university models.University
	if err := database.DB.First(&university, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "University not found"})
		return
	}

	// Check permissions: super_admin can edit any, university user can only edit their own
	role, _ := c.Get("role")
	if role == "university" {
		user, _ := c.Get("user")
		u := user.(models.User)
		if u.UniversityID == nil || *u.UniversityID != university.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot edit other university"})
			return
		}
	}

	var input struct {
		Name          string `json:"name"`
		NameEn        string `json:"name_en"`
		Website       string `json:"website"`
		City          string `json:"city"`
		ContactPerson string `json:"contact_person"`
		ContactEmail  string `json:"contact_email"`
		ContactPhone  string `json:"contact_phone"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	updates := map[string]interface{}{}
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.NameEn != "" {
		updates["name_en"] = input.NameEn
	}
	if input.Website != "" {
		updates["website"] = input.Website
	}
	if input.City != "" {
		updates["city"] = input.City
	}
	if input.ContactPerson != "" {
		updates["contact_person"] = input.ContactPerson
	}
	if input.ContactEmail != "" {
		updates["contact_email"] = input.ContactEmail
	}
	if input.ContactPhone != "" {
		updates["contact_phone"] = input.ContactPhone
	}

	database.DB.Model(&university).Updates(updates)
	database.DB.First(&university, id)

	c.JSON(http.StatusOK, gin.H{"university": university})
}
