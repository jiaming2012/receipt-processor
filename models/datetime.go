package models

import (
	"database/sql/driver"
	"errors"
	"time"
)

type DateTimeEST struct {
	time.Time
}

func (date *DateTimeEST) MarshalCSV() (string, error) {
	return date.Time.Format("20060201"), nil
}

func (date *DateTimeEST) UnmarshalCSV(csv string) (err error) {
	if len(csv) == 0 {
		date.Time = time.Time{}
		return nil
	}

	est, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}

	date.Time, err = time.ParseInLocation("1/2/06 3:04 PM", csv, est)
	return err
}

func (t *DateTimeEST) Scan(value interface{}) error {
	if value == nil {
		*t = DateTimeEST{}
		return nil
	}

	dbTime, ok := value.(time.Time)
	if !ok {
		return errors.New("invalid timestamp format from the database")
	}

	*t = DateTimeEST{
		Time: dbTime,
	}
	return nil
}

func (t DateTimeEST) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}
