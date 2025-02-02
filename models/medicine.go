package models

type Medicine struct {
	ID           uint    `json:"id" gorm:"primaryKey"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Category     string  `json:"category"`
	Dosage       string  `json:"dosage"`
	Manufacturer string  `json:"manufacturer"`
	Price        float64 `json:"price"`
	Availability bool    `json:"availability"`
}
