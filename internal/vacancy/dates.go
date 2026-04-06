package vacancy

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var englishMonths = []string{
	"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

var dutchMonths = []string{
	"januari", "februari", "maart", "april", "mei", "juni",
	"juli", "augustus", "september", "oktober", "november", "december",
}

var germanMonths = []string{
	"Januar", "Februar", "Marz", "April", "Mai", "Juni",
	"Juli", "August", "September", "Oktober", "November", "Dezember",
}

func StartOfDay(now time.Time) time.Time {
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, now.Location())
}

func AddDays(day time.Time, count int) time.Time {
	return day.AddDate(0, 0, count)
}

func FormatYMD(day time.Time) string {
	return day.Format("2006-01-02")
}

func ParseYMD(value string, loc *time.Location) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", value, loc)
}

func EffectiveToday(now time.Time, value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return StartOfDay(now), nil
	}
	return ParseYMD(value, now.Location())
}

func BuildTargetDates(ref time.Time, dayCount int) []string {
	targets := make([]string, 0, dayCount)
	for index := 0; index < dayCount; index++ {
		targets = append(targets, FormatYMD(AddDays(ref, index)))
	}
	return targets
}

func TargetDateSet(ref time.Time, dayCount int) map[string]struct{} {
	out := make(map[string]struct{}, dayCount)
	for _, day := range BuildTargetDates(ref, dayCount) {
		out[day] = struct{}{}
	}
	return out
}

func FormatDayMonth(ymd string, locale string) string {
	day, _ := ParseYMD(ymd, time.Local)
	return fmt.Sprintf("%d %s", day.Day(), monthName(day.Month(), locale))
}

func monthName(month time.Month, locale string) string {
	index := int(month) - 1
	return monthNames(locale)[index]
}

func monthNames(locale string) []string {
	switch normalizeLocale(locale) {
	case "nl":
		return dutchMonths
	case "de":
		return germanMonths
	default:
		return englishMonths
	}
}

func normalizeLocale(locale string) string {
	value := strings.ToLower(strings.TrimSpace(locale))
	parts := strings.Split(value, "-")
	return parts[0]
}

func ParseClock(value string) (int, error) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid clock value %q", value)
	}
	return parseClockParts(parts[0], parts[1])
}

func parseClockParts(hours string, minutes string) (int, error) {
	hour, err := strconv.Atoi(hours)
	if err != nil {
		return 0, err
	}
	return parseClockMinute(hour, minutes)
}

func parseClockMinute(hour int, minutes string) (int, error) {
	minute, err := strconv.Atoi(minutes)
	if err != nil {
		return 0, err
	}
	return clampClock(hour, minute)
}

func clampClock(hour int, minute int) (int, error) {
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return 0, fmt.Errorf("invalid clock time")
	}
	return hour*60 + minute, nil
}
