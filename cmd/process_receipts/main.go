package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"jiaming2012/receipt-processor/database"
	"jiaming2012/receipt-processor/models"
	"jiaming2012/receipt-processor/services"
	"jiaming2012/receipt-processor/utils"
)

func Run(handler *models.ReceiptHandler, db *gorm.DB) {
	for _, binding := range handler.Bindings {
		log.Debugf("Running processor for path: %s", binding.Path)
		unprocessedFiles := filepath.Join(handler.ProjectDir, "receipts", "unprocessed", binding.Path)
		moveToDir := filepath.Join(handler.ProjectDir, "receipts", "processed", binding.Path)

		err := processDirectory(unprocessedFiles, moveToDir, db, binding.Processor.Process)
		if err != nil {
			log.Errorf("error walking directory: %v", err)
			continue
		}
	}
}

// processDirectory walks the directory and calls the callback function for each file.
// The callback function is passed the contents of the file.
// The file is moved to the processedDir after the callback function is called.
func processDirectory(dir string, processedDir string, db *gorm.DB, callback func(string) (models.ProcessedReceiptData, error)) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error walking directory: %v", err)
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error opening file: %v", err)
		}

		defer file.Close()

		bytes, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("error reading file: %v", err)
		}

		if !info.IsDir() {
			receiptData, callbackErr := callback(string(bytes))
			if callbackErr != nil {
				return fmt.Errorf("error in callback: %v", callbackErr)
			}

			// Validate the receipt data
			if err := services.ValidateReceiptData(receiptData); err != nil {
				return fmt.Errorf("error validating receipt data: %v", err)
			}

			// Transform the receipt data

			// Save data to database
			if err := services.SaveReceiptData(receiptData, db); err != nil {
				return fmt.Errorf("error saving receipt data: %v", err)
			}

			err := utils.MoveFileToDirectory(path, processedDir)
			if err != nil {
				return fmt.Errorf("error moving file: %v", err)
			}
		}

		return nil
	})
}

func setupDB() {
	log.Info("Setting up database ...")
	if err := database.Setup(); err != nil {
		log.Errorf("failed to setup database: %v", err)
		return
	}
	db := database.GetDB()
	defer database.ReleaseDB()

	db.AutoMigrate(&models.ToastOrderDetail{})
	db.AutoMigrate(&models.Store{})
	db.AutoMigrate(&models.MetaV3{})
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

	Run(handler, db)
	// how to add a new store
	// 1. add a new folder under receipts/unprocessed/newstore
	// 2. create a NewStoreHandler which implements StoreHandler
	// StoreHandler methods: FindOrderDetails, FindOrCreateStore, FindOrCreateMeta, FindOrCreatePurchase
	// 3. go to controller.go and add a new case in the switch statement
	// linking receipts/unprocessed/newstore to NewStoreHandler

}
