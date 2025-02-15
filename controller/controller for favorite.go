package controllers

import (
	"Handbook/config"
	"Handbook/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Добавить лекарство в избранное
func GetFavorites(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var favorites []models.Favorite
	config.DB.Preload("Medicine").Where("user_id = ?", userID).Find(&favorites)

	c.JSON(http.StatusOK, favorites)
}

func AddToFavorites(c *gin.Context) {
	var favorite models.Favorite
	if err := c.ShouldBindJSON(&favorite); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var existing models.Favorite
	result := config.DB.Where("user_id = ? AND medicine_id = ?", favorite.UserID, favorite.MedicineID).First(&existing)
	if result.RowsAffected > 0 {
		c.JSON(http.StatusConflict, gin.H{"message": "Medicine is already in favorites"})
		return
	}

	config.DB.Create(&favorite)
	c.JSON(http.StatusOK, gin.H{"message": "Medicine added to favorites!"})
}

func RemoveFromFavorites(c *gin.Context) {
	id := c.Param("id")
	if err := config.DB.Delete(&models.Favorite{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from favorites"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Medicine removed from favorites!"})
}
