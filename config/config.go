package config

import (
	"Progetto_APL/models"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func LoadConfig() {
	dsn := "root:password@tcp(127.0.0.1:3307)/db"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Migrazione delle tabelle
	err = DB.AutoMigrate(&models.User{}, &models.Project{}, &models.Code{}, &models.File{}, &models.Task{}, &models.Execution{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Connected to the database!")
}
