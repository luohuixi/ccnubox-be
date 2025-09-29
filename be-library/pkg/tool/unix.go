package tool

import (
	"time"
)

func ParseToUnix(timeStr string) (int, error) {
	t, err := time.Parse("2006-01-02 15:04", timeStr)
	hhmm := t.Hour()*100 + t.Minute()
	return hhmm, err
}
