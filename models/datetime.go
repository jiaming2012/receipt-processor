package models

import (
	"database/sql/driver"
	"errors"
	"time"
)

type DateTime struct {
	time.Time
}

func (date *DateTime) MarshalCSV() (string, error) {
	return date.Time.Format("20060201"), nil
}

func (date *DateTime) UnmarshalCSV(csv string) (err error) {
	if len(csv) == 0 {
		date.Time = time.Time{}
		return nil
	}

	date.Time, err = time.Parse("1/2/06 3:04 PM", csv)
	return err
}

func (t *DateTime) Scan(value interface{}) error {
	if value == nil {
		*t = DateTime{}
		return nil
	}

	dbTime, ok := value.(time.Time)
	if !ok {
		return errors.New("invalid timestamp format from the database")
	}

	*t = DateTime{
		Time: dbTime,
	}
	return nil
}

func (t DateTime) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}
