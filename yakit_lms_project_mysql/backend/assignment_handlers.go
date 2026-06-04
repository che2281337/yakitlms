package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func createAssignment(c *gin.Context) {
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
		Title          string `json:"title"`
		Description    string `json:"description"`
		InstructionURL string `json:"instructionUrl"`
		Kind           string `json:"kind"`
		Deadline       string `json:"deadline"`
		MaxScore       int    `json:"maxScore"`
	}
	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		req.Title = c.PostForm("title")
		req.Description = c.PostForm("description")
		req.InstructionURL = c.PostForm("instructionUrl")
		req.Kind = c.PostForm("kind")
		req.Deadline = c.PostForm("deadline")
		req.MaxScore, _ = strconv.Atoi(c.PostForm("maxScore"))
		if file, err := c.FormFile("file"); err == nil {
			url, err := saveUploadedFile(c, file, "assignments")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			req.InstructionURL = url
		}
	} else if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	deadline, err := parseDeadline(req.Deadline, 7)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.MaxScore == 0 {
		req.MaxScore = 5
	}
	if req.Kind == "" {
		req.Kind = "assignment"
	}
	assignment := Assignment{CourseID: courseID, ModuleID: moduleID, Title: req.Title, Description: req.Description, InstructionURL: req.InstructionURL, Kind: req.Kind, Deadline: deadline, MaxScore: req.MaxScore}
	if err := db.Create(&assignment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, assignment)
}

func updateAssignment(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var assignment Assignment
	if err := db.First(&assignment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
		return
	}
	if !canManageCourse(c, assignment.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this assignment"})
		return
	}
	var req struct {
		Title          string `json:"title"`
		Description    string `json:"description"`
		InstructionURL string `json:"instructionUrl"`
		Kind           string `json:"kind"`
		Deadline       string `json:"deadline"`
		MaxScore       int    `json:"maxScore"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Title) != "" {
		assignment.Title = req.Title
	}
	assignment.Description = req.Description
	assignment.InstructionURL = req.InstructionURL
	if req.Kind != "" {
		assignment.Kind = req.Kind
	}
	if req.Deadline != "" {
		deadline, err := parseDeadline(req.Deadline, 7)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		assignment.Deadline = deadline
	}
	if req.MaxScore != 0 {
		assignment.MaxScore = req.MaxScore
	}
	if err := db.Save(&assignment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assignment)
}

func deleteAssignment(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var assignment Assignment
	if err := db.First(&assignment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
		return
	}
	if !canManageCourse(c, assignment.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this assignment"})
		return
	}
	deleteAssignmentData(id)
	db.Delete(&assignment)
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func deleteAssignmentData(assignmentID uint) {
	var submissionIDs []uint
	db.Model(&Submission{}).Where("assignment_id = ?", assignmentID).Pluck("id", &submissionIDs)
	if len(submissionIDs) > 0 {
		db.Where("submission_id IN ?", submissionIDs).Delete(&SubmissionFile{})
	}
	db.Where("assignment_id = ?", assignmentID).Delete(&Submission{})
	db.Where("assignment_id = ?", assignmentID).Delete(&Grade{})
}
