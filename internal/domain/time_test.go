package domain

import (
	"testing"
	"time"
)

func TestAcademicWeekType(t *testing.T) {
	week1Start := mustDate(t, "2026-02-09")

	tests := []struct {
		name string
		date string
		want string
	}{
		{name: "first week is numerator", date: "2026-02-09", want: "ch"},
		{name: "second week is denominator", date: "2026-02-16", want: "zn"},
		{name: "third week returns numerator", date: "2026-02-23", want: "ch"},
		{name: "date inside first week", date: "2026-02-14", want: "ch"},
		{name: "week before start alternates backwards", date: "2026-02-08", want: "zn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AcademicWeekType(mustDate(t, tt.date), week1Start)
			if got != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, got)
			}
		})
	}
}

func TestTimeRangesOverlap(t *testing.T) {
	start := time.Date(2026, time.February, 9, 10, 0, 0, 0, time.UTC)
	end := start.Add(90 * time.Minute)

	if !TimeRangesOverlap(start, end, start.Add(30*time.Minute), end.Add(time.Hour)) {
		t.Fatal("expected overlapping ranges")
	}
	if TimeRangesOverlap(start, end, end, end.Add(time.Hour)) {
		t.Fatal("adjacent ranges must not overlap")
	}
}

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}

	return parsed
}
