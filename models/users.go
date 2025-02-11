package models

import "gorm.io/gorm"

type User struct {
	ID uint `json:"id" gorm:"primaryKey"`
	gorm.Model
	FirstName string `json:"first_name" gorm:"not null"`
	LastName  string `json:"last_name" gorm:"not null"`
	Username  string `json:"username" gorm:"not null"`
	Email     string `json:"email" gorm:"unique;not null"`
	Phone     string `json:"phone" gorm:"unique;not null"`
	City      string `json:"city" gorm:"not null"`
	Password  string `json:"password" gorm:"not null"`
}
