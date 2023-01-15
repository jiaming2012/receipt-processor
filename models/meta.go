package models

import (
	"gorm.io/gorm"
	"jiaming2012/receipt-processor/custom"
	"strconv"
	"time"
)

type Meta struct {
	gorm.Model
	Item        Item       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	StoreName   string     `gorm:"uniqueIndex:composite;not null"`
	Timestamp   *time.Time `gorm:"uniqueIndex:composite;not null"`
	IsProcessed bool       `gorm:"-"`
	TotalUnits  *uint
	TotalCases  *uint
	TotalItems  *uint
	Subtotal    *float64
}

func (m *Meta) ProcessLine(line string) error {
	if m.IsProcessed {
		return nil
	}

	if line == "Restaurant Depot" {
		m.StoreName = line
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

	if m.TotalUnits == nil {
		matches := custom.ReceiptRegex["Total Units"].FindStringSubmatch(line)
		if len(matches) > 1 {
			val, err := strconv.Atoi(matches[1])
			if err != nil {
				return err
			}
			units := uint(val)
			m.TotalUnits = &units
		}
	}

	if m.TotalCases == nil {
		matches := custom.ReceiptRegex["Total Cases"].FindStringSubmatch(line)
		if len(matches) > 1 {
			val, err := strconv.Atoi(matches[1])
			if err != nil {
				return err
			}
			cases := uint(val)
			m.TotalCases = &cases
		}
	}

	if m.TotalItems == nil {
		matches := custom.ReceiptRegex["Total Items"].FindStringSubmatch(line)
		if len(matches) > 1 {
			val, err := strconv.Atoi(matches[1])
			if err != nil {
				return err
			}
			items := uint(val)
			m.TotalItems = &items
		}
	}

	if m.Subtotal == nil {
		val, err := parseSubtotal(line)
		if err != nil {
			return err
		}

		if val != nil {
			m.Subtotal = val
		}
	}

	if m.Timestamp != nil && m.TotalUnits != nil && m.TotalItems != nil && m.TotalCases != nil && m.Subtotal != nil {
		m.IsProcessed = true
	}

	return nil
}
