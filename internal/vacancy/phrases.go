package vacancy

import "time"

func PhraseIndex(now time.Time, step time.Duration, total int) int {
	if total < 1 || step <= 0 {
		return 0
	}
	return int(now.UnixMilli()/step.Milliseconds()) % total
}

func LanguageAt(languages []Language, index int) Language {
	if len(languages) == 0 {
		return defaultLanguage()
	}
	return languages[index%len(languages)]
}

func defaultLanguage() Language {
	return Language{
		Available:      "ROOMS AVAILABLE",
		AvailableToday: "ROOMS AVAILABLE TODAY",
		Full:           "FULL",
		FromPrefix:     "from",
		DateLocale:     "en",
	}
}

func HeadlineFor(language Language, available bool, today bool) string {
	if !available {
		return language.Full
	}
	if today {
		return language.AvailableToday
	}
	return language.Available
}

func SublineFor(language Language, firstAvailable string) string {
	if firstAvailable == "" {
		return ""
	}
	return language.FromPrefix + " " + FormatDayMonth(firstAvailable, language.DateLocale)
}
