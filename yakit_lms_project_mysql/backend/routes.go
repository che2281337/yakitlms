package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool { return true },
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:   []string{"Content-Length"},
		MaxAge:          12 * time.Hour,
	}))

	uploadDir := getenv("UPLOAD_DIR", "./uploads")
	_ = os.MkdirAll(uploadDir, 0755)
	r.Static("/uploads", uploadDir)

	frontendDir := getenv("FRONTEND_DIR", "../frontend")
	if _, err := os.Stat(frontendDir); err == nil {
		r.GET("/", func(c *gin.Context) { c.File(filepath.Join(frontendDir, "college-site.html")) })
		r.GET("/site", func(c *gin.Context) { c.File(filepath.Join(frontendDir, "college-site.html")) })
		r.GET("/college-site.html", func(c *gin.Context) { c.File(filepath.Join(frontendDir, "college-site.html")) })
		r.GET("/lms", func(c *gin.Context) { c.File(filepath.Join(frontendDir, "lms.html")) })
		r.GET("/lms.html", func(c *gin.Context) { c.File(filepath.Join(frontendDir, "lms.html")) })
	}

	r.GET("/api/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/api/public/news", listNews)
	r.GET("/api/public/news/:id", getNews)
	r.GET("/api/public/teachers", listTeachers)
	r.POST("/api/public/applications", createApplication)
	r.POST("/api/auth/register", register)
	r.POST("/api/auth/login", login)

	api := r.Group("/api", authMiddleware())
	api.GET("/me", me)
	api.GET("/lms/dashboard", dashboard)

	api.GET("/courses", listCourses)
	api.POST("/courses", requireRoles("admin"), createCourse)
	api.GET("/courses/:id", getCourse)
	api.PUT("/courses/:id", requireRoles("teacher", "admin"), updateCourse)
	api.DELETE("/courses/:id", requireRoles("admin"), deleteCourse)
	api.GET("/courses/:id/enrollments", requireRoles("teacher", "admin"), listEnrollments)
	api.POST("/courses/:id/enrollments", requireRoles("admin"), addEnrollment)
	api.DELETE("/courses/:id/enrollments/:userId", requireRoles("admin"), removeEnrollment)

	api.POST("/courses/:id/modules", requireRoles("teacher", "admin"), createModule)
	api.PUT("/modules/:id", requireRoles("teacher", "admin"), updateModule)
	api.DELETE("/modules/:id", requireRoles("teacher", "admin"), deleteModule)

	api.POST("/modules/:id/materials", requireRoles("teacher", "admin"), createMaterial)
	api.PUT("/materials/:id", requireRoles("teacher", "admin"), updateMaterial)
	api.DELETE("/materials/:id", requireRoles("teacher", "admin"), deleteMaterial)

	api.POST("/modules/:id/assignments", requireRoles("teacher", "admin"), createAssignment)
	api.PUT("/assignments/:id", requireRoles("teacher", "admin"), updateAssignment)
	api.DELETE("/assignments/:id", requireRoles("teacher", "admin"), deleteAssignment)
	api.GET("/assignments/:id/submissions", requireRoles("teacher", "admin"), listSubmissions)
	api.POST("/assignments/:id/submissions", requireRoles("student"), createSubmission)
	api.PUT("/submissions/:id", requireRoles("student"), updateSubmission)
	api.POST("/submissions/:id/files", requireRoles("student"), addSubmissionFiles)
	api.DELETE("/submissions/:id/file", requireRoles("student"), deleteAllSubmissionFiles)
	api.DELETE("/submissions/:id/files/:fileId", requireRoles("student"), deleteSubmissionFile)
	api.GET("/submissions", requireRoles("teacher", "admin"), listSubmissionsForTeacher)
	api.PUT("/submissions/:id/grade", requireRoles("teacher", "admin"), gradeSubmission)

	api.POST("/modules/:id/tests", requireRoles("teacher", "admin"), createTest)
	api.GET("/tests/:id", getTest)
	api.PUT("/tests/:id", requireRoles("teacher", "admin"), updateTest)
	api.DELETE("/tests/:id", requireRoles("teacher", "admin"), deleteTest)
	api.POST("/tests/:id/questions", requireRoles("teacher", "admin"), createQuestion)
	api.PUT("/questions/:id", requireRoles("teacher", "admin"), updateQuestion)
	api.DELETE("/questions/:id", requireRoles("teacher", "admin"), deleteQuestion)
	api.POST("/tests/:id/attempts", requireRoles("student"), createTestAttempt)
	api.GET("/tests/:id/attempts", requireRoles("teacher", "admin"), listTestAttempts)
	api.PUT("/test-attempts/:id/grade", requireRoles("teacher", "admin"), gradeTestAttempt)

	api.GET("/grades", listGrades)
	api.GET("/schedule", listSchedule)

	api.GET("/admin/users", requireRoles("admin"), listUsers)
	api.POST("/admin/users", requireRoles("admin"), createUser)
	api.PUT("/admin/users/:id", requireRoles("admin"), updateUser)
	api.DELETE("/admin/users/:id", requireRoles("admin"), deleteUser)
	api.GET("/admin/teachers", requireRoles("admin"), listTeacherUsers)
	api.GET("/admin/applications", requireRoles("admin"), listApplications)
	api.PUT("/admin/applications/:id", requireRoles("admin"), updateApplication)
	api.DELETE("/admin/applications/:id", requireRoles("admin"), deleteApplication)
	api.POST("/admin/news", requireRoles("admin"), createNews)
	api.PUT("/admin/news/:id", requireRoles("admin"), updateNews)
	api.DELETE("/admin/news/:id", requireRoles("admin"), deleteNews)

	return r
}
