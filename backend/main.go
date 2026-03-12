package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"website-eval-system/config"
	"website-eval-system/database"
	"website-eval-system/handlers"
	"website-eval-system/middleware"
	"website-eval-system/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Ensure upload directory exists
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatal("Failed to create upload directory:", err)
	}

	// Initialize database
	database.Init(cfg.DBPath)
	database.Seed()

	// Initialize services
	aiService := services.NewAIService(cfg)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg)
	universityHandler := handlers.NewUniversityHandler()
	academicYearHandler := handlers.NewAcademicYearHandler()
	categoryHandler := handlers.NewCategoryHandler()
	submissionHandler := handlers.NewSubmissionHandler(cfg)
	adminHandler := handlers.NewAdminHandler()
	statsHandler := handlers.NewStatsHandler()
	aiHandler := handlers.NewAIHandler(aiService)

	// Setup router
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Serve uploaded files
	r.Static("/uploads", cfg.UploadDir)

	// Health check (for Docker)
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
		}
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthRequired(cfg))
	{
		// Auth
		protected.GET("/auth/me", authHandler.Me)
		protected.POST("/auth/register", authHandler.Register)

		// Universities
		protected.GET("/universities", universityHandler.List)
		protected.GET("/universities/:id", universityHandler.Get)
		protected.PUT("/universities/:id", universityHandler.Update)

		// Academic Years
		protected.GET("/academic-years", academicYearHandler.List)
		protected.POST("/academic-years", middleware.RoleRequired("super_admin"), academicYearHandler.Create)
		protected.PUT("/academic-years/:id", middleware.RoleRequired("super_admin"), academicYearHandler.Update)

		// Categories
		protected.GET("/categories", categoryHandler.List)
		protected.GET("/categories/:id", categoryHandler.Get)

		// Submissions
		protected.GET("/submissions", submissionHandler.List)
		protected.GET("/submissions/:id", submissionHandler.Get)
		protected.POST("/submissions", middleware.RoleRequired("university"), submissionHandler.Create)
		protected.PUT("/submissions/:id", middleware.RoleRequired("university"), submissionHandler.Update)
		protected.POST("/submissions/:id/submit", middleware.RoleRequired("university"), submissionHandler.Submit)
		protected.GET("/submissions/:id/diff/:version", submissionHandler.Diff)
		protected.POST("/submissions/upload", submissionHandler.UploadFile)
		protected.POST("/upload", submissionHandler.UploadFile)

		// Admin review routes
		admin := protected.Group("/admin")
		admin.Use(middleware.RoleRequired("super_admin", "admin"))
		{
			// Submissions review
			admin.GET("/submissions", adminHandler.ListSubmissions)
			admin.GET("/submissions/:id", adminHandler.GetSubmission)
			admin.POST("/submissions/:id/review", adminHandler.ReviewSubmission)
			admin.PUT("/submissions/:id/approve", adminHandler.ApproveSubmission)
			admin.PUT("/submissions/:id/reject", adminHandler.RejectSubmission)

			// User management (super_admin only)
			admin.GET("/users", middleware.RoleRequired("super_admin"), adminHandler.ListUsers)
			admin.POST("/users", middleware.RoleRequired("super_admin"), adminHandler.CreateUser)
			admin.PUT("/users/:id", middleware.RoleRequired("super_admin"), adminHandler.UpdateUser)
			admin.DELETE("/users/:id", middleware.RoleRequired("super_admin"), adminHandler.DeleteUser)
			admin.PUT("/users/:id/assign-categories", middleware.RoleRequired("super_admin"), adminHandler.AssignCategories)
		}

		// Statistics
		stats := protected.Group("/stats")
		{
			stats.GET("/overview", statsHandler.Overview)
			stats.GET("/universities", statsHandler.Universities)
			stats.GET("/categories", statsHandler.Categories)
			stats.GET("/comparison/:universityId", statsHandler.Comparison)
		}

		// AI
		ai := protected.Group("/ai")
		ai.Use(middleware.RoleRequired("super_admin", "admin"))
		{
			ai.POST("/analyze-submission/:id", aiHandler.AnalyzeSubmission)
			ai.POST("/suggest-improvements/:id", aiHandler.SuggestImprovements)
			ai.POST("/compare-universities", aiHandler.CompareUniversities)
		}
	}

	// Serve frontend static files (for production single-container deployment)
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./static"
	}
	if _, err := os.Stat(staticDir); err == nil {
		staticFS := http.Dir(staticDir)
		fileServer := http.FileServer(staticFS)

		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path

			// If it's an API route, return 404
			if strings.HasPrefix(path, "/api/") {
				c.JSON(404, gin.H{"error": "not found"})
				return
			}

			// Try to serve the file directly
			if f, err := fs.Stat(os.DirFS(staticDir), strings.TrimPrefix(path, "/")); err == nil && !f.IsDir() {
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}

			// SPA fallback: serve index.html for all other routes
			c.Request.URL.Path = "/"
			fileServer.ServeHTTP(c.Writer, c.Request)
		})
		log.Printf("Serving frontend from %s", staticDir)
	}

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
