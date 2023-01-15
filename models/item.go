package models

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"regexp"
	"strconv"
	"strings"
)

type Item struct {
	gorm.Model
	buffer      []string `gorm:"-"`
	IsCase      bool
	Price       float64
	Quantity    uint
	Description string
	IsProcessed bool `gorm:"-"`
	SKU         string
	MetaId      uint
	Position    uint
}

type Items []Item

func (items Items) Total() float64 {
	var sum float64 = 0
	for _, i := range items {
		sum += i.Price
	}
	return sum
}

// Count returns a tuple, which shows a count of (units, cases)
func (items Items) Count() (uint, uint) {
	var units uint = 0
	var cases uint = 0

	for _, i := range items {
		if i.IsCase {
			cases += i.Quantity
		} else {
			units += i.Quantity
		}
	}

	return units, cases
}

func (i Item) String() string {
	return fmt.Sprintf("%s (%d, %s) = $%.2f", i.Description, i.Quantity, i.SKU, i.Price)
}

func (i *Item) ProcessLine(line string, position uint) error {
	re := regexp.MustCompile(`(?:(?:CASES)\s(\d+)\s)?(?:UNITS)\s(\d+)`)

	if re.MatchString(line) {
		buf := formatBuffer(i.buffer)
		i.Description = buf[0]
		i.Position = position

		dollarSignIndex := strings.LastIndex(buf[1], "$")
		if dollarSignIndex < 0 {
			log.Fatalf("expected to find $ on %s", buf[1])
		}

		data := regexp.MustCompile(`(\d+)`).FindAllString(buf[1], 1)
		i.SKU = data[0]

		price, err := strconv.ParseFloat(buf[1][dollarSignIndex+1:], 64)
		if err != nil {
			return err
		}

		i.Price = price

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

		i.Quantity = uint(qty)
		i.IsProcessed = true

		return nil
	}

	i.buffer = append(i.buffer, line)
	return nil
}

func NewItem() *Item {
	return &Item{
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

// formatBuffer applies business logic to remove irrelevant items
// from the buffer
func formatBuffer(buf []string) []string {
	if len(buf) > 4 { // handles the first line of a receipt
		// add check that this can only occur once
		var data []string
		for i := len(buf) - 1; i >= 0; i-- {
			if buf[i] == "" {
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
