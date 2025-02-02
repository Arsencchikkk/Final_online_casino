package controllers

import (
	"Handbook/config"
	"Handbook/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetMedicines(c *gin.Context) {
	var medicines []models.Medicine
	config.DB.Find(&medicines)
	c.JSON(http.StatusOK, medicines)
}

func SearchMedicine(c *gin.Context) {
	query := c.Query("q")
	var medicines []models.Medicine

	// Проверяем, введён ли запрос
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	// Выполняем поиск в БД
	result := config.DB.Where("name ILIKE ?", "%"+query+"%").Find(&medicines)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + result.Error.Error()})
		return
	}

	// Если ничего не найдено
	if len(medicines) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No Medicines Found"})
		return
	}

	// Возвращаем найденные лекарства
	c.JSON(http.StatusOK, medicines)
}
