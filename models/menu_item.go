package models

import "gorm.io/gorm"

type MenuItem struct {
	gorm.Model
	Name     string `gorm:"uniqueIndex:compositeItem;not null" csv:"Menu Item"`
	Group    string `gorm:"not null" csv:"Menu Group"`
	Menu     string `gorm:"uniqueIndex:compositeItem;not null" csv:"Menu"`
	SalesCat string `gorm:"not null" csv:"Sales Category"`
}

// todo: make more robust
var menuItemCache map[string]MenuItem

func PopulateMenuItemCache(db *gorm.DB) {
	var menuItems []MenuItem
	menuItemCache = make(map[string]MenuItem)

	db.Find(&menuItems)

	for _, menuItem := range menuItems {
		menuItemCache[menuItem.Name] = menuItem
	}
}

// FindOrCreateMenuItem returns a MenuItem from the cache if it exists, otherwise it creates a new MenuItem and returns it.
func FindOrCreateMenuItem(dto *ToastItemSelectionDetailDTO, db *gorm.DB) (*MenuItem, error) {
	if cached, ok := menuItemCache[dto.MenuItem]; ok {
		return &cached, nil
	}

	newMenuItem := &MenuItem{
		Name:     dto.MenuItem,
		Group:    dto.MenuGroup,
		Menu:     dto.Menu,
		SalesCat: dto.SalesCat,
	}

	tx := db.Create(newMenuItem)
	if tx.Error != nil {
		return nil, tx.Error
	}

	menuItemCache[newMenuItem.Name] = *newMenuItem

	return newMenuItem, nil
}
