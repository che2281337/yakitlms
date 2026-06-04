package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
)

func createTest(c *gin.Context) {
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
		Title           string `json:"title"`
		Description     string `json:"description"`
		Deadline        string `json:"deadline"`
		MaxScore        int    `json:"maxScore"`
		AttemptLimit    int    `json:"attemptLimit"`
		DurationMinutes int    `json:"durationMinutes"`
		ShowResult      *bool  `json:"showResult"`
		Questions       []struct {
			Question       string   `json:"question"`
			Type           string   `json:"type"`
			ImageURL       string   `json:"imageUrl"`
			Options        []string `json:"options"`
			CorrectAnswer  string   `json:"correctAnswer"`
			CorrectAnswers []string `json:"correctAnswers"`
			Points         int      `json:"points"`
			SortOrder      int      `json:"sortOrder"`
		} `json:"questions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Title) == "" {
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
	if req.AttemptLimit <= 0 {
		req.AttemptLimit = 1
	}
	showResult := true
	if req.ShowResult != nil {
		showResult = *req.ShowResult
	}
	test := Test{CourseID: courseID, ModuleID: moduleID, Title: req.Title, Description: req.Description, Deadline: deadline, MaxScore: req.MaxScore, AttemptLimit: req.AttemptLimit, DurationMinutes: req.DurationMinutes, ShowResult: showResult}
	if err := db.Create(&test).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for i, q := range req.Questions {
		if strings.TrimSpace(q.Question) == "" {
			continue
		}
		optionsRaw, _ := json.Marshal(q.Options)
		if q.Points == 0 {
			q.Points = 1
		}
		if q.SortOrder == 0 {
			q.SortOrder = i + 1
		}
		qType := q.Type
		if qType == "" {
			qType = "radio"
		}
		answers := q.CorrectAnswers
		if len(answers) == 0 && q.CorrectAnswer != "" {
			answers = []string{q.CorrectAnswer}
		}
		answersRaw, _ := json.Marshal(answers)
		db.Create(&TestQuestion{TestID: test.ID, Question: q.Question, Type: qType, ImageURL: q.ImageURL, OptionsJSON: string(optionsRaw), CorrectAnswer: q.CorrectAnswer, CorrectAnswersJSON: string(answersRaw), Points: q.Points, SortOrder: q.SortOrder})
	}
	c.JSON(http.StatusCreated, test)
}

func getTest(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u, ok := currentUser(c)
	if !ok {
		return
	}
	var test Test
	if err := db.Preload("Course").Preload("Module").Preload("Questions", func(q *gorm.DB) *gorm.DB { return q.Order("sort_order asc, id asc") }).First(&test, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "test not found"})
		return
	}
	if !canAccessCourse(*u, test.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot access this test"})
		return
	}
	c.JSON(http.StatusOK, testDTO(test, *u))
}

func updateTest(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var test Test
	if err := db.First(&test, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "test not found"})
		return
	}
	if !canManageCourse(c, test.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this test"})
		return
	}
	var req struct {
		Title           string `json:"title"`
		Description     string `json:"description"`
		Deadline        string `json:"deadline"`
		MaxScore        int    `json:"maxScore"`
		AttemptLimit    int    `json:"attemptLimit"`
		DurationMinutes int    `json:"durationMinutes"`
		ShowResult      *bool  `json:"showResult"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Title) != "" {
		test.Title = req.Title
	}
	test.Description = req.Description
	if req.Deadline != "" {
		deadline, err := parseDeadline(req.Deadline, 7)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		test.Deadline = deadline
	}
	if req.MaxScore != 0 {
		test.MaxScore = req.MaxScore
	}
	if req.AttemptLimit > 0 {
		test.AttemptLimit = req.AttemptLimit
	}
	if req.DurationMinutes >= 0 {
		test.DurationMinutes = req.DurationMinutes
	}
	if req.ShowResult != nil {
		test.ShowResult = *req.ShowResult
	}
	if err := db.Save(&test).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, test)
}

func deleteTest(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var test Test
	if err := db.First(&test, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "test not found"})
		return
	}
	if !canManageCourse(c, test.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this test"})
		return
	}
	deleteTestData(id)
	db.Delete(&test)
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func deleteTestData(testID uint) {
	db.Where("test_id = ?", testID).Delete(&TestQuestion{})
	db.Where("test_id = ?", testID).Delete(&TestAttempt{})
	db.Where("test_id = ?", testID).Delete(&Grade{})
}

func createQuestion(c *gin.Context) {
	testID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var test Test
	if err := db.First(&test, testID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "test not found"})
		return
	}
	if !canManageCourse(c, test.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this test"})
		return
	}
	var req struct {
		Question       string   `json:"question"`
		Type           string   `json:"type"`
		ImageURL       string   `json:"imageUrl"`
		Options        []string `json:"options"`
		CorrectAnswer  string   `json:"correctAnswer"`
		CorrectAnswers []string `json:"correctAnswers"`
		Points         int      `json:"points"`
		SortOrder      int      `json:"sortOrder"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Question) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question is required"})
		return
	}
	if req.Points == 0 {
		req.Points = 1
	}
	if req.SortOrder == 0 {
		var count int64
		db.Model(&TestQuestion{}).Where("test_id = ?", testID).Count(&count)
		req.SortOrder = int(count) + 1
	}
	if req.Type == "" {
		req.Type = "radio"
	}
	answers := req.CorrectAnswers
	if len(answers) == 0 && req.CorrectAnswer != "" {
		answers = []string{req.CorrectAnswer}
	}
	optionsRaw, _ := json.Marshal(req.Options)
	answersRaw, _ := json.Marshal(answers)
	q := TestQuestion{TestID: testID, Question: req.Question, Type: req.Type, ImageURL: req.ImageURL, OptionsJSON: string(optionsRaw), CorrectAnswer: req.CorrectAnswer, CorrectAnswersJSON: string(answersRaw), Points: req.Points, SortOrder: req.SortOrder}
	if err := db.Create(&q).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, q)
}

func updateQuestion(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var q TestQuestion
	if err := db.First(&q, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}
	var test Test
	db.First(&test, q.TestID)
	if !canManageCourse(c, test.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this question"})
		return
	}
	var req struct {
		Question       string   `json:"question"`
		Type           string   `json:"type"`
		ImageURL       string   `json:"imageUrl"`
		Options        []string `json:"options"`
		CorrectAnswer  string   `json:"correctAnswer"`
		CorrectAnswers []string `json:"correctAnswers"`
		Points         int      `json:"points"`
		SortOrder      int      `json:"sortOrder"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Question) != "" {
		q.Question = req.Question
	}
	if req.Type != "" {
		q.Type = req.Type
	}
	q.ImageURL = req.ImageURL
	if req.Options != nil {
		optionsRaw, _ := json.Marshal(req.Options)
		q.OptionsJSON = string(optionsRaw)
	}
	if strings.TrimSpace(req.CorrectAnswer) != "" {
		q.CorrectAnswer = req.CorrectAnswer
	}
	if req.CorrectAnswers != nil {
		answersRaw, _ := json.Marshal(req.CorrectAnswers)
		q.CorrectAnswersJSON = string(answersRaw)
	}
	if req.Points != 0 {
		q.Points = req.Points
	}
	if req.SortOrder != 0 {
		q.SortOrder = req.SortOrder
	}
	if err := db.Save(&q).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, q)
}

func deleteQuestion(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var q TestQuestion
	if err := db.First(&q, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}
	var test Test
	db.First(&test, q.TestID)
	if !canManageCourse(c, test.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this question"})
		return
	}
	db.Delete(&q)
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func createTestAttempt(c *gin.Context) {
	testID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u, ok := currentUser(c)
	if !ok {
		return
	}
	var test Test
	if err := db.Preload("Questions").First(&test, testID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "test not found"})
		return
	}
	if !studentHasCourse(u.ID, test.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "student is not enrolled to this course"})
		return
	}
	var req struct {
		Answers map[string]interface{} `json:"answers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Answers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "answers are required"})
		return
	}
	attemptLimit := test.AttemptLimit
	if attemptLimit <= 0 {
		attemptLimit = 1
	}
	var used int64
	db.Model(&TestAttempt{}).Where("test_id = ? AND student_id = ?", testID, u.ID).Count(&used)
	if int(used) >= attemptLimit {
		response := gin.H{"error": "attempt limit exceeded", "attemptsUsed": used, "attemptLimit": attemptLimit}
		if test.ShowResult {
			var latest TestAttempt
			if err := db.Where("test_id = ? AND student_id = ?", testID, u.ID).Order("created_at desc").First(&latest).Error; err == nil {
				response["score"] = latest.Score
				response["maxScore"] = latest.MaxScore
				response["status"] = latest.Status
			}
		}
		c.JSON(http.StatusForbidden, response)
		return
	}
	score, maxScore := calculateTestScore(test.Questions, req.Answers, nil)
	answersRaw, _ := json.Marshal(req.Answers)
	attempt := TestAttempt{TestID: testID, StudentID: u.ID, Answers: string(answersRaw), AttemptNumber: int(used) + 1, Score: score, AutoScore: score, MaxScore: maxScore, Status: "completed"}
	if err := db.Create(&attempt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	testIDCopy := testID
	attemptID := attempt.ID
	grade := Grade{UserID: u.ID, CourseID: test.CourseID, Kind: "test", TestID: &testIDCopy, TestAttemptID: &attemptID}
	updates := Grade{CourseID: test.CourseID, Kind: "test", TestID: &testIDCopy, TestAttemptID: &attemptID, Value: score, MaxValue: maxScore, Comment: "Автоматическая проверка теста"}
	if err := db.Where(&Grade{UserID: u.ID, Kind: "test", TestID: &testIDCopy}).Assign(updates).FirstOrCreate(&grade).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"attempt": attempt, "grade": grade})
}

func listTestAttempts(c *gin.Context) {
	testID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var test Test
	if err := db.First(&test, testID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "test not found"})
		return
	}
	if !canManageCourse(c, test.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot manage this test"})
		return
	}
	var attempts []TestAttempt
	db.Preload("Student").Preload("Test").Preload("Test.Questions", func(q *gorm.DB) *gorm.DB {
		return q.Order("sort_order asc, id asc")
	}).Where("test_id = ?", testID).Order("created_at desc").Find(&attempts)

	rows := make([]gin.H, 0, len(attempts))
	for _, attempt := range attempts {
		var answers map[string]interface{}
		if err := json.Unmarshal([]byte(attempt.Answers), &answers); err != nil {
			answers = map[string]interface{}{}
		}
		manualQuestionGrades := map[string]bool{}
		if strings.TrimSpace(attempt.ManualGrades) != "" {
			_ = json.Unmarshal([]byte(attempt.ManualGrades), &manualQuestionGrades)
		}
		questions := make([]gin.H, 0, len(attempt.Test.Questions))
		for _, q := range attempt.Test.Questions {
			questions = append(questions, questionDTO(q, true))
		}
		rows = append(rows, gin.H{
			"id":        attempt.ID,
			"testId":    attempt.TestID,
			"studentId": attempt.StudentID,
			"student": gin.H{
				"id":        attempt.Student.ID,
				"fullName":  attempt.Student.FullName,
				"email":     attempt.Student.Email,
				"groupName": attempt.Student.GroupName,
			},
			"test": gin.H{
				"id":        attempt.Test.ID,
				"title":     attempt.Test.Title,
				"questions": questions,
			},
			"answers":              answers,
			"attemptNumber":        attempt.AttemptNumber,
			"score":                attempt.Score,
			"autoScore":            attempt.AutoScore,
			"maxScore":             attempt.MaxScore,
			"status":               attempt.Status,
			"comment":              attempt.Comment,
			"manualQuestionGrades": manualQuestionGrades,
			"createdAt":            attempt.CreatedAt,
			"updatedAt":            attempt.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, rows)
}

func gradeTestAttempt(c *gin.Context) {
	attemptID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var attempt TestAttempt
	if err := db.Preload("Test").First(&attempt, attemptID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "test attempt not found"})
		return
	}
	if !canManageCourse(c, attempt.Test.CourseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot check this test attempt"})
		return
	}
	var req struct {
		Score              int             `json:"score"`
		MaxScore           int             `json:"maxScore"`
		Comment            string          `json:"comment"`
		QuestionGrades     map[string]bool `json:"questionGrades"`
		TextQuestionGrades map[string]bool `json:"textQuestionGrades"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	manualGrades := req.QuestionGrades
	if len(manualGrades) == 0 && len(req.TextQuestionGrades) > 0 {
		manualGrades = req.TextQuestionGrades
	}
	if len(manualGrades) > 0 {
		var test Test
		if err := db.Preload("Questions").First(&test, attempt.TestID).Error; err == nil {
			var answers map[string]interface{}
			_ = json.Unmarshal([]byte(attempt.Answers), &answers)
			req.Score, req.MaxScore = calculateTestScore(test.Questions, answers, manualGrades)
		}
	}
	if req.MaxScore <= 0 {
		req.MaxScore = attempt.MaxScore
	}
	if req.Score < 0 || (req.MaxScore > 0 && req.Score > req.MaxScore) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid score"})
		return
	}
	attempt.Score = req.Score
	attempt.MaxScore = req.MaxScore
	attempt.Comment = req.Comment
	if len(manualGrades) > 0 {
		manualRaw, _ := json.Marshal(manualGrades)
		attempt.ManualGrades = string(manualRaw)
	}
	attempt.Status = "checked"
	if err := db.Save(&attempt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	testIDCopy := attempt.TestID
	attemptIDCopy := attempt.ID
	grade := Grade{UserID: attempt.StudentID, Kind: "test", TestID: &testIDCopy}
	updates := Grade{CourseID: attempt.Test.CourseID, Kind: "test", TestID: &testIDCopy, TestAttemptID: &attemptIDCopy, Value: req.Score, MaxValue: req.MaxScore, Comment: req.Comment}
	if err := db.Where(&Grade{UserID: attempt.StudentID, Kind: "test", TestID: &testIDCopy}).Assign(updates).FirstOrCreate(&grade).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"attempt": attempt, "grade": grade})
}

func calculateTestScore(questions []TestQuestion, answers map[string]interface{}, manualQuestionGrades map[string]bool) (int, int) {
	score := 0
	maxScore := 0
	for _, q := range questions {
		points := q.Points
		if points <= 0 {
			points = 1
		}
		maxScore += points
		questionKey := strconv.Itoa(int(q.ID))
		if manualQuestionGrades != nil {
			if checked, exists := manualQuestionGrades[questionKey]; exists {
				if checked {
					score += points
				}
				continue
			}
		}
		answerValues := valuesToStrings(answers[questionKey])
		correctValues := correctAnswersFromQuestion(q)
		switch q.Type {
		case "checkbox":
			score += checkboxPartialScore(answerValues, correctValues, points)
		default:
			if answersEqual(answerValues, correctValues) {
				score += points
			}
		}
	}
	return score, maxScore
}

func checkboxPartialScore(answerValues []string, correctValues []string, points int) int {
	selected := normalizeAnswerList(answerValues)
	correct := normalizeAnswerList(correctValues)
	if len(selected) == 0 || len(correct) == 0 {
		return 0
	}
	correctSet := map[string]bool{}
	for _, value := range correct {
		correctSet[value] = true
	}
	correctPicked := 0
	for _, value := range selected {
		if !correctSet[value] {
			return 0
		}
		correctPicked++
	}
	if correctPicked == len(correct) {
		return points
	}
	partial := points * correctPicked / len(correct)
	if partial == 0 && correctPicked > 0 {
		return 1
	}
	return partial
}
