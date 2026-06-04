package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

func listCourses(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	var courses []Course
	query := db.Preload("Teacher").Order("id asc")
	if u.Role == "teacher" {
		query = query.Where("teacher_id = ?", u.ID)
	} else if u.Role == "student" {
		ids := courseIDsForStudent(u.ID)
		if len(ids) == 0 {
			c.JSON(http.StatusOK, []gin.H{})
			return
		}
		query = query.Where("id IN ?", ids)
	}
	query.Find(&courses)
	items := make([]gin.H, 0, len(courses))
	for _, course := range courses {
		items = append(items, courseListDTO(course, *u))
	}
	c.JSON(http.StatusOK, items)
}

func getCourse(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u, ok := currentUser(c)
	if !ok {
		return
	}
	if !canAccessCourse(*u, id) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot access this course"})
		return
	}
	var course Course
	err := db.Preload("Teacher").
		Preload("Modules", func(q *gorm.DB) *gorm.DB { return q.Order("sort_order asc, id asc") }).
		Preload("Modules.Materials", func(q *gorm.DB) *gorm.DB { return q.Order("sort_order asc, id asc") }).
		Preload("Modules.Assignments", func(q *gorm.DB) *gorm.DB { return q.Order("deadline asc, id asc") }).
		Preload("Modules.Tests", func(q *gorm.DB) *gorm.DB { return q.Order("deadline asc, id asc") }).
		Preload("Modules.Tests.Questions", func(q *gorm.DB) *gorm.DB { return q.Order("sort_order asc, id asc") }).
		First(&course, id).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
		return
	}
	c.JSON(http.StatusOK, courseDetailDTO(course, *u))
}

func createCourse(c *gin.Context) {
	var req struct {
		Title       string `json:"title"`
		Category    string `json:"category"`
		Description string `json:"description"`
		ImageURL    string `json:"imageUrl"`
		TeacherID   uint   `json:"teacherId"`
		Status      string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Title) == "" || req.TeacherID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title and teacherId are required"})
		return
	}
	var teacher User
	if err := db.Where("id = ? AND role = ?", req.TeacherID, "teacher").First(&teacher).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "teacherId must reference teacher"})
		return
	}
	if req.Status == "" {
		req.Status = "active"
	}
	course := Course{Title: req.Title, Category: req.Category, Description: req.Description, ImageURL: req.ImageURL, TeacherID: req.TeacherID, Status: req.Status}
	if err := db.Create(&course).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, course)
}

func updateCourse(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if !canManageCourse(c, id) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this course"})
		return
	}
	u, _ := currentUser(c)
	var course Course
	if err := db.First(&course, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
		return
	}
	var req struct {
		Title       string `json:"title"`
		Category    string `json:"category"`
		Description string `json:"description"`
		ImageURL    string `json:"imageUrl"`
		TeacherID   uint   `json:"teacherId"`
		Status      string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Title) != "" {
		course.Title = req.Title
	}
	course.Category = req.Category
	course.Description = req.Description
	course.ImageURL = req.ImageURL
	if req.Status != "" {
		course.Status = req.Status
	}
	if u.Role == "admin" && req.TeacherID != 0 {
		var teacher User
		if err := db.Where("id = ? AND role = ?", req.TeacherID, "teacher").First(&teacher).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "teacherId must reference teacher"})
			return
		}
		course.TeacherID = req.TeacherID
	}
	if err := db.Save(&course).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, course)
}

func deleteCourse(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	deleteCourseData(id)
	if err := db.Delete(&Course{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func deleteCourseData(courseID uint) {
	var assignments []Assignment
	db.Where("course_id = ?", courseID).Find(&assignments)
	for _, a := range assignments {
		deleteAssignmentData(a.ID)
	}
	var tests []Test
	db.Where("course_id = ?", courseID).Find(&tests)
	for _, t := range tests {
		deleteTestData(t.ID)
	}
	db.Where("course_id = ?", courseID).Delete(&Material{})
	db.Where("course_id = ?", courseID).Delete(&CourseModule{})
	db.Where("course_id = ?", courseID).Delete(&Enrollment{})
	db.Where("course_id = ?", courseID).Delete(&Grade{})
}

func listEnrollments(c *gin.Context) {
	courseID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if !canManageCourse(c, courseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this course"})
		return
	}
	var enrollments []Enrollment
	db.Preload("User").Where("course_id = ?", courseID).Order("id asc").Find(&enrollments)
	c.JSON(http.StatusOK, enrollments)
}

func addEnrollment(c *gin.Context) {
	courseID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		UserID   uint `json:"userId"`
		Progress int  `json:"progress"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}
	var user User
	if err := db.Where("id = ? AND role = ?", req.UserID, "student").First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId must reference student"})
		return
	}
	enrollment := Enrollment{UserID: req.UserID, CourseID: courseID, Progress: req.Progress}
	if err := db.Where(&Enrollment{UserID: req.UserID, CourseID: courseID}).Assign(Enrollment{Progress: req.Progress}).FirstOrCreate(&enrollment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, enrollment)
}

func removeEnrollment(c *gin.Context) {
	courseID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	userID, ok := parseUintParam(c, "userId")
	if !ok {
		return
	}
	db.Where("course_id = ? AND user_id = ?", courseID, userID).Delete(&Enrollment{})
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
