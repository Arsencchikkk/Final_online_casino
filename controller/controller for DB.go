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

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	result := config.DB.Select("id, name, description, category, dosage, manufacturer, price, availability, image_url").
		Where("LOWER(name) ILIKE LOWER(?) OR LOWER(manufacturer) ILIKE LOWER(?)",
			"%"+query+"%", "%"+query+"%").
		Find(&medicines)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if len(medicines) == 0 {
		c.JSON(http.StatusOK, []models.Medicine{}) // Возвращаем `[]`, а не null
		return
	}

	c.JSON(http.StatusOK, medicines)
}
func GetMedicinesByCategory(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category parameter is required"})
		return
	}
	var medicines []models.Medicine
	result := config.DB.Where("category = ?", category).Find(&medicines)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	c.JSON(http.StatusOK, medicines)
}
