package models

import "gorm.io/gorm"

type SalesInput struct {
	gorm.Model
	MenuItemID    uint `gorm:"not null"`
	MenuItem      MenuItem
	PurchaseItems []PurchaseItem `gorm:"one2many:SalesInputID"`
}
