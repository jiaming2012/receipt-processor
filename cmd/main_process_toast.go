package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"jiaming2012/receipt-processor/database"
	"jiaming2012/receipt-processor/models"
)

func setTimeToStartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func setTimeToEndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

func DeletePurchasesAfterPosition(meta *models.MetaV2, position int, db *gorm.DB) error {
	tx := db.Where("meta_id = ? AND position >= ?", meta.ID, position).Delete(&models.PurchaseV2{})
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func UpsertPurchase(position uint, purchaseDto *models.PurchaseV2DTO, meta *models.MetaV2, db *gorm.DB) error {
	item, err := models.FindOrCreateItemFromPurchaseV2DTO(purchaseDto, meta, db)
	if err != nil {
		return fmt.Errorf("failed to find or create item: %v", err)
	}

	newPurchase := purchaseDto.ConvertToPurchaseV2()
	newPurchase.ItemId = item.ID
	newPurchase.Position = position
	newPurchase.MetaId = meta.ID

	var purchase models.PurchaseV2
	if tx := db.Find(&purchase, "meta_id = ? AND position = ?", meta.ID, position); tx.Error == nil {
		if tx.RowsAffected == 0 {
			tx = db.Create(newPurchase)
			if tx.Error != nil {
				return tx.Error
			}
		} else {
			m := newPurchase.AsMap()
			db.Model(&purchase).Updates(m)
		}
	}

	return nil
}

func UpsertItemSelectionDetail(dto *models.ToastItemSelectionDetailDTO, db *gorm.DB) error {
	var itemSelectionDetail models.ToastItemSelectionDetail

	// find the item selection detail by order number and opened date
	startDate := setTimeToStartOfDay(dto.SentDate.Time)
	endDate := setTimeToEndOfDay(dto.SentDate.Time)

	if tx := db.Find(&itemSelectionDetail, "order_number = ? and sent_date between ? and ?", dto.OrderNumber, startDate, endDate); tx.Error == nil {
		if tx.RowsAffected == 0 {
			newItemSelectionDetail, err := dto.ConvertToToastItemSelectionDetail(db)
			if err != nil {
				return fmt.Errorf("failed to convert to toast item selection detail: %v", err)
			}

			tx = db.Create(&newItemSelectionDetail)
			if tx.Error != nil {
				return tx.Error
			}
		} else {
			m := itemSelectionDetail.AsMap()
			db.Model(&itemSelectionDetail).Updates(m)
		}
	} else {
		log.Errorf(tx.Error.Error())
	}

	return nil
}

func UpsertOrderDetail(newOrderDetail *models.ToastOrderDetail, db *gorm.DB) error {
	var orderDetail models.ToastOrderDetail

	// find the order detail by order number and opened date
	startDate := setTimeToStartOfDay(newOrderDetail.Opened.Time)
	endDate := setTimeToEndOfDay(newOrderDetail.Opened.Time)

	if tx := db.Find(&orderDetail, "order_number = ? and opened between ? and ?", newOrderDetail.OrderNumber, startDate, endDate); tx.Error == nil {
		if tx.RowsAffected == 0 {
			tx = db.Create(newOrderDetail)
			if tx.Error != nil {
				return tx.Error
			}
		} else {
			newOrderDetail.ID = orderDetail.ID
			m := newOrderDetail.AsMap()
			db.Model(&orderDetail).Updates(m)
		}
	} else {
		log.Errorf(tx.Error.Error())
	}

	return nil
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

func iterateFiles(dir string, upsertToDB func(string) error) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return upsertToDB(path)
		}

		return nil
	})
}

func processMetadata(metaData string) (models.StoreName, time.Time, float64, float64, error) {
	// Verify that "Restaurant Depot" is found in the text.
	if match, _ := regexp.MatchString("Restaurant Depot", metaData); !match {
		return "", time.Time{}, 0.0, 0.0, fmt.Errorf("Restaurant Depot not found in text")
	}

	storeName := models.StoreName("Restaurant Depot")

	// Capture the date in format "12/22/2022 2:50pm" from the text.
	re := regexp.MustCompile(`(\d{1,2}/\d{1,2}/\d{4} \d{1,2}:\d{2} [ap]m)`)
	date := re.FindStringSubmatch(metaData)[1]

	layout := "01/02/2006 3:04 pm"
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return "", time.Time{}, 0.0, 0.0, fmt.Errorf("failed to load location: %v", err)
	}

	parsedDate, err := time.ParseInLocation(layout, date, loc)
	if err != nil {
		return "", time.Time{}, 0.0, 0.0, fmt.Errorf("failed to parse date: %v", err)
	}

	// Search for the Sub-Total, Tax, and Total rows.
	// Parse the CSV data.
	reader := csv.NewReader(strings.NewReader(metaData))
	var records [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			if !errors.Is(err, csv.ErrFieldCount) {
				return "", time.Time{}, 0.0, 0.0, fmt.Errorf("failed to read csv data: %v", err)
			}
		}
		records = append(records, record)
	}

	var subTotal, tax float64
	var foundSubTotal, foundTax bool
	for _, record := range records {
		if len(record) >= 5 && record[1] == "Sub-Total" {
			foundSubTotal = true
			str := strings.Replace(record[4], "$", "", -1)
			str = strings.Replace(str, ",", "", -1)
			subTotal, err = strconv.ParseFloat(str, 64)
			if err != nil {
				return "", time.Time{}, 0.0, 0.0, fmt.Errorf("failed to parse sub-total: %v", err)
			}
		} else if len(record) >= 5 && record[1] == "Tax" {
			foundTax = true
			str := strings.Replace(record[4], "$", "", -1)
			str = strings.Replace(str, ",", "", -1)
			tax, err = strconv.ParseFloat(str, 64)
			if err != nil {
				return "", time.Time{}, 0.0, 0.0, fmt.Errorf("failed to parse tax: %v", err)
			}
		}
	}

	if !foundSubTotal || !foundTax {
		return "", time.Time{}, 0.0, 0.0, fmt.Errorf("failed to find sub-total or tax")
	}

	return storeName, parsedDate, subTotal, tax, nil
}

func processRestaurantDepotReceipts(db *gorm.DB) error {
	// iterate all files in receipts/unprocessed/restaurant_depot
	return iterateFiles("receipts/unprocessed/restaurant_depot", func(path string) error {
		log.Info("processRestaurantDepotReceipts: Processing file: ", path)

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		defer file.Close()

		// split the metadata from the product details by
		// finding and splitting the first line in the csv file that contains UPC
		scanner := bufio.NewScanner(file)
		var csvData, metaData string
		var foundUPC, foundSubTotal bool
		isMetadata := true

		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "UPC,") {
				isMetadata = false
				foundUPC = true
			}

			if strings.Contains(line, "Sub-Total") {
				isMetadata = true
				foundSubTotal = true
			}

			if isMetadata {
				metaData += line + "\n"
			} else {
				csvData += strings.ReplaceAll(line, "$", "") + "\n"
			}
		}

		if !foundUPC || !foundSubTotal {
			return fmt.Errorf("failed to find UPC or Sub-Total")
		}

		storeName, timestamp, subtotal, tax, err := processMetadata(metaData)
		if err != nil {
			return fmt.Errorf("failed to process metadata: %v", err)
		}

		store, err := models.FindOrCreateStore(storeName, db)
		if err != nil {
			return fmt.Errorf("failed to find or create store: %v", err)
		}

		// Parse the CSV data using gocsv.
		var dto_slice []models.PurchaseV2DTO
		if err = gocsv.UnmarshalString(csvData, &dto_slice); err != nil {
			return fmt.Errorf("failed to unmarshal csv data: %v", err)
		}

		meta, err := models.FindOrCreateMeta(store, timestamp, subtotal, tax, db)
		if err != nil {
			return fmt.Errorf("failed to find or create meta: %v", err)
		}

		var position uint = 0
		for _, dto := range dto_slice {
			if strings.Contains(dto.Description, "Balance") {
				log.Debug("Skipping purchase with description: ", dto.Description)
				continue
			}

			position += 1

			err = UpsertPurchase(position, &dto, meta, db)
			if err != nil {
				log.Errorf("failed to upsert product: %v", err)
				break
			}
		}

		if err := DeletePurchasesAfterPosition(meta, len(dto_slice), db); err != nil {
			log.Errorf("failed to delete purchases after position: %v", err)
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		// upsert each record into the database
		// move the file to receipts/processed/restaurant_depot
		// if product is not found, move the file to receipts/unprocessed/restaurant_depot/failed

		// write a function that moves csvFile to a new directory
		if err == nil {
			newDirectory := "receipts/processed/restaurant_depot"
			err = moveFileToDirectory(path, newDirectory)
			if err != nil {
				log.Errorf("failed to move file to directory: %v", err)
			}
		}

		return nil
	})
}

func handleToastItemSelectionDetail(path string, db *gorm.DB) error {
	// Parse the CSV data using gocsv.
	var records []models.ToastItemSelectionDetailDTO

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer f.Close()

	if err = gocsv.UnmarshalFile(f, &records); err != nil {
		return fmt.Errorf("failed to unmarshall file: %w", err)
	}

	for _, record := range records {
		err = UpsertItemSelectionDetail(&record, db)
		if err != nil {
			log.Errorf("failed to upsert order detail: %v", err)
			break
		}
	}

	// write a function that moves csvFile to a new directory
	if err == nil {
		newDirectory := "receipts/processed/toast"
		err = moveFileToDirectory(path, newDirectory)
		if err != nil {
			log.Errorf("failed to move file to directory: %v", err)
		}
	}

	return err
}

func handleToastOrderDetail(path string, db *gorm.DB) error {
	// Parse the CSV data using gocsv.
	var records []models.ToastOrderDetail

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer f.Close()

	if err = gocsv.UnmarshalFile(f, &records); err != nil {
		return fmt.Errorf("failed to unmarshall file: %w", err)
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
		err = moveFileToDirectory(path, newDirectory)
		if err != nil {
			log.Errorf("failed to move file to directory: %v", err)
		}
	}

	return err
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
	db.AutoMigrate(&models.MetaV2{})
	db.AutoMigrate(&models.PurchaseV2{})
	db.AutoMigrate(&models.PurchaseItem{})
	db.AutoMigrate(&models.ToastItemSelectionDetail{})
	db.AutoMigrate(&models.MenuItem{})
	db.AutoMigrate(&models.PurchaseItemGroup{})
	db.AutoMigrate(&models.PurchaseItem{})
	db.AutoMigrate(&models.Tag{})

	log.Info("Db setup complete!")
}

// for each file, parse the csv data using gocsv
// upsert each record into the database
// move the file to receipts/processed/toast

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

	models.PopulateItemsCache(db)
	models.PopulateMenuItemCache(db)

	if err := processRestaurantDepotReceipts(db); err != nil {
		panic(err)
	}

	// iterate all files in receipts/unprocessed/toast
	err = iterateFiles("receipts/unprocessed/toast", func(path string) error {
		log.Info("Processing file: ", path)

		filePrefix := strings.Split(filepath.Base(path), "_")

		if len(filePrefix) > 0 {
			// skip hidden files
			if strings.HasPrefix(filePrefix[0], ".") {
				return nil
			}

			switch filePrefix[0] {
			case "OrderDetails":
				return handleToastOrderDetail(path, db)
			case "ItemSelectionDetails":
				return handleToastItemSelectionDetail(path, db)
			default:
				return fmt.Errorf("unknown file prefix: %s", filePrefix[0])
			}
		}

		return nil
	})

	if err != nil {
		panic(err)
	}
}
