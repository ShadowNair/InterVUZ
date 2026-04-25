package news

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

var months = map[string]time.Month{
	"января":   time.January,
	"февраля":  time.February,
	"марта":    time.March,
	"апреля":   time.April,
	"мая":      time.May,
	"июня":     time.June,
	"июля":     time.July,
	"августа":  time.August,
	"сентября": time.September,
	"октября":  time.October,
	"ноября":   time.November,
	"декабря":  time.December,
}

var monthNames = map[time.Month]string{
	time.January:   "января",
	time.February:  "февраля",
	time.March:     "марта",
	time.April:     "апреля",
	time.May:       "мая",
	time.June:      "июня",
	time.July:      "июля",
	time.August:    "августа",
	time.September: "сентября",
	time.October:   "октября",
	time.November:  "ноября",
	time.December:  "декабря",
}

func ToNewsRecord(item domain.NewsItem) (domain.NewsRecord, error) {
	publishedDate, err := parsePublishedAt(item.PublishedAt)
	if err != nil {
		return domain.NewsRecord{}, fmt.Errorf("parse published_at for slug=%s: %w", item.Slug, err)
	}

	return domain.NewsRecord{
		Slug:          item.Slug,
		Title:         item.Title,
		PreviewText:   item.PreviewText,
		PublishedDate: publishedDate,
		ImagePreview:  item.ImagePreview,
		PageURL:       item.PageURL,
		Tags:          item.Tags,
	}, nil
}

func ToNewsItem(record domain.NewsRecord) domain.NewsItem {
	return domain.NewsItem{
		Slug:         record.Slug,
		Title:        record.Title,
		PreviewText:  record.PreviewText,
		PublishedAt:  toPublishedAt(record.PublishedDate),
		ImagePreview: record.ImagePreview,
		Tags:         record.Tags,
		PageURL:      record.PageURL,
	}
}

func parsePublishedAt(p domain.PublishedAt) (time.Time, error) {
	day, err := strconv.Atoi(strings.TrimSpace(p.Day))
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid day: %w", err)
	}

	year, err := strconv.Atoi(strings.TrimSpace(p.Year))
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid year: %w", err)
	}

	monthName := strings.ToLower(strings.TrimSpace(p.Month))
	month, ok := months[monthName]
	if !ok {
		return time.Time{}, fmt.Errorf("unknown month: %s", p.Month)
	}

	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC), nil
}

func toPublishedAt(value time.Time) domain.PublishedAt {
	month := monthNames[value.Month()]
	if month == "" {
		month = strings.ToLower(value.Month().String())
	}

	return domain.PublishedAt{
		Day:   strconv.Itoa(value.Day()),
		Month: month,
		Year:  strconv.Itoa(value.Year()),
	}
}
