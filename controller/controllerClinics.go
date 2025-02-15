package controllers

import (
	"Handbook/config"
	"Handbook/models"
	"net/http"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

func GetClinicsByCity(c *gin.Context) {

	city := c.Query("city")
	var clinics []models.Clinic
	var result *gorm.DB

	if city == "" {
		result = config.DB.Find(&clinics)
	} else {
		result = config.DB.Where("city = ?", city).Find(&clinics)
	}

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, clinics)
}
