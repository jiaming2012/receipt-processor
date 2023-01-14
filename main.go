package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"jiaming2012/receipt-processor/database"
	"jiaming2012/receipt-processor/models"
	"log"
	"os"
	"strings"
)

func printArray(data []string) string {
	if len(data) == 1 {
		return data[0]
	}

	return data[0] + " >> " + printArray(data[1:])
}

func run() {
	f, err := os.Open("receipts/unprocessed/3.txt")
	if err != nil {
		panic(err)
	}

	meta := models.Meta{}
	items := make(models.Items, 0)
	curItem := models.NewItem()

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("done")
				break
			}

			panic(err)
		}

		line = strings.Trim(line, "\n")
		meta.ProcessLine(line)
		err = curItem.ProcessLine(line)
		if err != nil {
			panic(err)
		}
		if curItem.IsProcessed {
			items = append(items, *curItem)
			curItem = models.NewItem()
		}
	}

	if !meta.IsProcessed {
		log.Fatalf("meta data not processed, %v", meta)
	}

	itemsTotal := items.Total()
	if itemsTotal != *meta.Subtotal {
		log.Fatalf("expected itemsTotal %f to equal receipt subtotal %f", itemsTotal, *meta.Subtotal)
	}

	if *meta.TotalUnits+*meta.TotalCases != *meta.TotalItems {
		log.Fatalf("expected TotalUnits %v + Total Cases %v to equal TotalItems %v", *meta.TotalUnits, *meta.TotalCases, *meta.TotalItems)
	}

	unitsCount, casesCounts := items.Count()
	if unitsCount != *meta.TotalUnits {
		log.Fatalf("expected unitsCount %v to equal TotalUnits %v", unitsCount, *meta.TotalUnits)
	}

	if casesCounts != *meta.TotalCases {
		log.Fatalf("expected casesCounts %v to equal TotalCases %v", casesCounts, *meta.TotalCases)
	}

	fmt.Println("itemsTotal: ", itemsTotal, " .. ", *meta.Subtotal)
	fmt.Println("meta:", meta)
}

func setupDB() {
	logrus.Info("Setting up database ...")
	if err := database.Setup(); err != nil {
		logrus.Errorf("failed to setup database: %v", err)
		return
	}
	db := database.GetDB()
	db.AutoMigrate(&models.Meta{})
	db.AutoMigrate(&models.Item{})
	database.ReleaseDB()
	logrus.Info("Db setup complete!")
}

func main() {
	logrus.Info("Receipt Processor App v0.01")

	setupDB()

	run()
}
