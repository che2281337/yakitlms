package main

import (
	"os"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var jwtSecret []byte

func openDatabase() (*gorm.DB, error) {
	dsn := getenv("MYSQL_DSN", "root@tcp(127.0.1.31:3306)"+
		"/yakit_lms?charset=utf8mb4&parseTime=True&loc=Local")
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
