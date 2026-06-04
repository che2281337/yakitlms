package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func listGrades(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	query := db.Preload("User").Preload("Course").Preload("Assignment").Preload("Test").Order("created_at desc")
	if u.Role == "student" {
		query = query.Where("user_id = ?", u.ID)
	} else if u.Role == "teacher" {
		ids := courseIDsForTeacher(u.ID)
		if len(ids) == 0 {
			c.JSON(http.StatusOK, []gin.H{})
			return
		}
		query = query.Where("course_id IN ?", ids)
	} else if userID := c.Query("userId"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	var grades []Grade
	query.Find(&grades)
	items := make([]gin.H, 0, len(grades))
	for _, g := range grades {
		items = append(items, gradeDTO(g))
	}
	c.JSON(http.StatusOK, items)
}

func listSchedule(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	query := db.Order("starts_at asc")
	if u.Role == "student" && u.GroupName != "" {
		query = query.Where("group_name = ?", u.GroupName)
	} else if u.Role == "teacher" {
		ids := courseIDsForTeacher(u.ID)
		if len(ids) == 0 {
			c.JSON(http.StatusOK, []ScheduleItem{})
			return
		}
		query = query.Where("course_id IN ?", ids)
	}
	var items []ScheduleItem
	query.Find(&items)
	c.JSON(http.StatusOK, items)
}
