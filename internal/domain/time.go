package domain

import "time"

func AcademicWeekType(date time.Time, week1Start time.Time) string {
	days := int(dateOnly(date).Sub(dateOnly(week1Start)).Hours() / 24)
	weekIndex := floorDiv(days, 7)
	if weekIndex%2 == 0 {
		return "ch"
	}

	return "zn"
}

func TimeRangesOverlap(startA time.Time, endA time.Time, startB time.Time, endB time.Time) bool {
	return startA.Before(endB) && startB.Before(endA)
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func floorDiv(value int, divisor int) int {
	quotient := value / divisor
	remainder := value % divisor
	if remainder != 0 && ((remainder < 0) != (divisor < 0)) {
		quotient--
	}

	return quotient
}
