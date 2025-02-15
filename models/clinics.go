package models

import "gorm.io/gorm"

type Clinic struct {
	gorm.Model
	Name        string `json:"name" gorm:"not null"`
	City        string `json:"city" gorm:"not null"`
	Address     string `json:"address" gorm:"not null"`
	Description string `json:"description"`
	URL         string `json:"url"`
	ImageURL    string `json:"image_url"`
}
