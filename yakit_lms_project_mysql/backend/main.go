package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	gin.SetMode(getenv("GIN_MODE", "debug"))
	jwtSecret = []byte(getenv("JWT_SECRET", "dev-secret-change-me"))

	var err error
	db, err = openDatabase()
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatal(err)
	}

	r := setupRouter()
	port := getenv("APP_PORT", "8080")
	log.Printf("YAKIT LMS backend started: http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
