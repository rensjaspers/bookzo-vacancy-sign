package config

import "testing"

func TestParseUsesDefaultsAndClamps(t *testing.T) {
	data := []byte(`{"apiKey":"key","dayCount":99,"headlineScale":10}`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	assertDefaultConfig(t, cfg)
}

func TestParseAllowsMissingFontPath(t *testing.T) {
	data := []byte(`{"apiKey":"key"}`)
	if _, err := Parse(data); err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
}

func assertDefaultConfig(t *testing.T, cfg Config) {
	t.Helper()
	if cfg.DayCount != 30 {
		t.Fatalf("expected clamped dayCount, got %d", cfg.DayCount)
	}
	if cfg.HeadlineScale != 40 {
		t.Fatalf("expected clamped headlineScale, got %d", cfg.HeadlineScale)
	}
	if len(cfg.RoomObjectIDs) != 8 {
		t.Fatalf("expected default room ids, got %d", len(cfg.RoomObjectIDs))
	}
}

func TestParseUsesTodayDateOverride(t *testing.T) {
	data := []byte(`{"apiKey":"key","todayDateOverride":"2026-04-10"}`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if cfg.TodayDateOverride != "2026-04-10" {
		t.Fatalf("unexpected todayDateOverride: %q", cfg.TodayDateOverride)
	}
}

func TestParseAcceptsLegacyReferenceDate(t *testing.T) {
	data := []byte(`{"apiKey":"key","referenceDate":"2026-04-11"}`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if cfg.TodayDateOverride != "2026-04-11" {
		t.Fatalf("unexpected todayDateOverride: %q", cfg.TodayDateOverride)
	}
}
