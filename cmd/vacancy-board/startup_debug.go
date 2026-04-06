package main

import (
	"fmt"

	"github.com/rensjaspers/bookzo-vacancy-sign/internal/config"
)

func debugStartupAttrs(cfg config.Config) []any {
	a := debugTimingAttrs(cfg)
	return append(a, debugDisplayAttrs(cfg)...)
}

func debugTimingAttrs(cfg config.Config) []any {
	return []any{
		"hotelName", cfg.HotelName,
		"pollInterval", cfg.PollInterval.String(),
		"phraseRotate", cfg.PhraseRotate.String(),
		"dayCount", cfg.DayCount,
		"themeMode", string(cfg.ThemeMode),
		"overrideMode", string(cfg.OverrideMode),
	}
}

func debugDisplayAttrs(cfg config.Config) []any {
	return []any{
		"requestConcurrency", cfg.RequestConcurrency,
		"roomObjectCount", len(cfg.RoomObjectIDs),
		"headlineScale", cfg.HeadlineScale,
		"fullscreen", cfg.Fullscreen,
		"window", fmt.Sprintf("%dx%d", cfg.WindowWidth, cfg.WindowHeight),
		"languages", len(cfg.Languages),
		"hasDateOverride", cfg.TodayDateOverride != "",
	}
}
