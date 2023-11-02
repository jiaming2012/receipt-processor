package models

import (
	"time"

	"gorm.io/gorm"
)

type DayOfWeek int

const (
	Sunday DayOfWeek = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

// find a weather api
type SalesDay struct {
	gorm.Model
	Date          time.Time
	DayOfWeek     DayOfWeek
	Temperature   float64
	Precipitation string
	WindSpeed     float64
	WeekInMonth   int
	MonthInYear   int
	IsHoliday     bool
}
