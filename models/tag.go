package models

import "gorm.io/gorm"

type Tag struct {
	gorm.Model
	Name               string               `gorm:"index;not null"`
	PurchaseItemGroups []*PurchaseItemGroup `gorm:"many2many:purchase_item_group_tags;"`
}
