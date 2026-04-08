package config

import (
	"testing"
)

func TestParseHexColorFormats(t *testing.T) {
	cases := []struct {
		in   string
		want RGBA
	}{
		{"#fff", RGBA{R: 255, G: 255, B: 255, A: 255}},
		{"#000000", RGBA{R: 0, G: 0, B: 0, A: 255}},
		{"#00000080", RGBA{R: 0, G: 0, B: 0, A: 128}},
	}
	for _, tc := range cases {
		got, err := parseHexColor(tc.in)
		if err != nil {
			t.Fatalf("parseHexColor(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("parseHexColor(%q) = %+v, want %+v", tc.in, got, tc.want)
		}
	}
}

func TestMergeThemeTextCopiesToHeadlines(t *testing.T) {
	raw := &rawThemeColors{
		Light: rawThemePalettePatch{
			Background: "#ffffff",
			Text:       "#111111",
		},
	}
	light, _, err := mergeThemePalettes(raw)
	if err != nil {
		t.Fatalf("mergeThemePalettes: %v", err)
	}
	if light.Available != light.Text || light.Full != light.Text {
		t.Fatalf("expected available and full to follow text: %+v", light)
	}
}

func TestMergeThemeExplicitAvailableOverridesText(t *testing.T) {
	raw := &rawThemeColors{
		Light: rawThemePalettePatch{
			Text:      "#111111",
			Available: "#00aa00",
		},
	}
	light, _, err := mergeThemePalettes(raw)
	if err != nil {
		t.Fatalf("mergeThemePalettes: %v", err)
	}
	if light.Available.R != 0 || light.Available.G != 170 {
		t.Fatalf("unexpected available: %+v", light.Available)
	}
	if light.Full.R != 17 {
		t.Fatalf("full should follow text: %+v", light.Full)
	}
}

func TestParseRejectsInvalidThemeColor(t *testing.T) {
	_, err := Parse([]byte(`{"apiKey":"key","requestOrigin":"https://cfg.test","themeColors":{"light":{"background":"nope"}}}`))
	if err == nil {
		t.Fatal("expected error for invalid hex")
	}
}
