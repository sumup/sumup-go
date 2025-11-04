// Package datetime provides custom date and time types with JSON marshaling support.
package datetime

import (
	"encoding/json"
	"time"
)

// Date represents a date without time using RFC3339 date format (YYYY-MM-DD).
type Date struct{ time.Time }

// String returns the date formatted as YYYY-MM-DD.
func (d Date) String() string {
	return d.Format(time.DateOnly)
}

const jsonDateFormat = `"` + time.DateOnly + `"`

var _ json.Unmarshaler = (*Date)(nil)

// UnmarshalJSON parses a JSON string in YYYY-MM-DD format.
func (d *Date) UnmarshalJSON(b []byte) (err error) {
	date, err := time.Parse(jsonDateFormat, string(b))
	if err != nil {
		return err
	}
	d.Time = date
	return
}

var _ json.Marshaler = (*Date)(nil)

// MarshalJSON formats the date as a JSON string in YYYY-MM-DD format.
func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(d.Format(jsonDateFormat)), nil
}

// Time represents a time of day using RFC3339 time format (HH:MM:SS).
type Time struct{ time.Time }

// String returns the time formatted as HH:MM:SS.
func (t Time) String() string {
	return t.Format(time.TimeOnly)
}

const jsonTimeFormat = `"` + time.TimeOnly + `"`

var _ json.Unmarshaler = (*Time)(nil)

// UnmarshalJSON parses a JSON string in HH:MM:SS format.
func (t *Time) UnmarshalJSON(b []byte) (err error) {
	date, err := time.Parse(jsonTimeFormat, string(b))
	if err != nil {
		return err
	}
	t.Time = date
	return
}

var _ json.Marshaler = (*Time)(nil)

// MarshalJSON formats the time as a JSON string in HH:MM:SS format.
func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(t.Format(jsonTimeFormat)), nil
}
