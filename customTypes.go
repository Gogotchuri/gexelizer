package gexelizer

import (
	"fmt"
	"time"
)

type GexValuer interface {
	GexelizerValue() any
}

type Date string

func (d Date) String() string {
	return string(d)
}
func (d Date) ToTime() (time.Time, error) {
	timeFormats := []string{
		"2006-01-02",
		"2006/01/02",
		"2006.01.02",
		"02-01-2006",
		"02/01/2006",
		"02.01.2006",
		"02-01-06",
		"02/01/06",
		"02.01.06",
	}

	for _, format := range timeFormats {
		t, err := time.Parse(format, string(d)[:10])
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse date %s - unknown format", string(d))
}

func DateFromTime(t time.Time) Date {
	return Date(t.Format("2006-01-02"))
}
