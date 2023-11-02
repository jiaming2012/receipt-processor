package models

import (
	"fmt"
	"jiaming2012/receipt-processor/custom"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PurchaseV2 struct {
	gorm.Model
	IsCase   bool    `gorm:"not null"`
	Price    float64 `gorm:"not null"`
	Quantity int     `gorm:"not null"`
	MetaId   uint    `gorm:"not null"`
	ItemId   uint    `gorm:"not null"`
	Position uint    `gorm:"not null"`
}

func (p PurchaseV2) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"is_case":  p.IsCase,
		"price":    p.Price,
		"quantity": p.Quantity,
		"meta_id":  p.MetaId,
		"item_id":  p.ItemId,
		"position": p.Position,
	}
}

type PurchasesV2 []PurchaseV2

func (purchase *PurchaseV2) Update(position uint, item *PurchaseItem, meta *MetaV2, db *gorm.DB) error {
	purchase.ItemId = item.ID
	purchase.MetaId = meta.ID
	purchase.Position = position

	tx := db.Save(purchase)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

// Total calculates the total number of cases, units and the sum of prices for a slice of PurchasesV2.
func (purchases PurchasesV2) Total() (uint, uint, float64) {
	var cases uint = 0
	var units uint = 0
	var sum float64 = 0
	for _, i := range purchases {
		if i.IsCase {
			cases += uint(i.Quantity)
		} else {
			units += uint(i.Quantity)
		}

		sum += i.Price
	}
	return units, cases, sum
}

type PurchaseV2DTO struct {
	SKU         string  `csv:"UPC"`
	Description string  `csv:"Description"`
	UnitQty     int     `csv:"UnitQty"`
	CaseQty     int     `csv:"CaseQty"`
	Price       float64 `csv:"Price"`
}

func (p PurchaseV2DTO) ConvertToPurchaseV2() *PurchaseV2 {
	return &PurchaseV2{
		IsCase:   p.CaseQty > 0,
		Price:    p.Price,
		Quantity: p.UnitQty + p.CaseQty,
	}
}

type Purchase struct {
	gorm.Model
	buffer      []string    `gorm:"-"`
	IsCase      bool        `gorm:"not null"`
	Price       float64     `gorm:"not null"`
	Quantity    int         `gorm:"not null"`
	Description Description `gorm:"-"`
	IsProcessed bool        `gorm:"-"`
	SKU         SKU         `gorm:"-"`
	MetaId      uint        `gorm:"not null"`
	ItemId      uint        `gorm:"not null"`
	Position    uint        `gorm:"not null"`
}

type Purchases []Purchase

func (purchases Purchases) Total() float64 {
	var sum float64 = 0
	for _, i := range purchases {
		sum += i.Price
	}
	return sum
}

// Count returns a tuple, which shows a count of (units, cases)
func (purchases Purchases) Count() (int, int) {
	var units int = 0
	var cases int = 0

	for _, i := range purchases {
		if i.IsCase {
			cases += i.Quantity
		} else {
			units += i.Quantity
		}
	}

	return units, cases
}

func (i Purchase) String() string {
	return fmt.Sprintf("%s (%d, %s) = $%.2f", i.Description, i.Quantity, i.SKU, i.Price)
}

func (i *Purchase) ProcessLine(line string, position uint) error {
	re := custom.ReceiptRegex["Purchase Delimiter"]

	if re.MatchString(line) {
		buf := formatBuffer(i.buffer)
		i.Description = Description(buf[0])
		i.Position = position

		dollarSignIndex := strings.LastIndex(buf[1], "$")
		isVoid := strings.LastIndex(buf[1], "-$") >= 0

		if dollarSignIndex < 0 {
			log.Fatalf("expected to find $ on line starting with: %s. Check that the $ is at the end of the line", buf[1])
		}

		data := regexp.MustCompile(`(\d+)`).FindAllString(buf[1], 1)
		i.SKU = SKU(data[0])

		price, err := strconv.ParseFloat(buf[1][dollarSignIndex+1:], 64)
		if err != nil {
			return fmt.Errorf("%v. Verify the receipt amount is a valid USD amount", err)
		}

		i.Price = price
		if isVoid {
			i.Price *= -1
		}

		data = re.FindStringSubmatch(line)
		var qtyStr string
		if len(data) == 3 {
			if strings.Index(data[0], "CASES") >= 0 {
				i.IsCase = true
				qtyStr = data[1]
			} else {
				i.IsCase = false
				qtyStr = data[2]
			}
		} else {
			log.Fatalf("unexpected data length %d, %v", len(data), data)
		}

		qty, err := strconv.Atoi(qtyStr)
		if err != nil {
			return nil
		}

		i.Quantity = qty
		i.IsProcessed = true

		return nil
	}

	i.buffer = append(i.buffer, line)
	return nil
}

func NewPurchase() *Purchase {
	return &Purchase{
		buffer:      make([]string, 0),
		IsProcessed: false,
	}
}

func parseSubtotal(line string) (*float64, error) {
	if strings.Index(line, "SUBTOTAL") >= 0 {
		dollarSignIndex := strings.Index(line, "$")
		if dollarSignIndex > 0 {
			amt, err := strconv.ParseFloat(line[dollarSignIndex+1:], 64)
			if err != nil {
				return nil, err
			}

			return &amt, nil
		}
	}

	return nil, nil
}

// formatBuffer applies business logic to mark the start
// of a new purchase
func formatBuffer(buf []string) []string {
	if len(buf) > 4 { // handles the first line of a receipt
		// add check that this can only occur once
		var data []string
		for i := len(buf) - 1; i >= 0; i-- {
			// we mark the start of a new item by finding empty line in the buffer
			// or when that the cashier created a subtotal in the middle of ringing
			// the purchases
			if buf[i] == "" || strings.Index(strings.ToLower(buf[i]), "subtotal") >= 0 {
				break
			}
			data = append(data, buf[i])
		}

		var result []string
		for i := len(data) - 1; i >= 0; i-- {
			result = append(result, data[i])
		}

		return result
	}

	return buf
}
