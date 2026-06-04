package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func listNews(c *gin.Context) {
	var news []News
	db.Order("published_at desc, id desc").Find(&news)
	c.JSON(http.StatusOK, news)
}

func getNews(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var news News
	if err := db.First(&news, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "news not found"})
		return
	}
	c.JSON(http.StatusOK, news)
}

func listTeachers(c *gin.Context) {
	var teachers []TeacherProfile
	db.Order("id asc").Find(&teachers)
	c.JSON(http.StatusOK, teachers)
}

func createApplication(c *gin.Context) {
	var req struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Phone     string `json:"phone"`
		Direction string `json:"direction"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.FirstName) == "" || strings.TrimSpace(req.LastName) == "" || strings.TrimSpace(req.Phone) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "firstName, lastName and phone are required"})
		return
	}
	app := Application{FirstName: req.FirstName, LastName: req.LastName, Phone: req.Phone, Direction: req.Direction, Status: "new"}
	if err := db.Create(&app).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, app)
}
