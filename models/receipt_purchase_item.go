package models

import "fmt"

type ReceiptPurchaseItem struct {
	Sku         string
	Description string
	UnitPrice   *string
	Price       float64
	Quantity    float64
	IsCase      bool
}

func (i ReceiptPurchaseItem) String() string {
	if i.Sku == "" {
		return fmt.Sprintf("%s=%.2f", i.Description, i.Price)
	}

	return fmt.Sprintf("%s: %s=%.2f", i.Sku, i.Description, i.Price)
}

func (i ReceiptPurchaseItem) GetSku() string {
	return i.Sku
}

func (i ReceiptPurchaseItem) GetDescription() string {
	return i.Description
}
