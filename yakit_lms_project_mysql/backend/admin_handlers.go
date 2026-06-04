package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func listUsers(c *gin.Context) {
	var users []User
	db.Order("role asc, full_name asc").Find(&users)
	result := make([]gin.H, 0, len(users))
	for _, u := range users {
		result = append(result, publicUser(u))
	}
	c.JSON(http.StatusOK, result)
}

func listTeacherUsers(c *gin.Context) {
	var users []User
	db.Where("role = ?", "teacher").Order("full_name asc").Find(&users)
	result := make([]gin.H, 0, len(users))
	for _, u := range users {
		result = append(result, publicUser(u))
	}
	c.JSON(http.StatusOK, result)
}

func createUser(c *gin.Context) {
	var req struct {
		FullName  string `json:"fullName"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		Role      string `json:"role"`
		GroupName string `json:"groupName"`
		Specialty string `json:"specialty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.FullName) == "" || strings.TrimSpace(req.Email) == "" || !validateRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fullName, email and valid role are required"})
		return
	}
	if req.Password == "" {
		req.Password = "123456"
	}
	hash, err := hashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	u := User{FullName: req.FullName, Email: strings.ToLower(req.Email), PasswordHash: hash, Role: req.Role, GroupName: req.GroupName, Specialty: req.Specialty}
	if err := db.Create(&u).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}
	c.JSON(http.StatusCreated, publicUser(u))
}

func updateUser(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var u User
	if err := db.First(&u, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	var req struct {
		FullName  string `json:"fullName"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		Role      string `json:"role"`
		GroupName string `json:"groupName"`
		Specialty string `json:"specialty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.FullName) != "" {
		u.FullName = req.FullName
	}
	if strings.TrimSpace(req.Email) != "" {
		u.Email = strings.ToLower(req.Email)
	}
	if req.Role != "" {
		if !validateRole(req.Role) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
			return
		}
		u.Role = req.Role
	}
	u.GroupName = req.GroupName
	u.Specialty = req.Specialty
	if req.Password != "" {
		hash, err := hashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		u.PasswordHash = hash
	}
	if err := db.Save(&u).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, publicUser(u))
}

func deleteUser(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	current, _ := currentUser(c)
	if current != nil && current.ID == id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you cannot delete yourself"})
		return
	}
	db.Where("user_id = ?", id).Delete(&Grade{})
	var submissionIDs []uint
	db.Model(&Submission{}).Where("student_id = ?", id).Pluck("id", &submissionIDs)
	if len(submissionIDs) > 0 {
		db.Where("submission_id IN ?", submissionIDs).Delete(&SubmissionFile{})
	}
	db.Where("student_id = ?", id).Delete(&Submission{})
	db.Where("student_id = ?", id).Delete(&TestAttempt{})
	db.Where("user_id = ?", id).Delete(&Enrollment{})
	if err := db.Delete(&User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func listApplications(c *gin.Context) {
	var apps []Application
	db.Order("created_at desc").Find(&apps)
	c.JSON(http.StatusOK, apps)
}

func updateApplication(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var app Application
	if err := db.First(&app, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Status) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}
	app.Status = req.Status
	db.Save(&app)
	c.JSON(http.StatusOK, app)
}

func deleteApplication(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := db.Delete(&Application{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func parseNewsPayload(c *gin.Context) (News, error) {
	news := News{PublishedAt: time.Now()}
	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		news.Title = c.PostForm("title")
		news.Excerpt = c.PostForm("excerpt")
		news.Content = c.PostForm("content")
		news.ImageURL = c.PostForm("imageUrl")
		if file, err := c.FormFile("image"); err == nil {
			url, err := saveUploadedFile(c, file, "news")
			if err != nil {
				return news, err
			}
			news.ImageURL = url
		}
		return news, nil
	}
	var req struct {
		Title    string `json:"title"`
		Excerpt  string `json:"excerpt"`
		Content  string `json:"content"`
		ImageURL string `json:"imageUrl"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return news, err
	}
	news.Title = req.Title
	news.Excerpt = req.Excerpt
	news.Content = req.Content
	news.ImageURL = req.ImageURL
	return news, nil
}

func saveUploadedFile(c *gin.Context, file *multipart.FileHeader, subdir string) (string, error) {
	uploadDir := filepath.Join(getenv("UPLOAD_DIR", "./uploads"), subdir)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	path := filepath.Join(uploadDir, name)
	if err := c.SaveUploadedFile(file, path); err != nil {
		return "", err
	}
	return "/uploads/" + subdir + "/" + name, nil
}

func createNews(c *gin.Context) {
	news, err := parseNewsPayload(c)
	if err != nil || strings.TrimSpace(news.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	if err := db.Create(&news).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, news)
}

func updateNews(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var existing News
	if err := db.First(&existing, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "news not found"})
		return
	}
	news, err := parseNewsPayload(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(news.Title) != "" {
		existing.Title = news.Title
	}
	existing.Excerpt = news.Excerpt
	existing.Content = news.Content
	if news.ImageURL != "" {
		existing.ImageURL = news.ImageURL
	}
	if err := db.Save(&existing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, existing)
}

func deleteNews(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := db.Delete(&News{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
