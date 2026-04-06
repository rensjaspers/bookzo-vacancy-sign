package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadResolvesFontPathRelativeToConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	data := []byte(`{"apiKey":"key","fontPath":"fonts/board.ttf"}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	want := filepath.Join(dir, "fonts/board.ttf")
	if cfg.FontPath != want {
		t.Fatalf("unexpected fontPath: %q", cfg.FontPath)
	}
}

func TestLoadBytesResolvesFontPathRelativeToBaseDir(t *testing.T) {
	dir := t.TempDir()
	cfg, err := LoadBytes([]byte(`{"apiKey":"key","fontPath":"fonts/board.ttf"}`), dir)
	if err != nil {
		t.Fatalf("LoadBytes returned error: %v", err)
	}
	want := filepath.Join(dir, "fonts/board.ttf")
	if cfg.FontPath != want {
		t.Fatalf("unexpected fontPath: %q", cfg.FontPath)
	}
}

func TestLoadBytesFallsBackToBundledFont(t *testing.T) {
	dir := t.TempDir()
	fontDir := filepath.Join(dir, "fonts")
	if err := os.MkdirAll(fontDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	fontPath := filepath.Join(fontDir, "board.ttf")
	if err := os.WriteFile(fontPath, []byte("fake-font"), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	cfg, err := LoadBytes([]byte(`{"apiKey":"key"}`), dir)
	if err != nil {
		t.Fatalf("LoadBytes returned error: %v", err)
	}
	if cfg.FontPath != fontPath {
		t.Fatalf("unexpected fontPath: %q", cfg.FontPath)
	}
}
