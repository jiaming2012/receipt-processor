package models

import (
	"time"

	"gorm.io/gorm"
)

type ToastItemSelectionDetailDTO struct {
	Location    string      `csv:"Location"`
	OrderNumber int         `csv:"Order #"`
	SentDate    DateTimeEST `csv:"Sent Date"`
	MenuItem    string      `csv:"Menu Item"`
	MenuGroup   string      `csv:"Menu Group"`
	Menu        string      `csv:"Menu"`
	SalesCat    string      `csv:"Sales Category"`
	NetPrice    float64     `csv:"Net Price"`
	Qty         int         `csv:"Qty"`
	Void        bool        `csv:"Void?"`
}

func (dto ToastItemSelectionDetailDTO) ConvertToToastItemSelectionDetail(db *gorm.DB) (*ToastItemSelectionDetail, error) {
	menuItem, err := FindOrCreateMenuItem(&dto, db)
	if err != nil {
		return nil, err
	}

	return &ToastItemSelectionDetail{
		MenuItemID:  menuItem.ID,
		MenuItem:    menuItem,
		Location:    dto.Location,
		OrderNumber: dto.OrderNumber,
		SentDate:    dto.SentDate.Time,
		NetPrice:    dto.NetPrice,
		Qty:         dto.Qty,
		Void:        dto.Void,
	}, nil
}

type ToastItemSelectionDetail struct {
	gorm.Model
	MenuItemID  uint      `gorm:"index;not null;column:menu_item_id"`
	MenuItem    *MenuItem `gorm:"foreignKey:MenuItemID;references:ID;not null"`
	Location    string    `gorm:"not null;column:location" csv:"Location"`
	OrderNumber int       `gorm:"not null;column:order_number" csv:"Order #"`
	SentDate    time.Time `gorm:"not null;column:sent_date" csv:"Sent Date"`
	NetPrice    float64   `gorm:"not null;column:net_price" csv:"Net Price"`
	Qty         int       `gorm:"not null;column:qty" csv:"Qty"`
	Void        bool      `gorm:"not null;column:void" csv:"Void?"`
}

func (o ToastItemSelectionDetail) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"menu_item_id": o.MenuItemID,
		"location":     o.Location,
		"order_number": o.OrderNumber,
		"sent_date":    o.SentDate,
		"net_price":    o.NetPrice,
		"qty":          o.Qty,
		"void":         o.Void,
	}
}
