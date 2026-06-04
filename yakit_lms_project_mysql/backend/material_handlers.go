package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func createMaterial(c *gin.Context) {
	moduleID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	courseID, exists := moduleCourseID(moduleID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
		return
	}
	if !canManageCourse(c, courseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this course"})
		return
	}

	var req struct {
		Title     string `json:"title"`
		Kind      string `json:"kind"`
		URL       string `json:"url"`
		Content   string `json:"content"`
		SortOrder int    `json:"sortOrder"`
	}
	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		req.Title = c.PostForm("title")
		req.Kind = c.PostForm("kind")
		req.URL = c.PostForm("url")
		req.Content = c.PostForm("content")
		req.SortOrder, _ = strconv.Atoi(c.PostForm("sortOrder"))
		if file, err := c.FormFile("file"); err == nil {
			url, err := saveUploadedFile(c, file, "materials")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			req.URL = url
			req.Kind = "file"
		}
	} else if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	if req.Kind == "" {
		req.Kind = "link"
	}
	if req.SortOrder == 0 {
		req.SortOrder = 1
	}
	material := Material{CourseID: courseID, ModuleID: moduleID, Title: req.Title, Kind: req.Kind, URL: req.URL, Content: req.Content, SortOrder: req.SortOrder}
	if err := db.Create(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, material)
}

func updateMaterial(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var material Material
	if err := db.First(&material, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "material not found"})
		return
	}
	if !canManageCourse(c, material.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this material"})
		return
	}
	var req struct {
		Title     string `json:"title"`
		Kind      string `json:"kind"`
		URL       string `json:"url"`
		Content   string `json:"content"`
		SortOrder int    `json:"sortOrder"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Title) != "" {
		material.Title = req.Title
	}
	if req.Kind != "" {
		material.Kind = req.Kind
	}
	material.URL = req.URL
	material.Content = req.Content
	if req.SortOrder != 0 {
		material.SortOrder = req.SortOrder
	}
	if err := db.Save(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, material)
}

func deleteMaterial(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var material Material
	if err := db.First(&material, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "material not found"})
		return
	}
	if !canManageCourse(c, material.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this material"})
		return
	}
	db.Delete(&material)
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
