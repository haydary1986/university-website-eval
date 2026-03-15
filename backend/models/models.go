package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONArray is a custom type for storing JSON arrays in SQLite
type JSONArray []int

func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = JSONArray{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return errors.New("unsupported type for JSONArray")
	}
	return json.Unmarshal(bytes, j)
}

type User struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	Username           string     `json:"username" gorm:"uniqueIndex;not null"`
	Password           string     `json:"-" gorm:"not null"`
	FullName           string     `json:"full_name"`
	Email              string     `json:"email"`
	Phone              string     `json:"phone"`
	Role               string     `json:"role" gorm:"not null;default:'university'"` // super_admin, admin, university
	MustChangePassword bool       `json:"must_change_password" gorm:"default:false"`
	IsBlocked          bool       `json:"is_blocked" gorm:"default:false"`
	BlockedUntil       *time.Time `json:"blocked_until"`
	FailedAttempts     int        `json:"failed_attempts" gorm:"default:0"`
	LastFailedAt       *time.Time `json:"last_failed_at"`
	LastLoginAt        *time.Time `json:"last_login_at"`
	LastLoginIP        string     `json:"last_login_ip"`
	PasswordChangedAt  *time.Time `json:"password_changed_at"`
	UniversityID       *uint      `json:"university_id"`
	University         *University `json:"university,omitempty" gorm:"foreignKey:UniversityID"`
	AssignedCategories JSONArray  `json:"assigned_categories" gorm:"type:text"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type LoginAttempt struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"index;not null"`
	UserID    *uint     `json:"user_id"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	IPAddress string    `json:"ip_address" gorm:"index"`
	UserAgent string    `json:"user_agent"`
	Success   bool      `json:"success" gorm:"default:false"`
	Reason    string    `json:"reason"` // success, invalid_password, user_not_found, account_blocked
	CreatedAt time.Time `json:"created_at"`
}

type AuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Action    string    `json:"action" gorm:"not null"` // password_change, login, login_failed, account_blocked, account_unblocked, etc.
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

type BlockedIP struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	IPAddress string    `json:"ip_address" gorm:"uniqueIndex;not null"`
	Reason    string    `json:"reason"`
	BlockedBy uint      `json:"blocked_by"`
	Admin     *User     `json:"admin,omitempty" gorm:"foreignKey:BlockedBy"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type ActiveSession struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	TokenHash string    `json:"-" gorm:"uniqueIndex;not null"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type University struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	NameEn       string    `json:"name_en"`
	Type         string    `json:"type" gorm:"not null;default:'government'"` // government, private
	Website      string    `json:"website"`
	City         string    `json:"city"`
	ContactPerson string  `json:"contact_person"`
	ContactEmail  string  `json:"contact_email"`
	ContactPhone  string  `json:"contact_phone"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AcademicYear struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	Name               string     `json:"name" gorm:"not null;uniqueIndex"`
	StartDate          time.Time  `json:"start_date"`
	EndDate            time.Time  `json:"end_date"`
	SubmissionDeadline *time.Time `json:"submission_deadline"`
	IsActive           bool       `json:"is_active" gorm:"default:false"`
	CreatedAt          time.Time  `json:"created_at"`
}

type Category struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Number    int        `json:"number"`
	NameAr    string     `json:"name_ar" gorm:"not null"`
	Weight    float64    `json:"weight"`
	SortOrder int        `json:"sort_order"`
	IsBonus   bool       `json:"is_bonus" gorm:"default:false"`
	Criteria  []Criteria `json:"criteria,omitempty" gorm:"foreignKey:CategoryID"`
}

type Criteria struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	CategoryID  uint    `json:"category_id" gorm:"not null"`
	Category    *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	NameAr      string  `json:"name_ar" gorm:"not null"`
	Description string  `json:"description"`
	MaxScore    float64 `json:"max_score"`
	SortOrder   int     `json:"sort_order"`
}

type Submission struct {
	ID               uint             `json:"id" gorm:"primaryKey"`
	UniversityID     uint             `json:"university_id" gorm:"not null;index"`
	University       *University      `json:"university,omitempty" gorm:"foreignKey:UniversityID"`
	AcademicYearID   uint             `json:"academic_year_id" gorm:"not null;index"`
	AcademicYear     *AcademicYear    `json:"academic_year,omitempty" gorm:"foreignKey:AcademicYearID"`
	Version          int              `json:"version" gorm:"not null;default:1"`
	Status           string           `json:"status" gorm:"not null;default:'draft'"` // draft, submitted, under_review, approved, rejected
	AuthorizedPerson string           `json:"authorized_person"`
	AuthorizedPhone  string           `json:"authorized_phone"`
	AuthorizedEmail  string           `json:"authorized_email"`
	SubmittedAt      *time.Time       `json:"submitted_at"`
	ReviewedAt       *time.Time       `json:"reviewed_at"`
	TotalScore       float64          `json:"total_score" gorm:"default:0"`
	RejectReason     string           `json:"reject_reason"`
	Items            []SubmissionItem `json:"items,omitempty" gorm:"foreignKey:SubmissionID"`
	Reviews          []Review         `json:"reviews,omitempty" gorm:"foreignKey:SubmissionID"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type SubmissionItem struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	SubmissionID uint      `json:"submission_id" gorm:"not null;index"`
	CriteriaID   uint      `json:"criteria_id" gorm:"not null;index"`
	Criteria     *Criteria `json:"criteria,omitempty" gorm:"foreignKey:CriteriaID"`
	Evidence     string    `json:"evidence"`
	EvidenceFile string    `json:"evidence_file"`
	Score        float64   `json:"score" gorm:"default:0"`
	AdminComment string    `json:"admin_comment"`
	Status       string    `json:"status" gorm:"default:'pending'"` // pending, approved, rejected
}

type Review struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	SubmissionID uint      `json:"submission_id" gorm:"not null"`
	AdminID      uint      `json:"admin_id" gorm:"not null"`
	Admin        *User     `json:"admin,omitempty" gorm:"foreignKey:AdminID"`
	CategoryID   uint      `json:"category_id" gorm:"not null"`
	Category     *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Status       string    `json:"status" gorm:"default:'pending'"` // pending, reviewed
	Comments     string    `json:"comments"`
	CreatedAt    time.Time `json:"created_at"`
}

// Request/Response types

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateUserRequest struct {
	Username           string    `json:"username" binding:"required"`
	Password           string    `json:"password" binding:"required"`
	FullName           string    `json:"full_name"`
	Email              string    `json:"email"`
	Phone              string    `json:"phone"`
	Role               string    `json:"role" binding:"required"`
	UniversityID       *uint     `json:"university_id"`
	AssignedCategories JSONArray `json:"assigned_categories"`
}

type UpdateUserRequest struct {
	FullName           string    `json:"full_name"`
	Email              string    `json:"email"`
	Phone              string    `json:"phone"`
	Role               string    `json:"role"`
	UniversityID       *uint     `json:"university_id"`
	AssignedCategories JSONArray `json:"assigned_categories"`
	Password           string    `json:"password"`
}

type CreateSubmissionRequest struct {
	AcademicYearID   uint   `json:"academic_year_id" binding:"required"`
	AuthorizedPerson string `json:"authorized_person"`
	AuthorizedPhone  string `json:"authorized_phone"`
	AuthorizedEmail  string `json:"authorized_email"`
}

type UpdateSubmissionRequest struct {
	AuthorizedPerson string           `json:"authorized_person"`
	AuthorizedPhone  string           `json:"authorized_phone"`
	AuthorizedEmail  string           `json:"authorized_email"`
	Items            []SubmissionItemInput `json:"items"`
}

type SubmissionItemInput struct {
	CriteriaID   uint   `json:"criteria_id"`
	Evidence     string `json:"evidence"`
	EvidenceFile string `json:"evidence_file"`
}

type ReviewInput struct {
	CategoryID uint              `json:"category_id" binding:"required"`
	Items      []ReviewItemInput `json:"items"`
	Comments   string            `json:"comments"`
}

type ReviewItemInput struct {
	SubmissionItemID uint    `json:"submission_item_id" binding:"required"`
	Score            float64 `json:"score"`
	AdminComment     string  `json:"admin_comment"`
	Status           string  `json:"status"`
}

type AssignCategoriesRequest struct {
	CategoryIDs JSONArray `json:"category_ids" binding:"required"`
}

type AcademicYearRequest struct {
	Name               string `json:"name" binding:"required"`
	StartDate          string `json:"start_date" binding:"required"`
	EndDate            string `json:"end_date" binding:"required"`
	SubmissionDeadline string `json:"submission_deadline"`
	IsActive           bool   `json:"is_active"`
}

type AIAnalysisRequest struct {
	Provider string `json:"provider"` // deepseek or gemini
}

type AICompareRequest struct {
	UniversityIDs []uint `json:"university_ids" binding:"required"`
	AcademicYearID uint  `json:"academic_year_id" binding:"required"`
	Provider       string `json:"provider"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// System Settings
type SystemSetting struct {
	Key   string `json:"key" gorm:"primaryKey"`
	Value string `json:"value"`
}

type SystemSettingsResponse struct {
	SiteTitle       string `json:"site_title"`
	SiteDescription string `json:"site_description"`
	SubmissionsOpen bool   `json:"submissions_open"`
	DeepSeekAPIKey  string `json:"deepseek_api_key"`
	DeepSeekURL     string `json:"deepseek_url"`
	GeminiAPIKey    string `json:"gemini_api_key"`
	GeminiURL       string `json:"gemini_url"`
	HasDeepSeekKey       bool   `json:"has_deepseek_key"`
	HasGeminiKey         bool   `json:"has_gemini_key"`
	MaxLoginAttempts     string `json:"max_login_attempts"`
	BlockDurationMinutes string `json:"block_duration_minutes"`
	MaxFileSizeMB        string `json:"max_file_size_mb"`
}

type UpdateSettingsRequest struct {
	SiteTitle            *string `json:"site_title"`
	SiteDescription      *string `json:"site_description"`
	SubmissionsOpen      *bool   `json:"submissions_open"`
	DeepSeekAPIKey       *string `json:"deepseek_api_key"`
	DeepSeekURL          *string `json:"deepseek_url"`
	GeminiAPIKey         *string `json:"gemini_api_key"`
	GeminiURL            *string `json:"gemini_url"`
	MaxLoginAttempts     *string `json:"max_login_attempts"`
	BlockDurationMinutes *string `json:"block_duration_minutes"`
	MaxFileSizeMB        *string `json:"max_file_size_mb"`
}

type TestAIRequest struct {
	Provider string `json:"provider" binding:"required"`
	APIKey   string `json:"api_key" binding:"required"`
	BaseURL  string `json:"base_url"`
}
