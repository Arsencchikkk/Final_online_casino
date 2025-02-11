package models

type Favorite struct {
	ID         uint     `json:"id" gorm:"primaryKey"`
	UserID     uint     `json:"user_id"`
	MedicineID uint     `json:"medicine_id"`
	Medicine   Medicine `json:"medicine" gorm:"foreignKey:MedicineID"`
}
