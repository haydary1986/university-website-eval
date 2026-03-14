package database

import (
	"log"

	"website-eval-system/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(dbPath string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate
	err = DB.AutoMigrate(
		&models.User{},
		&models.University{},
		&models.AcademicYear{},
		&models.Category{},
		&models.Criteria{},
		&models.Submission{},
		&models.SubmissionItem{},
		&models.Review{},
		&models.AuditLog{},
		&models.LoginAttempt{},
		&models.BlockedIP{},
		&models.ActiveSession{},
		&models.SystemSetting{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database initialized successfully")
}
