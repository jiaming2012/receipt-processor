package main

import (
	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"jiaming2012/receipt-processor/database"
	"jiaming2012/receipt-processor/models"
	"os"
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

const csvFile = "receipts/unprocessed/toast/OrderDetails_2022_01_01-2023_09_23.csv"

func main() {
	if err := database.Setup(); err != nil {
		panic(err)
	}

	setupDB()

	db := database.GetDB()
	defer database.ReleaseDB()

	// Parse the CSV data using gocsv.
	var records []models.OrderDetail
	f, err := os.Open(csvFile)
	if err != nil {
		panic(err)
	}

	if err = gocsv.UnmarshalFile(f, &records); err != nil {
		panic(err)
	}

	for _, record := range records {
		UpsertOrderDetail(&record, db)
	}
}
