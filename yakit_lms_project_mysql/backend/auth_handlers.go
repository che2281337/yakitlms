package main

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

func publicUser(u User) gin.H {
	return gin.H{
		"id":        u.ID,
		"fullName":  u.FullName,
		"email":     u.Email,
		"role":      u.Role,
		"groupName": u.GroupName,
		"specialty": u.Specialty,
		"createdAt": u.CreatedAt,
		"updatedAt": u.UpdatedAt,
	}
}

func validateRole(role string) bool {
	return role == "student" || role == "teacher" || role == "admin"
}

func makeToken(u User) (string, error) {
	claims := jwt.MapClaims{"userId": u.ID, "role": u.Role, "exp": time.Now().Add(24 * time.Hour).Unix()}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func register(c *gin.Context) {
	var req struct {
		FullName  string `json:"fullName"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		GroupName string `json:"groupName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.FullName) == "" || strings.TrimSpace(req.Email) == "" || len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fullName, valid email and password length >= 6 are required"})
		return
	}
	hash, err := hashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	u := User{FullName: req.FullName, Email: strings.ToLower(req.Email), PasswordHash: hash, Role: "student", GroupName: req.GroupName, Specialty: "09.02.07 Информационные системы и программирование"}
	if err := db.Create(&u).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}
	token, err := makeToken(u)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": token, "user": publicUser(u)})
}

func login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Email) == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}
	var u User
	if err := db.Where("email = ?", strings.ToLower(req.Email)).First(&u).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	token, err := makeToken(u)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": publicUser(u)})
}

func me(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, publicUser(*u))
}
