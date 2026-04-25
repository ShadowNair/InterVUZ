package domain

import (
	"strings"
	"unicode"
)

func NormalizeRoomName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "аудитория")
	value = strings.TrimPrefix(value, "ауд.")
	value = strings.TrimPrefix(value, "ауд")
	value = strings.TrimSpace(value)

	return normalizeToken(value)
}

func NormalizeBuildingName(value string) string {
	return normalizeToken(strings.ToLower(strings.TrimSpace(value)))
}

func normalizeToken(value string) string {
	var builder strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}
