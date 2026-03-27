package utils

import (
	"fmt"
	"strings"
	"time"
)

func FormatMoney(amount float64, currency string) string {
	if currency == "" {
		currency = "AUD"
	}
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}

	whole := int64(amount)
	frac := int64((amount - float64(whole)) * 100 + 0.5)

	parts := []string{}
	if whole == 0 {
		parts = append(parts, "0")
	} else {
		for whole > 0 {
			group := whole % 1000
			whole /= 1000
			if whole > 0 {
				parts = append([]string{fmt.Sprintf("%03d", group)}, parts...)
			} else {
				parts = append([]string{fmt.Sprintf("%d", group)}, parts...)
			}
		}
	}

	return fmt.Sprintf("%s$%s.%02d", sign, strings.Join(parts, ","), frac)
}

func FormatUnixDate(ts int64) string {
	if ts == 0 {
		return ""
	}
	t := time.Unix(ts, 0)
	return t.Format("02 Jan 2006")
}

func FormatUnixDateTime(ts int64) string {
	if ts == 0 {
		return ""
	}
	t := time.Unix(ts, 0)
	return t.Format("02 Jan 2006 15:04")
}

func FormatDate(dateStr string) string {
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
		"02/01/2006",
	} {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t.Format("02 Jan 2006")
		}
	}
	return dateStr
}

func FormatDateTime(dateStr string) string {
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
	} {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t.Format("02 Jan 2006 15:04")
		}
	}
	return dateStr
}

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
