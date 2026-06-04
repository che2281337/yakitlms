package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func createModule(c *gin.Context) {
	courseID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if !canManageCourse(c, courseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this course"})
		return
	}
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		SortOrder   int    `json:"sortOrder"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	if req.SortOrder == 0 {
		var count int64
		db.Model(&CourseModule{}).Where("course_id = ?", courseID).Count(&count)
		req.SortOrder = int(count) + 1
	}
	module := CourseModule{CourseID: courseID, Title: req.Title, Description: req.Description, SortOrder: req.SortOrder}
	if err := db.Create(&module).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, module)
}

func updateModule(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var module CourseModule
	if err := db.First(&module, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
		return
	}
	if !canManageCourse(c, module.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this module"})
		return
	}
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		SortOrder   int    `json:"sortOrder"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Title) != "" {
		module.Title = req.Title
	}
	module.Description = req.Description
	if req.SortOrder != 0 {
		module.SortOrder = req.SortOrder
	}
	if err := db.Save(&module).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, module)
}

func deleteModule(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var module CourseModule
	if err := db.First(&module, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
		return
	}
	if !canManageCourse(c, module.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this module"})
		return
	}
	var assignments []Assignment
	db.Where("module_id = ?", id).Find(&assignments)
	for _, a := range assignments {
		deleteAssignmentData(a.ID)
	}
	var tests []Test
	db.Where("module_id = ?", id).Find(&tests)
	for _, t := range tests {
		deleteTestData(t.ID)
	}
	db.Where("module_id = ?", id).Delete(&Material{})
	db.Delete(&module)
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
