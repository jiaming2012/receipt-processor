package services

import (
	"fmt"

	"jiaming2012/receipt-processor/models"
)

func fetchStoreName(receiptData models.ProcessedReceiptData) (string, error) {
	// Fetch the store name
	storeName := receiptData.GetStoreName()
	if storeName == "" {
		return "", fmt.Errorf("store name is empty")
	}

	return storeName, nil
}
