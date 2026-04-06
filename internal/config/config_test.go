package config

import (
	"strings"
	"testing"
)

func TestParseRequiresRequestOrigin(t *testing.T) {
	_, err := Parse([]byte(`{"apiKey":"key"}`))
	if err == nil {
		t.Fatal("expected error for missing requestOrigin")
	}
	if !strings.Contains(err.Error(), "requestOrigin") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseRejectsPlaceholderRequestOrigin(t *testing.T) {
	_, err := Parse([]byte(`{"apiKey":"key","requestOrigin":"https://example.com"}`))
	if err == nil {
		t.Fatal("expected error for placeholder requestOrigin")
	}
}

func TestParseUsesDefaultsAndClamps(t *testing.T) {
	data := []byte(`{"apiKey":"key","requestOrigin":"https://cfg.test","dayCount":99,"headlineScale":10}`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	assertDefaultConfig(t, cfg)
}

func TestParseAllowsMissingFontPath(t *testing.T) {
	data := []byte(`{"apiKey":"key","requestOrigin":"https://cfg.test"}`)
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
	data := []byte(`{"apiKey":"key","requestOrigin":"https://cfg.test","todayDateOverride":"2026-04-10"}`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if cfg.TodayDateOverride != "2026-04-10" {
		t.Fatalf("unexpected todayDateOverride: %q", cfg.TodayDateOverride)
	}
}

func TestParseAcceptsLegacyReferenceDate(t *testing.T) {
	data := []byte(`{"apiKey":"key","requestOrigin":"https://cfg.test","referenceDate":"2026-04-11"}`)
	cfg, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if cfg.TodayDateOverride != "2026-04-11" {
		t.Fatalf("unexpected todayDateOverride: %q", cfg.TodayDateOverride)
	}
}
