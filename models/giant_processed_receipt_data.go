package models

import "time"

// GiantProcessedReceiptData implements ProcessedReceiptData
type GiantProcessedReceiptData struct {
	StoreName   string
	Address     string
	Telephone   string
	DateTime    time.Time
	TotalAmount float64
	TaxAmount   float64
	PaymentType string
	CardLast4   string
	Items       []ReceiptPurchaseItem
}

func (g GiantProcessedReceiptData) GetStoreName() string {
	return g.StoreName
}

func (g GiantProcessedReceiptData) GetAddress() string {
	return g.Address
}

func (g GiantProcessedReceiptData) GetTelephone() string {
	return g.Telephone
}

func (g GiantProcessedReceiptData) GetDateTime() time.Time {
	return g.DateTime
}

func (g GiantProcessedReceiptData) GetTotalAmount() float64 {
	return g.TotalAmount
}

func (g GiantProcessedReceiptData) GetTaxAmount() float64 {
	return g.TaxAmount
}

func (g GiantProcessedReceiptData) GetPaymentType() string {
	return g.PaymentType
}

func (g GiantProcessedReceiptData) GetCardLast4() string {
	return g.CardLast4
}

func (g GiantProcessedReceiptData) GetPurchaseItems() []ReceiptPurchaseItem {
	return g.Items
}
