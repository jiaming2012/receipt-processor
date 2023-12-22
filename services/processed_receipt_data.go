package services

import (
	"fmt"
	"math"

	"gorm.io/gorm"

	"jiaming2012/receipt-processor/models"
)

func purchasedItemsTotal(purchasedItems []models.ReceiptPurchaseItem) float64 {
	total := 0.0
	for _, item := range purchasedItems {
		total += item.Price
	}

	return total
}

func fetchOrCreateMeta(receiptData models.ProcessedReceiptData) (*models.Meta, error) {
	// Fetch or create the meta
	meta, err := fetchOrCreateMeta(receiptData)
	if err != nil {
		return nil, fmt.Errorf("error fetching or creating meta: %v", err)
	}

	return meta, nil
}

func SaveReceiptData(receiptData models.ProcessedReceiptData, db *gorm.DB) error {
	meta, err := fetchOrCreateMetaFromProcessedReceiptData(receiptData, db)
	if err != nil {
		return fmt.Errorf("SaveReceiptData: fetch meta: %v", err)
	}

	fmt.Println("meta", meta)

	return nil
}

func ValidateReceiptData(receiptData models.ProcessedReceiptData) error {
	// Validate the store name
	if receiptData.GetStoreName() == "" {
		return fmt.Errorf("store name is empty")
	}

	// Validate the address
	if receiptData.GetAddress() == "" {
		return fmt.Errorf("address is empty")
	}

	// Validate the date time
	if receiptData.GetDateTime().IsZero() {
		return fmt.Errorf("date time is empty")
	}

	// Validate the payment type
	if receiptData.GetPaymentType() == "" {
		return fmt.Errorf("payment type is empty")
	}

	// Validate the purchase items
	if len(receiptData.GetPurchaseItems()) == 0 {
		return fmt.Errorf("purchase items is empty")
	}

	// Calculate the total amount
	itemsTotalAmt := purchasedItemsTotal(receiptData.GetPurchaseItems())
	receiptTotalAmt := receiptData.GetTotalAmount() - receiptData.GetTaxAmount()
	if math.Abs(itemsTotalAmt-receiptTotalAmt) > 0.01 {
		return fmt.Errorf("total amount mismatch: itemsTotalAmt %f != %f, receipt subtotal %f, tax %f", itemsTotalAmt, receiptTotalAmt, receiptData.GetTotalAmount(), receiptData.GetTaxAmount())
	}

	return nil
}

func fetchOrCreateMetaFromProcessedReceiptData(data models.ProcessedReceiptData, db *gorm.DB) (*models.MetaV3, error) {
	storeName := models.StoreName(data.GetStoreName())

	store, err := models.FindOrCreateStore(storeName, db)
	if err != nil {
		return nil, fmt.Errorf("error finding or creating store: %v", err)
	}

	ts := data.GetDateTime()
	meta, err := models.FindOrCreateMeta(store, &ts, db)
	if err != nil {
		return nil, fmt.Errorf("error finding or creating meta: %v", err)
	}

	if err := models.UpdateMeta(data, meta, db); err != nil {
		return nil, fmt.Errorf("error updating meta: %v", err)
	}

	if err := UpdateOrCreatePurchase(data, meta, db); err != nil {
		return nil, fmt.Errorf("error updating or creating purchases: %v", err)
	}

	return meta, nil
}

func findOrCreatePurchaseItem(meta *models.MetaV3, item models.ReceiptPurchaseItem, db *gorm.DB) (*models.PurchaseItem, error) {
	var purchaseItem models.PurchaseItem

	if tx := db.Find(&purchaseItem, "sku = ? AND description = ? AND store_id = ?", item.Sku, item.Description, meta.StoreID); tx.Error == nil {
		if tx.RowsAffected == 0 {
			purchaseItem.SKU = models.SKU(item.Sku)
			purchaseItem.Description = models.Description(item.Description)
			purchaseItem.StoreId = meta.StoreID

			tx = db.Save(&purchaseItem)
			if tx.Error != nil {
				return nil, tx.Error
			}
		}
	}

	return &purchaseItem, nil
}

func UpdateOrCreatePurchase(data models.ProcessedReceiptData, meta *models.MetaV3, db *gorm.DB) error {
	// Update the purchase items
	for position, item := range data.GetPurchaseItems() {
		purchaseItem, err := findOrCreatePurchaseItem(meta, item, db)
		if err != nil {
			return err
		}

		var purchase models.PurchaseV2
		if tx := db.Find(&purchase, "meta_id = ? AND position = ?", meta.ID, position); tx.Error == nil {
			if tx.RowsAffected == 0 {
				purchase := models.PurchaseV2{
					MetaId:   meta.ID,
					Position: uint(position),
					IsCase:   item.IsCase,
					Price:    item.Price,
					Quantity: int(item.Quantity),
					ItemId:   purchaseItem.ID,
				}

				tx = db.Save(&purchase)
				if tx.Error != nil {
					return tx.Error
				}
			} else {
				purchase.IsCase = item.IsCase
				purchase.Price = item.Price
				purchase.Quantity = int(item.Quantity)
				purchase.ItemId = purchaseItem.ID

				tx = db.Save(&purchase)
				if tx.Error != nil {
					return tx.Error
				}
			}
		}
	}

	return nil
}
