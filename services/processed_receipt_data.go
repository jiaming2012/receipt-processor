package services

import (
	"fmt"
	"math"

	"gorm.io/gorm"

	"jiaming2012/receipt-processor/services/models"
	service_models "jiaming2012/receipt-processor/services/models"
)

func purchasedItemsTotal(purchasedItems []service_models.ReceiptPurchaseItem) float64 {
	total := 0.0
	for _, item := range purchasedItems {
		total += item.Price
	}

	return total
}

func fetchOrCreateMeta(receiptData service_models.ProcessedReceiptData) (*models.Meta, error) {
	// Fetch or create the meta
	meta, err := fetchOrCreateMeta(receiptData)
	if err != nil {
		return nil, fmt.Errorf("error fetching or creating meta: %v", err)
	}

	return meta, nil
}

func SaveReceiptData(receiptData service_models.ProcessedReceiptData, db *gorm.DB) error {
	meta, err := fetchOrCreateMetaFromProcessedReceiptData(receiptData, db)
	if err != nil {
		return fmt.Errorf("SaveReceiptData: fetch meta: %v", err)
	}

	fmt.Println("meta", meta)

	return nil
}

func ValidateReceiptData(receiptData service_models.ProcessedReceiptData) error {
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

func fetchOrCreateMetaFromProcessedReceiptData(data service_models.ProcessedReceiptData, db *gorm.DB) (*service_models.Meta, error) {
	storeName := service_models.StoreName(data.GetStoreName())

	store, err := service_models.FindOrCreateStore(storeName, db)
	if err != nil {
		return nil, fmt.Errorf("error finding or creating store: %v", err)
	}

	ts := data.GetDateTime()
	meta, err := service_models.FindOrCreateMeta(store, &ts, db)
	if err != nil {
		return nil, fmt.Errorf("error finding or creating meta: %v", err)
	}

	if err := service_models.UpdateMeta(data, meta, db); err != nil {
		return nil, fmt.Errorf("error updating meta: %v", err)
	}

	return meta, nil
}
