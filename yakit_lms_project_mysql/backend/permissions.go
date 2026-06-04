package main

import (
	"github.com/gin-gonic/gin"
)

func canAccessCourse(u User, courseID uint) bool {
	if u.Role == "admin" {
		return true
	}
	if u.Role == "teacher" {
		var count int64
		db.Model(&Course{}).Where("id = ? AND teacher_id = ?", courseID, u.ID).Count(&count)
		return count > 0
	}
	return studentHasCourse(u.ID, courseID)
}

func canManageCourse(c *gin.Context, courseID uint) bool {
	u, ok := currentUser(c)
	if !ok {
		return false
	}
	if u.Role == "admin" {
		return true
	}
	if u.Role != "teacher" {
		return false
	}
	var count int64
	db.Model(&Course{}).Where("id = ? AND teacher_id = ?", courseID, u.ID).Count(&count)
	return count > 0
}

func moduleCourseID(moduleID uint) (uint, bool) {
	var module CourseModule
	if err := db.First(&module, moduleID).Error; err != nil {
		return 0, false
	}
	return module.CourseID, true
}

func studentHasCourse(userID, courseID uint) bool {
	var count int64
	db.Model(&Enrollment{}).Where("user_id = ? AND course_id = ?", userID, courseID).Count(&count)
	return count > 0
}
