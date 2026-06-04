package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type submissionInput struct {
	Answer string
	Files  []SubmissionFile
}

func submissionPayload(c *gin.Context) (submissionInput, error) {
	input := submissionInput{Files: []SubmissionFile{}}
	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		input.Answer = c.PostForm("answer")
		if url := strings.TrimSpace(c.PostForm("fileUrl")); url != "" {
			input.Files = append(input.Files, SubmissionFile{FileURL: url, FileName: c.PostForm("fileName")})
		}
		form, _ := c.MultipartForm()
		if form != nil {
			for _, field := range []string{"files", "file"} {
				for _, file := range form.File[field] {
					url, err := saveUploadedFile(c, file, "submissions")
					if err != nil {
						return input, err
					}
					input.Files = append(input.Files, SubmissionFile{FileURL: url, FileName: file.Filename})
				}
			}
		}
		return input, nil
	}
	var req struct {
		Answer   string `json:"answer"`
		FileURL  string `json:"fileUrl"`
		FileName string `json:"fileName"`
		Files    []struct {
			FileURL  string `json:"fileUrl"`
			FileName string `json:"fileName"`
		} `json:"files"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return input, err
	}
	input.Answer = req.Answer
	if strings.TrimSpace(req.FileURL) != "" {
		input.Files = append(input.Files, SubmissionFile{FileURL: req.FileURL, FileName: req.FileName})
	}
	for _, f := range req.Files {
		if strings.TrimSpace(f.FileURL) != "" {
			input.Files = append(input.Files, SubmissionFile{FileURL: f.FileURL, FileName: f.FileName})
		}
	}
	return input, nil
}

func addFilesToSubmission(submissionID uint, files []SubmissionFile) error {
	for i := range files {
		files[i].SubmissionID = submissionID
		if err := db.Create(&files[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func countSubmissionFiles(submissionID uint) int64 {
	var count int64
	db.Model(&SubmissionFile{}).Where("submission_id = ?", submissionID).Count(&count)
	return count
}

func resetSubmissionForRecheck(sub *Submission) {
	sub.Status = "pending"
	sub.Grade = nil
	sub.Comment = ""
}

func createSubmission(c *gin.Context) {
	assignmentID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u, ok := currentUser(c)
	if !ok {
		return
	}
	var assignment Assignment
	if err := db.First(&assignment, assignmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
		return
	}
	if !studentHasCourse(u.ID, assignment.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "student is not enrolled to this course"})
		return
	}
	input, err := submissionPayload(c)
	if err != nil || (strings.TrimSpace(input.Answer) == "" && len(input.Files) == 0) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "answer text or at least one attached file is required"})
		return
	}
	var sub Submission
	err = db.Where("assignment_id = ? AND student_id = ?", assignmentID, u.ID).First(&sub).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		sub = Submission{AssignmentID: assignmentID, StudentID: u.ID, Answer: input.Answer, Status: "pending"}
		err = db.Create(&sub).Error
	} else if err == nil {
		sub.Answer = input.Answer
		resetSubmissionForRecheck(&sub)
		err = db.Save(&sub).Error
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := addFilesToSubmission(sub.ID, input.Files); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	db.Where("user_id = ? AND assignment_id = ?", u.ID, assignmentID).Delete(&Grade{})
	db.Preload("Files").First(&sub, sub.ID)
	c.JSON(http.StatusCreated, sub)
}

func updateSubmission(c *gin.Context) {
	submissionID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u, ok := currentUser(c)
	if !ok {
		return
	}
	var sub Submission
	if err := db.Preload("Assignment").Preload("Files").First(&sub, submissionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
		return
	}
	if sub.StudentID != u.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can edit only your own submission"})
		return
	}
	input, err := submissionPayload(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(input.Answer) == "" && len(input.Files) == 0 && countSubmissionFiles(sub.ID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "answer text or at least one attached file is required"})
		return
	}
	sub.Answer = input.Answer
	resetSubmissionForRecheck(&sub)
	if err := db.Save(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := addFilesToSubmission(sub.ID, input.Files); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	db.Where("user_id = ? AND assignment_id = ?", u.ID, sub.AssignmentID).Delete(&Grade{})
	db.Preload("Files").First(&sub, sub.ID)
	c.JSON(http.StatusOK, sub)
}

func addSubmissionFiles(c *gin.Context) {
	submissionID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u, ok := currentUser(c)
	if !ok {
		return
	}
	var sub Submission
	if err := db.Preload("Assignment").First(&sub, submissionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
		return
	}
	if sub.StudentID != u.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can edit only your own submission files"})
		return
	}
	input, err := submissionPayload(c)
	if err != nil || len(input.Files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one file is required"})
		return
	}
	if err := addFilesToSubmission(sub.ID, input.Files); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resetSubmissionForRecheck(&sub)
	if err := db.Save(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	db.Where("user_id = ? AND assignment_id = ?", u.ID, sub.AssignmentID).Delete(&Grade{})
	db.Preload("Files").First(&sub, sub.ID)
	c.JSON(http.StatusOK, sub)
}

func deleteAllSubmissionFiles(c *gin.Context) {
	submissionID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u, ok := currentUser(c)
	if !ok {
		return
	}
	var sub Submission
	if err := db.First(&sub, submissionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
		return
	}
	if sub.StudentID != u.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can delete only your own submission files"})
		return
	}
	if strings.TrimSpace(sub.Answer) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete all files without answer text"})
		return
	}
	db.Where("submission_id = ?", sub.ID).Delete(&SubmissionFile{})
	sub.FileURL = ""
	sub.FileName = ""
	resetSubmissionForRecheck(&sub)
	if err := db.Save(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	db.Where("user_id = ? AND assignment_id = ?", u.ID, sub.AssignmentID).Delete(&Grade{})
	db.Preload("Files").First(&sub, sub.ID)
	c.JSON(http.StatusOK, sub)
}

func deleteSubmissionFile(c *gin.Context) {
	submissionID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	fileID, ok := parseUintParam(c, "fileId")
	if !ok {
		return
	}
	u, ok := currentUser(c)
	if !ok {
		return
	}
	var sub Submission
	if err := db.First(&sub, submissionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
		return
	}
	if sub.StudentID != u.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can delete only your own submission file"})
		return
	}
	var file SubmissionFile
	if err := db.Where("id = ? AND submission_id = ?", fileID, sub.ID).First(&file).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	if countSubmissionFiles(sub.ID) <= 1 && strings.TrimSpace(sub.Answer) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete the only answer content; add text first"})
		return
	}
	if err := db.Delete(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resetSubmissionForRecheck(&sub)
	if err := db.Save(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	db.Where("user_id = ? AND assignment_id = ?", u.ID, sub.AssignmentID).Delete(&Grade{})
	db.Preload("Files").First(&sub, sub.ID)
	c.JSON(http.StatusOK, sub)
}

func listSubmissions(c *gin.Context) {
	assignmentID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var assignment Assignment
	if err := db.First(&assignment, assignmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
		return
	}
	if !canManageCourse(c, assignment.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this assignment"})
		return
	}
	var subs []Submission
	db.Preload("Student").Preload("Files").Preload("Assignment").Where("assignment_id = ?", assignmentID).Order("created_at desc").Find(&subs)
	c.JSON(http.StatusOK, subs)
}

func listSubmissionsForTeacher(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	query := db.Preload("Student").Preload("Files").Preload("Assignment").Preload("Assignment.Course").Preload("Assignment.Module").Order("updated_at desc")
	if u.Role == "teacher" {
		ids := courseIDsForTeacher(u.ID)
		if len(ids) == 0 {
			c.JSON(http.StatusOK, []Submission{})
			return
		}
		query = query.Joins("JOIN assignments ON assignments.id = submissions.assignment_id").Where("assignments.course_id IN ?", ids)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("submissions.status = ?", status)
	}
	var subs []Submission
	query.Find(&subs)
	c.JSON(http.StatusOK, subs)
}

func gradeSubmission(c *gin.Context) {
	submissionID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var sub Submission
	if err := db.Preload("Assignment").Preload("Files").First(&sub, submissionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
		return
	}
	if !canManageCourse(c, sub.Assignment.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot grade this submission"})
		return
	}
	var req struct {
		Grade   int    `json:"grade"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Grade < 1 || req.Grade > sub.Assignment.MaxScore {
		c.JSON(http.StatusBadRequest, gin.H{"error": "grade must be between 1 and assignment maxScore"})
		return
	}
	sub.Grade = &req.Grade
	sub.Comment = req.Comment
	sub.Status = "graded"
	if err := db.Save(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	assignmentID := sub.AssignmentID
	grade := Grade{UserID: sub.StudentID, CourseID: sub.Assignment.CourseID, Kind: "assignment", AssignmentID: &assignmentID}
	updates := Grade{CourseID: sub.Assignment.CourseID, Kind: "assignment", AssignmentID: &assignmentID, Value: req.Grade, MaxValue: sub.Assignment.MaxScore, Comment: req.Comment}
	if err := db.Where(&Grade{UserID: sub.StudentID, Kind: "assignment", AssignmentID: &assignmentID}).Assign(updates).FirstOrCreate(&grade).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"submission": sub, "grade": grade})
}
