package models

import (
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SKU string

type Description string

type Item struct {
	gorm.Model
	SKU         SKU         `gorm:"uniqueIndex:compositeItem;not null"`
	Description Description `gorm:"not null"`
	StoreId     uint        `gorm:"uniqueIndex:compositeItem;not null"`
}

var cache map[SKU]Item

func PopulateItemsCache(db *gorm.DB) {
	var items []Item
	cache = make(map[SKU]Item)

	db.Find(&items)

	for _, item := range items {
		cache[item.SKU] = item
	}
}

func FindOrCreateItemFromPurchaseV2DTO(dto *PurchaseV2DTO, meta *MetaV2, db *gorm.DB) (*Item, error) {
	if cached, ok := cache[SKU(dto.SKU)]; ok {
		if string(cached.Description) != dto.Description {
			// todo: consider making this fatal
			log.Warnf("item description [%v] should match stored description [%s], for sku %s @ position %v", dto.Description, cached.Description, dto.SKU)
		}
		return &cached, nil
	}

	newItem := &Item{
		SKU:         SKU(dto.SKU),
		Description: Description(dto.Description),
		StoreId:     meta.StoreID,
	}

	tx := db.Create(newItem)
	if tx.Error != nil {
		return nil, tx.Error
	}

	cache[newItem.SKU] = *newItem

	return newItem, nil
}

func FindOrCreateItem(purchase *Purchase, meta *Meta, db *gorm.DB) (*Item, error) {
	if item, ok := cache[purchase.SKU]; ok {
		if item.Description != purchase.Description {
			// todo: consider making this fatal
			log.Warnf("item description [%v] should match stored description [%v], for sku %v @ position %v", purchase.Description, item.Description, purchase.SKU, purchase.Position)
		}
		return &item, nil
	}

	if meta.StoreId == nil {
		if len(meta.StoreName) == 0 {
			log.Fatal("Unable to find the store name. Please check that store name is properly printed on a single line")
		}

		log.Fatalf("Unable to find the store id for %s", meta.StoreName)
	}

	item := &Item{
		SKU:         purchase.SKU,
		Description: purchase.Description,
		StoreId:     *meta.StoreId,
	}

	tx := db.Create(item)
	if tx.Error != nil {
		return nil, tx.Error
	}

	cache[item.SKU] = *item

	return item, nil
}
