package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func dashboard(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	if u.Role == "admin" {
		var users, courses, news, applications int64
		db.Model(&User{}).Count(&users)
		db.Model(&Course{}).Count(&courses)
		db.Model(&News{}).Count(&news)
		db.Model(&Application{}).Where("status = ?", "new").Count(&applications)
		c.JSON(http.StatusOK, gin.H{"user": publicUser(*u), "adminStats": gin.H{"users": users, "courses": courses, "news": news, "newApplications": applications}})
		return
	}

	var schedule []ScheduleItem
	today := time.Now()
	start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)
	end := start.Add(24 * time.Hour)
	scheduleQuery := db.Where("starts_at >= ? AND starts_at < ?", start, end).Order("starts_at asc")
	if u.GroupName != "" && u.Role == "student" {
		scheduleQuery = scheduleQuery.Where("group_name = ?", u.GroupName)
	}
	if u.Role == "teacher" {
		var courseTitles []string
		db.Model(&Course{}).Where("teacher_id = ?", u.ID).Pluck("title", &courseTitles)
		if len(courseTitles) > 0 {
			scheduleQuery = scheduleQuery.Where("course_title IN ?", courseTitles)
		}
	}
	scheduleQuery.Find(&schedule)

	c.JSON(http.StatusOK, gin.H{
		"user":          publicUser(*u),
		"averageGrade":  averageGradeForUser(u.ID),
		"todaySchedule": schedule,
		"currentTasks":  upcomingTasksForUser(*u, 6),
		"recentGrades":  recentGradesForUser(*u, 5),
	})
}

func upcomingTasksForUser(u User, limit int) []gin.H {
	items := []gin.H{}
	var assignments []Assignment
	query := db.Preload("Course").Preload("Module").Where("deadline >= ?", time.Now()).Order("deadline asc").Limit(limit)
	if u.Role == "student" {
		ids := courseIDsForStudent(u.ID)
		if len(ids) == 0 {
			return items
		}
		query = query.Where("course_id IN ?", ids)
	} else if u.Role == "teacher" {
		ids := courseIDsForTeacher(u.ID)
		if len(ids) == 0 {
			return items
		}
		query = query.Where("course_id IN ?", ids)
	}
	query.Find(&assignments)
	for _, a := range assignments {
		item := assignmentDTO(a, u)
		item["kind"] = "assignment"
		items = append(items, item)
	}

	var tests []Test
	tq := db.Preload("Course").Preload("Module").Preload("Questions", func(q *gorm.DB) *gorm.DB { return q.Order("sort_order asc, id asc") }).Where("deadline >= ?", time.Now()).Order("deadline asc").Limit(limit)
	if u.Role == "student" {
		ids := courseIDsForStudent(u.ID)
		if len(ids) == 0 {
			return items
		}
		tq = tq.Where("course_id IN ?", ids)
	} else if u.Role == "teacher" {
		ids := courseIDsForTeacher(u.ID)
		if len(ids) == 0 {
			return items
		}
		tq = tq.Where("course_id IN ?", ids)
	}
	tq.Find(&tests)
	for _, t := range tests {
		item := testDTO(t, u)
		item["kind"] = "test"
		item["courseTitle"] = t.Course.Title
		item["moduleTitle"] = t.Module.Title
		items = append(items, item)
	}
	if len(items) > limit {
		return items[:limit]
	}
	return items
}

func courseIDsForStudent(userID uint) []uint {
	var ids []uint
	db.Model(&Enrollment{}).Where("user_id = ?", userID).Pluck("course_id", &ids)
	return ids
}

func courseIDsForTeacher(userID uint) []uint {
	var ids []uint
	db.Model(&Course{}).Where("teacher_id = ?", userID).Pluck("id", &ids)
	return ids
}

func recentGradesForUser(u User, limit int) []gin.H {
	var grades []Grade
	query := db.Preload("User").Preload("Course").Preload("Assignment").Preload("Test").Order("created_at desc").Limit(limit)
	if u.Role == "student" {
		query = query.Where("user_id = ?", u.ID)
	} else if u.Role == "teacher" {
		ids := courseIDsForTeacher(u.ID)
		if len(ids) == 0 {
			return []gin.H{}
		}
		query = query.Where("course_id IN ?", ids)
	}
	query.Find(&grades)
	items := make([]gin.H, 0, len(grades))
	for _, g := range grades {
		items = append(items, gradeDTO(g))
	}
	return items
}

func averageGradeForUser(userID uint) float64 {
	var grades []Grade
	db.Where("user_id = ?", userID).Find(&grades)
	if len(grades) == 0 {
		return 0
	}
	sum := 0.0
	for _, g := range grades {
		maxValue := g.MaxValue
		if maxValue <= 0 {
			maxValue = 5
		}
		sum += (float64(g.Value) / float64(maxValue)) * 5.0
	}
	return sum / float64(len(grades))
}
