package main

import (
	"bufio"
	"errors"
	"io"
	"jiaming2012/receipt-processor/database"
	"jiaming2012/receipt-processor/models"
	"math"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func printArray(data []string) string {
	if len(data) == 1 {
		return data[0]
	}

	return data[0] + " >> " + printArray(data[1:])
}

const float64EqualityThreshold = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

func run() {
	var programExitMessage string

	f, fileErr := os.Open("receipts/unprocessed/6.txt")
	if fileErr != nil {
		panic(fileErr)
	}

	db := database.GetDB()
	defer database.ReleaseDB()

	meta := models.Meta{}
	purchases := make(models.Purchases, 0)
	curPurchase := models.NewPurchase()
	purchaseNo := uint(0)

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Info("Finished parsing receipt")
				break
			}

			panic(err)
		}

		line = strings.Trim(line, "\n")
		meta.ProcessLine(line, db)

		err = curPurchase.ProcessLine(line, purchaseNo)
		if err != nil {
			panic(err)
		}
		if curPurchase.IsProcessed {
			item, fetchErr := models.FindOrCreateItem(curPurchase, &meta, db)
			if fetchErr != nil {
				log.Fatal(fetchErr)
			}

			curPurchase.ItemId = item.ID
			purchases = append(purchases, *curPurchase)
			curPurchase = models.NewPurchase()
			purchaseNo += 1
		}
	}

	if !meta.IsProcessed {
		log.Fatalf("meta data not processed: %v", meta.Unprocessed())
	}

	if *meta.TotalUnits+*meta.TotalCases != *meta.TotalItems {
		log.Fatalf("expected TotalUnits %v + Total Cases %v to equal TotalItems %v", *meta.TotalUnits, *meta.TotalCases, *meta.TotalItems)
	}

	unitsCount, casesCounts := purchases.Count()
	if unitsCount != int(*meta.TotalUnits) {
		log.Fatalf("expected unitsCount %v to equal TotalUnits %v. Please verify spelling and that the last line for each item on the receipt starts with either \"UNITS  \" or \"CASES   UNITS \"", unitsCount, *meta.TotalUnits)
	}

	if casesCounts != int(*meta.TotalCases) {
		log.Fatalf("expected casesCounts %v to equal TotalCases %v. Please verify spelling and that the last line for each item on the receipt starts with either \"UNITS  \" or \"CASES   UNITS \"", casesCounts, *meta.TotalCases)
	}

	purchasesTotal := purchases.Total()
	if !almostEqual(purchasesTotal, *meta.Subtotal) {
		log.Fatalf("expected purchasesTotal %f to equal receipt subtotal %f", purchasesTotal, *meta.Subtotal)
	}

	var metaSaved models.Meta
	tx := db.Where(models.Meta{
		StoreId:   meta.StoreId,
		Timestamp: meta.Timestamp,
	}).Find(&metaSaved)

	if tx.Error != nil {
		panic(tx.Error)
	}

	rowsAffected := tx.RowsAffected

	tx = tx.Updates(&meta)

	if tx.Error != nil {
		panic(tx.Error)
	}

	var metaId uint
	if rowsAffected == 0 {
		db.Create(&meta)
		metaId = meta.Model.ID
		programExitMessage = "New receipt processed"
	} else {
		metaId = metaSaved.Model.ID
		programExitMessage = "Existing receipt reprocessed"
	}

	if tx.Error != nil {
		panic(tx.Error)
	}

	for _, purchase := range purchases {
		purchase.MetaId = metaId

		var purchaseSaved models.Purchase
		tx = db.Where(models.Purchase{
			MetaId:   purchase.MetaId,
			Position: purchase.Position,
		}).Find(&purchaseSaved)

		if tx.Error != nil {
			panic(tx.Error)
		}

		rowsAffected = tx.RowsAffected

		tx = tx.Updates(&meta)

		if rowsAffected == 0 {
			db.Create(&purchase)
		}

		if tx.Error != nil {
			panic(tx.Error)
		}
	}

	log.Infof("%s: %v", programExitMessage, meta)
}

func setupDB() {
	log.Info("Setting up database ...")
	if err := database.Setup(); err != nil {
		log.Errorf("failed to setup database: %v", err)
		return
	}
	db := database.GetDB()
	defer database.ReleaseDB()

	db.AutoMigrate(&models.Meta{})
	db.AutoMigrate(&models.Purchase{})
	db.AutoMigrate(&models.PurchaseItem{})
	db.AutoMigrate(&models.Store{})

	models.PopulateItemsCache(db)

	log.Info("Db setup complete!")
}

func main() {
	log.Info("Receipt Processor App v0.01")
	setupDB()
	run()
}
