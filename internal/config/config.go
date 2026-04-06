package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rensjaspers/bookzo-vacancy-sign/internal/vacancy"
)

type Config struct {
	APIKey             string
	RequestOrigin      string
	HotelName          string
	RoomObjectIDs      []int
	TodayDateOverride  string
	DayCount           int
	OverrideMode       vacancy.OverrideMode
	PollInterval       time.Duration
	PhraseRotate       time.Duration
	ThemeMode          vacancy.ThemeMode
	LightWindow        vacancy.ThemeWindow
	HeadlineScale      int
	ShowErrorHint      bool
	LogLevel           string
	FontPath           string
	Fullscreen         bool
	WindowWidth        int
	WindowHeight       int
	RequestConcurrency int
	Languages          []vacancy.Language
}

type rawConfig struct {
	APIKey               string         `json:"apiKey"`
	RequestOrigin        string         `json:"requestOrigin"`
	HotelName            string         `json:"hotelName"`
	RoomObjectIDs        []int          `json:"roomObjectIds"`
	TodayDateOverride    string         `json:"todayDateOverride"`
	ReferenceDate        string         `json:"referenceDate"`
	DayCount             int            `json:"dayCount"`
	OverrideMode         string         `json:"overrideMode"`
	PollInterval         string         `json:"pollInterval"`
	PhraseRotateInterval string         `json:"phraseRotateInterval"`
	ThemeMode            string         `json:"themeMode"`
	LightWindow          rawLightWindow `json:"lightWindow"`
	HeadlineScale        int            `json:"headlineScale"`
	ShowErrorHint        bool           `json:"showErrorHint"`
	LogLevel             string         `json:"logLevel"`
	FontPath             string         `json:"fontPath"`
	Fullscreen           bool           `json:"fullscreen"`
	WindowWidth          int            `json:"windowWidth"`
	WindowHeight         int            `json:"windowHeight"`
	RequestConcurrency   int            `json:"requestConcurrency"`
	Languages            []rawLanguage  `json:"languages"`
}

type rawLightWindow struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type rawLanguage struct {
	Available      string `json:"available"`
	AvailableToday string `json:"availableToday"`
	Full           string `json:"full"`
	FromPrefix     string `json:"fromPrefix"`
	DateLocale     string `json:"dateLocale"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	return LoadBytes(data, filepath.Dir(path))
}

func LoadBytes(data []byte, baseDir string) (Config, error) {
	cfg, err := Parse(data)
	if err != nil {
		return Config{}, err
	}
	cfg = resolvePaths(cfg, baseDir)
	if cfg.FontPath == "" {
		return Config{}, fmt.Errorf("no usable font found")
	}
	return cfg, nil
}

func Parse(data []byte) (Config, error) {
	raw := defaultRawConfig()
	if err := json.Unmarshal(data, &raw); err != nil {
		return Config{}, err
	}
	return validate(raw)
}

func defaultRawConfig() rawConfig {
	return rawConfig{
		RequestOrigin:        "",
		HotelName:            "Your Hotel",
		RoomObjectIDs:        []int{1, 3, 21, 4, 2, 6, 8, 10},
		DayCount:             2,
		OverrideMode:         string(vacancy.OverrideAuto),
		PollInterval:         "15m",
		PhraseRotateInterval: "5s",
		ThemeMode:            string(vacancy.ThemeDark),
		LightWindow:          rawLightWindow{Start: "07:00", End: "19:00"},
		HeadlineScale:        80,
		ShowErrorHint:        true,
		LogLevel:             "info",
		Fullscreen:           true,
		WindowWidth:          1920,
		WindowHeight:         1080,
		RequestConcurrency:   2,
		Languages:            defaultLanguages(),
	}
}

func defaultLanguages() []rawLanguage {
	return []rawLanguage{
		{Available: "KAMERS VRIJ", AvailableToday: "VANDAAG KAMERS VRIJ", Full: "VOL", FromPrefix: "vanaf", DateLocale: "nl"},
		{Available: "ZIMMER FREI", AvailableToday: "HEUTE ZIMMER FREI", Full: "VOLL", FromPrefix: "ab", DateLocale: "de"},
		{Available: "ROOMS AVAILABLE", AvailableToday: "ROOMS AVAILABLE TODAY", Full: "FULL", FromPrefix: "from", DateLocale: "en"},
	}
}

func validate(raw rawConfig) (Config, error) {
	if err := validateRequired(raw); err != nil {
		return Config{}, err
	}
	return buildConfig(raw)
}

func validateRequired(raw rawConfig) error {
	if strings.TrimSpace(raw.APIKey) == "" {
		return fmt.Errorf("apiKey is required")
	}
	return validateRequestOrigin(raw.RequestOrigin)
}

func validateRequestOrigin(value string) error {
	origin := strings.TrimSpace(value)
	if origin == "" {
		return fmt.Errorf("requestOrigin is required (HTTPS URL of the public Bookzo booking site)")
	}
	if disallowedBookzoOrigin(origin) {
		return fmt.Errorf("requestOrigin must be the real booking site URL registered with Bookzo, not %q", origin)
	}
	return nil
}

func disallowedBookzoOrigin(origin string) bool {
	return origin == "https://example.com" || origin == "http://example.com"
}

func buildConfig(raw rawConfig) (Config, error) {
	poll, err := time.ParseDuration(raw.PollInterval)
	if err != nil {
		return Config{}, err
	}
	return withPhraseDuration(raw, poll)
}

func withPhraseDuration(raw rawConfig, poll time.Duration) (Config, error) {
	rotate, err := time.ParseDuration(raw.PhraseRotateInterval)
	if err != nil {
		return Config{}, err
	}
	return withLightWindow(raw, poll, rotate)
}

func withLightWindow(raw rawConfig, poll time.Duration, rotate time.Duration) (Config, error) {
	window, err := parseLightWindow(raw.LightWindow)
	if err != nil {
		return Config{}, err
	}
	return withLanguages(raw, poll, rotate, window)
}

func withLanguages(
	raw rawConfig,
	poll time.Duration,
	rotate time.Duration,
	window vacancy.ThemeWindow,
) (Config, error) {
	languages, err := parseLanguages(raw.Languages)
	if err != nil {
		return Config{}, err
	}
	return buildFinalConfig(raw, poll, rotate, window, languages)
}

func buildFinalConfig(
	raw rawConfig,
	poll time.Duration,
	rotate time.Duration,
	window vacancy.ThemeWindow,
	languages []vacancy.Language,
) (Config, error) {
	todayDateOverride := todayDateOverride(raw)
	if err := validateTodayDateOverride(todayDateOverride); err != nil {
		return Config{}, err
	}
	return Config{
		APIKey:             strings.TrimSpace(raw.APIKey),
		RequestOrigin:      strings.TrimSpace(raw.RequestOrigin),
		HotelName:          strings.TrimSpace(raw.HotelName),
		RoomObjectIDs:      cleanRoomIDs(raw.RoomObjectIDs),
		TodayDateOverride:  todayDateOverride,
		DayCount:           clampDayCount(raw.DayCount),
		OverrideMode:       parseOverride(raw.OverrideMode),
		PollInterval:       poll,
		PhraseRotate:       rotate,
		ThemeMode:          parseTheme(raw.ThemeMode),
		LightWindow:        window,
		HeadlineScale:      clampHeadlineScale(raw.HeadlineScale),
		ShowErrorHint:      raw.ShowErrorHint,
		LogLevel:           normalizeLogLevel(raw.LogLevel),
		FontPath:           strings.TrimSpace(raw.FontPath),
		Fullscreen:         raw.Fullscreen,
		WindowWidth:        clampWindow(raw.WindowWidth, 1920),
		WindowHeight:       clampWindow(raw.WindowHeight, 1080),
		RequestConcurrency: clampConcurrency(raw.RequestConcurrency),
		Languages:          languages,
	}, nil
}

func validateTodayDateOverride(value string) error {
	if value == "" {
		return nil
	}
	_, err := vacancy.EffectiveToday(time.Now(), value)
	return err
}

func todayDateOverride(raw rawConfig) string {
	value := strings.TrimSpace(raw.TodayDateOverride)
	if value != "" {
		return value
	}
	return strings.TrimSpace(raw.ReferenceDate)
}

func parseLightWindow(raw rawLightWindow) (vacancy.ThemeWindow, error) {
	start, err := vacancy.ParseClock(raw.Start)
	if err != nil {
		return vacancy.ThemeWindow{}, err
	}
	return parseLightWindowEnd(raw.End, start)
}

func parseLightWindowEnd(value string, start int) (vacancy.ThemeWindow, error) {
	end, err := vacancy.ParseClock(value)
	if err != nil {
		return vacancy.ThemeWindow{}, err
	}
	return vacancy.ThemeWindow{StartMinute: start, EndMinute: end}, nil
}

func parseLanguages(raw []rawLanguage) ([]vacancy.Language, error) {
	if len(raw) == 0 {
		raw = defaultLanguages()
	}
	out := make([]vacancy.Language, 0, len(raw))
	for _, language := range raw {
		out = append(out, normalizeLanguage(language))
	}
	return validateLanguages(out)
}

func normalizeLanguage(language rawLanguage) vacancy.Language {
	return vacancy.Language{
		Available:      strings.TrimSpace(language.Available),
		AvailableToday: strings.TrimSpace(language.AvailableToday),
		Full:           strings.TrimSpace(language.Full),
		FromPrefix:     strings.TrimSpace(language.FromPrefix),
		DateLocale:     strings.TrimSpace(language.DateLocale),
	}
}

func validateLanguages(languages []vacancy.Language) ([]vacancy.Language, error) {
	for _, language := range languages {
		if err := validateLanguage(language); err != nil {
			return nil, err
		}
	}
	return languages, nil
}

func validateLanguage(language vacancy.Language) error {
	if language.Available == "" || language.AvailableToday == "" {
		return fmt.Errorf("language phrases are required")
	}
	if language.Full == "" || language.FromPrefix == "" {
		return fmt.Errorf("language phrases are required")
	}
	return nil
}

func cleanRoomIDs(values []int) []int {
	out := []int{}
	for _, value := range values {
		if value > 0 {
			out = append(out, value)
		}
	}
	return out
}

func clampDayCount(value int) int {
	if value < 1 {
		return 1
	}
	if value > 30 {
		return 30
	}
	return value
}

func clampHeadlineScale(value int) int {
	if value < 40 {
		return 40
	}
	if value > 100 {
		return 100
	}
	return value
}

func clampWindow(value int, fallback int) int {
	if value < 320 {
		return fallback
	}
	return value
}

func clampConcurrency(value int) int {
	if value < 1 {
		return 1
	}
	if value > 4 {
		return 4
	}
	return value
}

func parseOverride(value string) vacancy.OverrideMode {
	switch strings.TrimSpace(value) {
	case string(vacancy.OverrideAvailable):
		return vacancy.OverrideAvailable
	case string(vacancy.OverrideUnavailable):
		return vacancy.OverrideUnavailable
	default:
		return vacancy.OverrideAuto
	}
}

func parseTheme(value string) vacancy.ThemeMode {
	switch strings.TrimSpace(value) {
	case string(vacancy.ThemeLight):
		return vacancy.ThemeLight
	case string(vacancy.ThemeTime):
		return vacancy.ThemeTime
	default:
		return vacancy.ThemeDark
	}
}

func normalizeLogLevel(value string) string {
	level := strings.ToLower(strings.TrimSpace(value))
	if level == "debug" || level == "warn" || level == "error" {
		return level
	}
	return "info"
}

func resolvePaths(cfg Config, baseDir string) Config {
	cfg.FontPath = resolveFontPath(baseDir, cfg.FontPath)
	return cfg
}

func resolvePath(baseDir string, value string) string {
	if value == "" || filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(baseDir, value)
}

func resolveFontPath(baseDir string, value string) string {
	if value != "" {
		return resolvePath(baseDir, value)
	}
	return detectFontPath(baseDir)
}

func detectFontPath(baseDir string) string {
	for _, candidate := range fontCandidates(baseDir) {
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

func fontCandidates(baseDir string) []string {
	candidates := []string{
		filepath.Join(baseDir, "fonts", "board.ttf"),
		"/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
	}
	if runtime.GOOS == "darwin" {
		return append(candidates,
			"/System/Library/Fonts/Supplemental/Arial.ttf",
			"/System/Library/Fonts/Supplemental/Arial Bold.ttf",
			"/System/Library/Fonts/Supplemental/Helvetica.ttc",
			"/Library/Fonts/Arial.ttf",
		)
	}
	if runtime.GOOS == "windows" {
		return append(candidates,
			`C:\Windows\Fonts\arial.ttf`,
			`C:\Windows\Fonts\arialbd.ttf`,
			`C:\Windows\Fonts\segoeui.ttf`,
		)
	}
	return candidates
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
