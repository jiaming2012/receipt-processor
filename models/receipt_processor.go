package models

import (
	service_models "jiaming2012/receipt-processor/services/models"
)

type ReceiptProcessor interface {
	Process(receiptText string) (service_models.ProcessedReceiptData, error)
}
