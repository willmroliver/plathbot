package util

import "time"

func LastMonday(from *time.Time) time.Time {
	diff := int(from.Weekday() - time.Monday)
	if diff < 0 {
		diff += 7
	}

	date := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	return date.AddDate(0, 0, -diff)
}

func FirstOfMonth(from *time.Time) time.Time {
	diff := int(from.Day() - 1)

	date := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	return date.AddDate(0, 0, diff)
}
