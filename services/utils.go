package services

import (
	"fmt"

	service_models "jiaming2012/receipt-processor/services/models"
)

func fetchStoreName(receiptData service_models.ProcessedReceiptData) (string, error) {
	// Fetch the store name
	storeName := receiptData.GetStoreName()
	if storeName == "" {
		return "", fmt.Errorf("store name is empty")
	}

	return storeName, nil
}
