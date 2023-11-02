package models

import "gorm.io/gorm"

type PurchaseItemGroup struct {
	gorm.Model
	Name          string          `gorm:"uniqueIndex:compositeItem;not null"`
	PurchaseItems []*PurchaseItem `gorm:"many2many:purchase_item_group_purchase_items;"`
	Tags          []*Tag          `gorm:"many2many:purchase_item_group_tags;"`
}
