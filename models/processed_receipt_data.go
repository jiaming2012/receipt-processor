package models

import (
	"time"
)

type ProcessedReceiptData interface {
	GetStoreName() string
	GetAddress() string
	GetTelephone() string
	GetDateTime() time.Time
	GetTotalAmount() float64
	GetTaxAmount() float64
	GetPaymentType() string
	GetCardLast4() string
	GetPurchaseItems() []ReceiptPurchaseItem
}
