package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"ocean-strategy/internal/models"
)

var DB *gorm.DB

func InitDB() error {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if user == "" {
		user = "ocean"
	}
	if password == "" {
		password = "ocean_strategy_2024"
	}
	if dbname == "" {
		dbname = "ocean_strategy"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	log.Println("Database connection established")

	err = DB.AutoMigrate(
		&models.Game{},
		&models.Player{},
	)
	if err != nil {
		return err
	}

	log.Println("Database migration completed")
	return nil
}
