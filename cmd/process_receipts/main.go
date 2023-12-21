package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"jiaming2012/receipt-processor/database"
	"jiaming2012/receipt-processor/models"
	service_models "jiaming2012/receipt-processor/services/models"
)

func setupDB() {
	log.Info("Setting up database ...")
	if err := database.Setup(); err != nil {
		log.Errorf("failed to setup database: %v", err)
		return
	}
	db := database.GetDB()
	defer database.ReleaseDB()

	db.AutoMigrate(&models.ToastOrderDetail{})
	db.AutoMigrate(&service_models.Store{})
	db.AutoMigrate(&service_models.Meta{})
	db.AutoMigrate(&models.PurchaseV2{})
	db.AutoMigrate(&models.PurchaseItem{})
	db.AutoMigrate(&models.ToastItemSelectionDetail{})
	db.AutoMigrate(&models.MenuItem{})
	db.AutoMigrate(&models.PurchaseItemGroup{})
	db.AutoMigrate(&models.PurchaseItem{})
	db.AutoMigrate(&models.Tag{})

	log.Info("Db setup complete!")
}

func main() {
	// Get the PROJECT_DIR environment variable
	projectDir := os.Getenv("PROJECT_DIR")
	if projectDir == "" {
		log.Fatalf("PROJECT_DIR environment variable not set")
	}

	// Connect to the database
	if err := database.Setup(); err != nil {
		panic(err)
	}

	setupDB()

	db := database.GetDB()
	defer database.ReleaseDB()

	handler := models.NewReceiptHandler(projectDir)
	handler.Add("giant", &models.GiantReceiptProcessor{})

	handler.Run(db)
	// how to add a new store
	// 1. add a new folder under receipts/unprocessed/newstore
	// 2. create a NewStoreHandler which implements StoreHandler
	// StoreHandler methods: FindOrderDetails, FindOrCreateStore, FindOrCreateMeta, FindOrCreatePurchase
	// 3. go to controller.go and add a new case in the switch statement
	// linking receipts/unprocessed/newstore to NewStoreHandler

}
