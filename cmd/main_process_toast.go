package main

import (
	"fmt"
	"jiaming2012/receipt-processor/database"
	"jiaming2012/receipt-processor/models"
	"os"
	"path/filepath"

	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func UpsertOrderDetail(newOrderDetail *models.OrderDetail, db *gorm.DB) error {
	var orderDetail models.OrderDetail

	if tx := db.Find(&orderDetail, "order_number = ? and opened between ? and ?", newOrderDetail.OrderNumber); tx.Error == nil {
		if tx.RowsAffected == 0 {
			tx = db.Create(newOrderDetail)
			if tx.Error != nil {
				return tx.Error
			}
		} else {
			newOrderDetail.ID = orderDetail.ID
			db.Updates(newOrderDetail)
		}
	} else {
		log.Errorf(tx.Error.Error())
	}

	return nil
}

func setupDB() {
	log.Info("Setting up database ...")
	if err := database.Setup(); err != nil {
		log.Errorf("failed to setup database: %v", err)
		return
	}
	db := database.GetDB()
	defer database.ReleaseDB()

	db.AutoMigrate(&models.OrderDetail{})

	log.Info("Db setup complete!")
}

func moveFileToDirectory(csvFile string, newDirectory string) error {
	// Get the full path of the csvFile
	fullPath, err := filepath.Abs(csvFile)
	if err != nil {
		return err
	}

	// Get the full path of the new directory
	newDirFullPath, err := filepath.Abs(newDirectory)
	if err != nil {
		return err
	}

	// Create the new directory if it doesn't exist
	if _, err := os.Stat(newDirFullPath); os.IsNotExist(err) {
		if err := os.MkdirAll(newDirFullPath, 0755); err != nil {
			return err
		}
	}

	// Move the file to the new directory
	newFilePath := filepath.Join(newDirFullPath, filepath.Base(fullPath))
	if err := os.Rename(fullPath, newFilePath); err != nil {
		return err
	}

	fmt.Printf("Moved %s to %s\n", fullPath, newFilePath)
	return nil
}

func main() {
	// Get the path to the directory containing the go.mod file.
	rootDir, err := filepath.Abs(filepath.Dir("../"))
	if err != nil {
		panic(err)
	}

	// Change the current working directory to the root directory.
	if err := os.Chdir(rootDir); err != nil {
		panic(err)
	}

	if err := database.Setup(); err != nil {
		panic(err)
	}

	setupDB()

	db := database.GetDB()
	defer database.ReleaseDB()

	// Parse the CSV data using gocsv.
	csvFile := "receipts/unprocessed/toast/OrderDetails_2022_01_01-2023_09_23.csv"
	var records []models.OrderDetail
	f, err := os.Open(csvFile)
	if err != nil {
		panic(err)
	}

	if err = gocsv.UnmarshalFile(f, &records); err != nil {
		panic(err)
	}

	for _, record := range records {
		err = UpsertOrderDetail(&record, db)
		if err != nil {
			log.Errorf("failed to upsert order detail: %v", err)
			break
		}
	}

	// write a function that moves csvFile to a new directory
	if err == nil {
		newDirectory := "receipts/processed/toast"
		err = moveFileToDirectory(csvFile, newDirectory)
		if err != nil {
			log.Errorf("failed to move file to directory: %v", err)
		}
	}
}
