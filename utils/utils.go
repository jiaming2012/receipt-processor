package utils

import (
	"jiaming2012/receipt-processor/database"
	"jiaming2012/receipt-processor/models"

	log "github.com/sirupsen/logrus"
)

func SetupDB() {
	log.Info("Setting up database ...")
	if err := database.Setup(); err != nil {
		log.Errorf("failed to setup database: %v", err)
		return
	}
	db := database.GetDB()
	defer database.ReleaseDB()

	db.AutoMigrate(&models.ToastOrderDetail{})
	db.AutoMigrate(&models.Store{})
	db.AutoMigrate(&models.MetaV2{})
	db.AutoMigrate(&models.PurchaseV2{})
	db.AutoMigrate(&models.PurchaseItem{})
	db.AutoMigrate(&models.ToastItemSelectionDetail{})
	db.AutoMigrate(&models.MenuItem{})
	db.AutoMigrate(&models.PurchaseItemGroup{})
	db.AutoMigrate(&models.PurchaseItem{})

	log.Info("Db setup complete!")
}
