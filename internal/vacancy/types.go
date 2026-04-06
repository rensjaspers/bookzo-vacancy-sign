package vacancy

import "time"

type OverrideMode string

const (
	OverrideAuto        OverrideMode = "auto"
	OverrideAvailable   OverrideMode = "available"
	OverrideUnavailable OverrideMode = "unavailable"
)

type ThemeMode string

const (
	ThemeDark  ThemeMode = "dark"
	ThemeLight ThemeMode = "light"
	ThemeTime  ThemeMode = "time"
)

type Language struct {
	Available      string
	AvailableToday string
	Full           string
	FromPrefix     string
	DateLocale     string
}

type ThemeWindow struct {
	StartMinute int
	EndMinute   int
}

type Settings struct {
	HotelName         string
	TodayDateOverride string
	DayCount          int
	OverrideMode      OverrideMode
	ThemeMode         ThemeMode
	LightWindow       ThemeWindow
	PhraseRotate      time.Duration
	HeadlineScale     int
	Languages         []Language
	ShowErrorHint     bool
}

type CheckInAvailability struct {
	AnyRoomAvailable   bool
	FirstAvailableDate string
}

type LoadResult struct {
	AnyRoomAvailable   bool
	FirstAvailableDate string
	FetchedAt          time.Time
	FetchFailed        bool
}

type ViewModel struct {
	HotelName     string
	Headline      string
	Subline       string
	Dark          bool
	Available     bool
	ShowSubline   bool
	ShowErrorHint bool
	HeadlineScale int
}
