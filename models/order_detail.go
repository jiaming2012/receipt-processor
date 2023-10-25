package models

type OrderDetail struct {
	ID             uint     `gorm:"primaryKey;column:id" csv:"ID"`
	Location       string   `gorm:"column:location" csv:"Location"`
	OrderNumber    int      `gorm:"column:order_number;uniqueIndex:compositeItemDetail;not null" csv:"Order #"`
	Opened         DateTime `gorm:"column:opened;uniqueIndex:compositeItemDetail;not null" csv:"Opened"`
	NumGuests      int      `gorm:"column:num_guests" csv:"# of Guests"`
	Server         string   `gorm:"column:server" csv:"Server"`
	Table          string   `gorm:"column:table" csv:"Table"`
	DiscountAmount float64  `gorm:"column:discount_amount" csv:"Discount Amount"`
	Amount         float64  `gorm:"column:amount" csv:"Amount"`
	Tax            float64  `gorm:"column:tax" csv:"Tax"`
	Tip            float64  `gorm:"column:tip" csv:"Tip"`
	Gratuity       float64  `gorm:"column:gratuity" csv:"Gratuity"`
}

// TableName sets the table name explicitly.
func (OrderDetail) TableName() string {
	return "toast_order_details"
}
