package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type MetaV3 struct {
	gorm.Model
	StoreID     uint       `gorm:"not null"`
	Store       *Store     `gorm:"foreignKey:StoreID;references:ID;not null"`
	StoreId     uint       `gorm:"uniqueIndex:compositeMeta;not null"` // created for createing indexes
	Timestamp   *time.Time `gorm:"uniqueIndex:compositeMeta;not null"`
	TotalUnits  *float64
	TotalCases  *float64
	Subtotal    *float64
	Tax         *float64
	PaymentType *string
	CardLast4   *string
}

func calculateDerivedFields(data ProcessedReceiptData) (float64, float64, string) {
	totalUnits := 0.0
	totalCases := 0.0
	for _, item := range data.GetPurchaseItems() {
		if item.IsCase {
			totalCases += item.Quantity
		} else {
			totalUnits += item.Quantity
		}
	}

	// derive payment type
	cardLast4 := data.GetCardLast4()
	cardLast4 = strings.ReplaceAll(cardLast4, "*", "")
	cardLast4 = strings.ReplaceAll(cardLast4, " ", "")

	return totalUnits, totalCases, cardLast4
}

func UpdateMeta(data ProcessedReceiptData, meta *MetaV3, db *gorm.DB) error {
	tax := data.GetTaxAmount()
	subtotal := data.GetTotalAmount() - tax
	cardLast4 := data.GetCardLast4()
	paymentType := data.GetPaymentType()

	totalUnits, totalCases, cardLast4 := calculateDerivedFields(data)

	// Update the meta
	meta.TotalUnits = &totalUnits
	meta.TotalCases = &totalCases
	meta.Subtotal = &subtotal
	meta.Tax = &tax
	meta.CardLast4 = &cardLast4
	meta.PaymentType = &paymentType

	tx := db.Save(&meta)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func FindOrCreateMeta(store *Store, timestamp *time.Time, db *gorm.DB) (*MetaV3, error) {
	var meta MetaV3

	if tx := db.Find(&meta, "store_id = ? AND timestamp = ?", store.ID, timestamp); tx.Error == nil {
		if tx.RowsAffected == 0 {
			meta.StoreID = store.ID
			meta.Timestamp = timestamp
			tx = db.Save(&meta)
			if tx.Error != nil {
				return nil, tx.Error
			}
		}
	}

	return &meta, nil
}
