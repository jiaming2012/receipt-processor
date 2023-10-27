package models

type ToastOrderDetail struct {
	ID             uint        `gorm:"primaryKey;column:id" csv:"ID"`
	Location       string      `gorm:"column:location" csv:"Location"`
	OrderNumber    int         `gorm:"column:order_number;uniqueIndex:compositeItemDetail;not null" csv:"Order #"`
	Opened         DateTimeEST `gorm:"column:opened;uniqueIndex:compositeItemDetail;not null" csv:"Opened"`
	NumGuests      int         `gorm:"column:num_guests" csv:"# of Guests"`
	Server         string      `gorm:"column:server" csv:"Server"`
	Table          string      `gorm:"column:table" csv:"Table"`
	DiscountAmount float64     `gorm:"column:discount_amount" csv:"Discount Amount"`
	Amount         float64     `gorm:"column:amount" csv:"Amount"`
	Tax            float64     `gorm:"column:tax" csv:"Tax"`
	Tip            float64     `gorm:"column:tip" csv:"Tip"`
	Gratuity       float64     `gorm:"column:gratuity" csv:"Gratuity"`
}

func (o ToastOrderDetail) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"location":        o.Location,
		"order_number":    o.OrderNumber,
		"opened":          o.Opened,
		"num_guests":      o.NumGuests,
		"server":          o.Server,
		"table":           o.Table,
		"discount_amount": o.DiscountAmount,
		"amount":          o.Amount,
		"tax":             o.Tax,
		"tip":             o.Tip,
		"gratuity":        o.Gratuity,
	}
}

// TableName sets the table name explicitly.
func (ToastOrderDetail) TableName() string {
	return "toast_order_details"
}
