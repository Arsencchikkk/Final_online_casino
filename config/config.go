package config

import (
	"Handbook/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	dsn := os.Getenv("DB_DSN")
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Не удалось подключиться к БД")
	}

	DB = database
	log.Println("✅ База данных подключена успешно!")

	AutoMigrate()
}

func AutoMigrate() {
	err := DB.AutoMigrate(&models.User{}, &models.Favorite{}) // ❗ Добавили Favorite
	if err != nil {
		log.Fatal("❌ Ошибка миграции:", err)
	}
	log.Println("✅ Миграция выполнена успешно! Таблицы обновлены.")
}
