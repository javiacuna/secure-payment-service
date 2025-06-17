package models

import (
	"gorm.io/gorm"
)

type Transfer struct {
	gorm.Model
	TransferID  string `gorm:"uniqueIndex"`
	FromAccount string
	ToAccount   string
	Amount      float64
	Currency    string
	Status      string
}
