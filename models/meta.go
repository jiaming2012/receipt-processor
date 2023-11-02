package models

import (
	"fmt"
	"jiaming2012/receipt-processor/custom"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type MetaV2 struct {
	gorm.Model
	StoreID    uint       `gorm:"not null"`
	Store      *Store     `gorm:"foreignKey:StoreID;references:ID;not null"`
	StoreId    uint       `gorm:"uniqueIndex:compositeMeta;not null"` // created for createing indexes
	Timestamp  *time.Time `gorm:"uniqueIndex:compositeMeta;not null"`
	TotalUnits *uint
	TotalCases *uint
	Subtotal   *float64
	Tax        *float64
}

func (m *MetaV2) Update(purchases PurchasesV2, db *gorm.DB) error {
	totalUnits, totalCases, subtotal := purchases.Total()

	m.TotalUnits = &totalUnits
	m.TotalCases = &totalCases
	m.Subtotal = &subtotal

	tx := db.Save(m)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func FindOrCreateMeta(store *Store, ts time.Time, subtotal float64, tax float64, db *gorm.DB) (*MetaV2, error) {
	var meta MetaV2

	if tx := db.Find(&meta, "store_id = ? AND timestamp = ?", store.ID, ts); tx.Error == nil {
		if tx.RowsAffected == 0 {
			meta.Store = store
			meta.Timestamp = &ts
			meta.Subtotal = &subtotal
			meta.Tax = &tax

			tx = db.Save(&meta)
			if tx.Error != nil {
				return nil, tx.Error
			}
		}
	}

	return &meta, nil
}

type Meta struct {
	gorm.Model
	StoreName   string     `gorm:"-"`
	StoreId     *uint      `gorm:"uniqueIndex:compositeMeta;not null"`
	Timestamp   *time.Time `gorm:"uniqueIndex:compositeMeta;not null"`
	IsProcessed bool       `gorm:"-"`
	TotalUnits  *uint      `gorm:"not null"`
	TotalCases  *uint      `gorm:"not null"`
	TotalItems  *uint      `gorm:"not null"`
	Subtotal    *float64   `gorm:"not null"`
}

func (m Meta) String() string {
	return fmt.Sprintf("%s: %v items for $%v", m.StoreName, *m.TotalItems, *m.Subtotal)
}

func (m *Meta) ProcessLine(line string, db *gorm.DB) error {
	if len(m.StoreName) == 0 {
		if storeName := formattedStoreName(line); len(storeName) > 0 {
			m.StoreName = storeName

			if store, dbErr := FindOrCreateStore(StoreName(m.StoreName), db); dbErr == nil {
				m.StoreId = &store.ID
			} else {
				return dbErr
			}
		}
	}

	if len(m.StoreName) > 0 && m.Timestamp == nil {
		re := custom.ReceiptRegex[m.StoreName]
		matches := re.FindAllString(line, 1)
		if len(matches) == 1 {
			ts, err := time.Parse(custom.DateTimeParse[m.StoreName], matches[0])
			if err != nil {
				return err
			}
			m.Timestamp = &ts
		}
	}

	// parse total units
	matches := custom.ReceiptRegex["Total Units"].FindStringSubmatch(line)
	if len(matches) > 1 {
		val, err := strconv.Atoi(matches[1])
		if err != nil {
			return err
		}
		units := uint(val)
		m.TotalUnits = &units
	}

	// parse total cases
	matches = custom.ReceiptRegex["Total Cases"].FindStringSubmatch(line)
	if len(matches) > 1 {
		val, err := strconv.Atoi(matches[1])
		if err != nil {
			return err
		}
		cases := uint(val)
		m.TotalCases = &cases
	}

	// parse total items
	matches = custom.ReceiptRegex["Total Purchases"].FindStringSubmatch(line)
	if len(matches) > 1 {
		val, err := strconv.Atoi(matches[1])
		if err != nil {
			return err
		}
		purchases := uint(val)
		m.TotalItems = &purchases
	}

	// add additional items
	matches = custom.ReceiptRegex["Additional Purchases"].FindStringSubmatch(line)
	if len(matches) > 1 {
		if m.TotalItems != nil {
			val, err := strconv.Atoi(matches[1])
			if err != nil {
				return err
			}
			purchases := uint(val)
			*m.TotalItems += purchases
		}
	}

	// parse subtotal
	val, err := parseSubtotal(line)
	if err != nil {
		return err
	}

	if val != nil {
		m.Subtotal = val
	}

	if len(m.Unprocessed()) == 0 {
		m.IsProcessed = true
	}

	return nil
}

func (m *Meta) Unprocessed() []string {
	var result []string

	if m.Timestamp == nil {
		result = append(result, "Could not find a timestamp on the receipt.")
	}
	if m.TotalUnits == nil {
		result = append(result, "Could not find \"TOTAL UNITS ENTERED\" on the receipt, please make sure there is a number after \"TOTAL UNITS ENTERED\".")
	}
	if m.TotalItems == nil {
		result = append(result, "Could not find \"TOTAL ITEMS RUNG UP\" on the receipt, please make sure there is a number after \"TOTAL ITEMS RUNG UP\".")
	}
	if m.TotalCases == nil {
		result = append(result, "Could not find \"TOTAL CASES ENTERED\" on the receipt, please make sure there is a number after \"TOTAL CASES ENTERED\".")
	}
	if m.Subtotal == nil {
		result = append(result, "Could not find the subtotal on the receipt.")
	}
	if m.StoreId == nil {
		result = append(result, "Could not find the store name on the receipt.")
	}

	return result
}

func formattedStoreName(line string) string {
	line = strings.ToLower(line)

	if strings.Index(line, "restaurant") >= 0 && strings.Index(line, "depot") >= 0 {
		return "Restaurant Depot"
	}

	return ""
}
