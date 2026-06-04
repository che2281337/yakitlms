package main

import (
	"github.com/gin-gonic/gin"
)

func questionDTO(q TestQuestion, includeAnswer bool) gin.H {
	qType := q.Type
	if qType == "" {
		qType = "radio"
	}
	item := gin.H{
		"id":        q.ID,
		"testId":    q.TestID,
		"question":  q.Question,
		"type":      qType,
		"imageUrl":  q.ImageURL,
		"options":   optionsFromJSON(q.OptionsJSON),
		"points":    q.Points,
		"sortOrder": q.SortOrder,
	}
	if includeAnswer {
		answers := correctAnswersFromQuestion(q)
		item["correctAnswer"] = q.CorrectAnswer
		item["correctAnswers"] = answers
	}
	return item
}

func testDTO(t Test, u User) gin.H {
	includeAnswer := u.Role == "teacher" || u.Role == "admin"
	questions := make([]gin.H, 0, len(t.Questions))
	for _, q := range t.Questions {
		questions = append(questions, questionDTO(q, includeAnswer))
	}
	item := gin.H{
		"id":              t.ID,
		"courseId":        t.CourseID,
		"moduleId":        t.ModuleID,
		"title":           t.Title,
		"description":     t.Description,
		"deadline":        t.Deadline,
		"maxScore":        t.MaxScore,
		"attemptLimit":    t.AttemptLimit,
		"durationMinutes": t.DurationMinutes,
		"showResult":      t.ShowResult,
		"questions":       questions,
		"createdAt":       t.CreatedAt,
	}
	if u.Role == "student" {
		attemptLimit := t.AttemptLimit
		if attemptLimit <= 0 {
			attemptLimit = 1
		}
		var attemptsUsed int64
		db.Model(&TestAttempt{}).Where("test_id = ? AND student_id = ?", t.ID, u.ID).Count(&attemptsUsed)
		item["attemptsUsed"] = attemptsUsed
		item["attemptsLeft"] = attemptLimit - int(attemptsUsed)
		item["canAttempt"] = int(attemptsUsed) < attemptLimit
		var attempt TestAttempt
		if err := db.Where("test_id = ? AND student_id = ?", t.ID, u.ID).Order("created_at desc").First(&attempt).Error; err == nil {
			item["attemptId"] = attempt.ID
			item["attemptStatus"] = attempt.Status
			if t.ShowResult && int(attemptsUsed) >= attemptLimit {
				item["score"] = attempt.Score
				item["maxScore"] = attempt.MaxScore
			}
		}
	} else {
		var attempts int64
		db.Model(&TestAttempt{}).Where("test_id = ?", t.ID).Count(&attempts)
		item["attemptsCount"] = attempts
	}
	return item
}

func assignmentDTO(a Assignment, u User) gin.H {
	courseTitle := a.Course.Title
	moduleTitle := a.Module.Title
	item := gin.H{
		"id":             a.ID,
		"courseId":       a.CourseID,
		"moduleId":       a.ModuleID,
		"courseTitle":    courseTitle,
		"moduleTitle":    moduleTitle,
		"title":          a.Title,
		"description":    a.Description,
		"instructionUrl": a.InstructionURL,
		"kind":           a.Kind,
		"deadline":       a.Deadline,
		"maxScore":       a.MaxScore,
		"createdAt":      a.CreatedAt,
	}
	if u.Role == "student" {
		var sub Submission
		if err := db.Preload("Files").Where("assignment_id = ? AND student_id = ?", a.ID, u.ID).First(&sub).Error; err == nil {
			item["submissionId"] = sub.ID
			item["status"] = sub.Status
			item["grade"] = sub.Grade
			item["comment"] = sub.Comment
			item["answer"] = sub.Answer
			item["files"] = sub.Files
			if len(sub.Files) > 0 {
				item["fileUrl"] = sub.Files[0].FileURL
				item["fileName"] = sub.Files[0].FileName
			} else {
				item["fileUrl"] = sub.FileURL
				item["fileName"] = sub.FileName
			}
		} else {
			item["status"] = "not_submitted"
		}
	} else {
		var pending int64
		db.Model(&Submission{}).Where("assignment_id = ? AND status = ?", a.ID, "pending").Count(&pending)
		item["pendingSubmissions"] = pending
	}
	return item
}

func gradeDTO(g Grade) gin.H {
	title := ""
	if g.Kind == "assignment" {
		title = g.Assignment.Title
	} else if g.Kind == "test" {
		title = g.Test.Title
	}
	return gin.H{
		"id":          g.ID,
		"userId":      g.UserID,
		"studentName": g.User.FullName,
		"courseId":    g.CourseID,
		"courseTitle": g.Course.Title,
		"kind":        g.Kind,
		"workTitle":   title,
		"value":       g.Value,
		"maxValue":    g.MaxValue,
		"comment":     g.Comment,
		"createdAt":   g.CreatedAt,
	}
}

func courseListDTO(course Course, u User) gin.H {
	var modulesCount, materialsCount, assignmentsCount, testsCount int64
	db.Model(&CourseModule{}).Where("course_id = ?", course.ID).Count(&modulesCount)
	db.Model(&Material{}).Where("course_id = ?", course.ID).Count(&materialsCount)
	db.Model(&Assignment{}).Where("course_id = ?", course.ID).Count(&assignmentsCount)
	db.Model(&Test{}).Where("course_id = ?", course.ID).Count(&testsCount)
	progress := 0
	if u.Role == "student" {
		var enrollment Enrollment
		if err := db.Where("user_id = ? AND course_id = ?", u.ID, course.ID).First(&enrollment).Error; err == nil {
			progress = enrollment.Progress
		}
	}
	return gin.H{
		"id":               course.ID,
		"title":            course.Title,
		"category":         course.Category,
		"description":      course.Description,
		"imageUrl":         course.ImageURL,
		"teacherId":        course.TeacherID,
		"teacherName":      course.Teacher.FullName,
		"status":           course.Status,
		"modulesCount":     modulesCount,
		"materialsCount":   materialsCount,
		"assignmentsCount": assignmentsCount,
		"testsCount":       testsCount,
		"progress":         progress,
		"createdAt":        course.CreatedAt,
		"updatedAt":        course.UpdatedAt,
	}
}

func courseDetailDTO(course Course, u User) gin.H {
	modules := make([]gin.H, 0, len(course.Modules))
	for _, m := range course.Modules {
		assignments := make([]gin.H, 0, len(m.Assignments))
		for _, a := range m.Assignments {
			a.Course = course
			a.Module = m
			assignments = append(assignments, assignmentDTO(a, u))
		}
		tests := make([]gin.H, 0, len(m.Tests))
		for _, t := range m.Tests {
			tests = append(tests, testDTO(t, u))
		}
		materials := make([]gin.H, 0, len(m.Materials))
		for _, mat := range m.Materials {
			materials = append(materials, gin.H{
				"id": mat.ID, "courseId": mat.CourseID, "moduleId": mat.ModuleID,
				"title": mat.Title, "kind": mat.Kind, "url": mat.URL, "content": mat.Content, "sortOrder": mat.SortOrder,
			})
		}
		items := make([]gin.H, 0, len(materials)+len(assignments)+len(tests))
		for _, mat := range materials {
			mat["itemType"] = "material"
			items = append(items, mat)
		}
		for _, a := range assignments {
			a["itemType"] = "assignment"
			items = append(items, a)
		}
		for _, t := range tests {
			t["itemType"] = "test"
			items = append(items, t)
		}
		modules = append(modules, gin.H{
			"id":          m.ID,
			"courseId":    m.CourseID,
			"title":       m.Title,
			"description": m.Description,
			"sortOrder":   m.SortOrder,
			"items":       items,
			"materials":   materials,
			"assignments": assignments,
			"tests":       tests,
		})
	}
	return gin.H{
		"id":          course.ID,
		"title":       course.Title,
		"category":    course.Category,
		"description": course.Description,
		"imageUrl":    course.ImageURL,
		"teacherId":   course.TeacherID,
		"teacherName": course.Teacher.FullName,
		"status":      course.Status,
		"modules":     modules,
		"createdAt":   course.CreatedAt,
		"updatedAt":   course.UpdatedAt,
	}
}
