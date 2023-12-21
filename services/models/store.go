package models

import (
	"gorm.io/gorm"
)

type StoreName string

type Store struct {
	gorm.Model
	Name string `gorm:"not null"`
}

func FindOrCreateStore(name StoreName, db *gorm.DB) (*Store, error) {
	var store Store

	if tx := db.Find(&store, "name = ?", name); tx.Error == nil {
		if tx.RowsAffected == 0 {
			store.Name = string(name)
			tx = db.Save(&store)
			if tx.Error != nil {
				return nil, tx.Error
			}
		}
	}

	return &store, nil
}
