package config

import (
	"fmt"
	"strconv"
	"strings"
)

type RGBA struct {
	R, G, B, A uint8
}

type ThemePalette struct {
	Background RGBA
	Text       RGBA
	Hotel      RGBA
	Available  RGBA
	Full       RGBA
	Hint       RGBA
}

type rawThemeColors struct {
	Light rawThemePalettePatch `json:"light"`
	Dark  rawThemePalettePatch `json:"dark"`
}

type rawThemePalettePatch struct {
	Background string `json:"background"`
	Text       string `json:"text"`
	Hotel      string `json:"hotel"`
	Available  string `json:"available"`
	Full       string `json:"full"`
	Hint       string `json:"hint"`
}

func LegacyLightPalette() ThemePalette {
	return ThemePalette{
		Background: RGBA{R: 248, G: 248, B: 244, A: 255},
		Text:       RGBA{R: 87, G: 83, B: 78, A: 255},
		Hotel:      RGBA{R: 107, G: 114, B: 128, A: 255},
		Available:  RGBA{R: 6, G: 95, B: 70, A: 255},
		Full:       RGBA{R: 87, G: 83, B: 78, A: 255},
		Hint:       RGBA{R: 120, G: 53, B: 15, A: 255},
	}
}

func LegacyDarkPalette() ThemePalette {
	return ThemePalette{
		Background: RGBA{R: 18, G: 18, B: 22, A: 255},
		Text:       RGBA{R: 229, G: 231, B: 235, A: 255},
		Hotel:      RGBA{R: 156, G: 163, B: 175, A: 255},
		Available:  RGBA{R: 52, G: 211, B: 153, A: 255},
		Full:       RGBA{R: 120, G: 113, B: 108, A: 255},
		Hint:       RGBA{R: 217, G: 249, B: 157, A: 255},
	}
}

func mergeThemePalettes(raw *rawThemeColors) (ThemePalette, ThemePalette, error) {
	if raw == nil {
		return LegacyLightPalette(), LegacyDarkPalette(), nil
	}
	light, err := mergeOnePalette(LegacyLightPalette(), raw.Light)
	if err != nil {
		return ThemePalette{}, ThemePalette{}, err
	}
	dark, err := mergeOnePalette(LegacyDarkPalette(), raw.Dark)
	if err != nil {
		return ThemePalette{}, ThemePalette{}, err
	}
	return light, dark, nil
}

func mergeOnePalette(base ThemePalette, patch rawThemePalettePatch) (ThemePalette, error) {
	out, err := applyColorPatches(base, patch)
	if err != nil {
		return ThemePalette{}, err
	}
	return syncTextToHeadlines(out, patch), nil
}

func applyColorPatches(base ThemePalette, patch rawThemePalettePatch) (ThemePalette, error) {
	out := base
	fields := []struct {
		value string
		set   func(RGBA)
	}{
		{patch.Background, func(c RGBA) { out.Background = c }},
		{patch.Text, func(c RGBA) { out.Text = c }},
		{patch.Hotel, func(c RGBA) { out.Hotel = c }},
		{patch.Available, func(c RGBA) { out.Available = c }},
		{patch.Full, func(c RGBA) { out.Full = c }},
		{patch.Hint, func(c RGBA) { out.Hint = c }},
	}
	for _, field := range fields {
		if field.value == "" {
			continue
		}
		c, err := parseHexColor(field.value)
		if err != nil {
			return ThemePalette{}, err
		}
		field.set(c)
	}
	return out, nil
}

func syncTextToHeadlines(out ThemePalette, patch rawThemePalettePatch) ThemePalette {
	if patch.Text == "" {
		return out
	}
	if patch.Available == "" {
		out.Available = out.Text
	}
	if patch.Full == "" {
		out.Full = out.Text
	}
	return out
}

func parseHexColor(s string) (RGBA, error) {
	t := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(s), "#"))
	if t == "" {
		return RGBA{}, fmt.Errorf("empty color")
	}
	switch len(t) {
	case 3:
		return parseHexNibble3(t)
	case 6:
		return parseHexPairs(t, 255)
	case 8:
		a, err := parseHexByte(t[6:8])
		if err != nil {
			return RGBA{}, err
		}
		return parseHexPairs(t[:6], a)
	default:
		return RGBA{}, fmt.Errorf("invalid color %q", s)
	}
}

func parseHexNibble3(t string) (RGBA, error) {
	r, err := parseHexNibble(t[0:1])
	if err != nil {
		return RGBA{}, err
	}
	g, err := parseHexNibble(t[1:2])
	if err != nil {
		return RGBA{}, err
	}
	b, err := parseHexNibble(t[2:3])
	if err != nil {
		return RGBA{}, err
	}
	return RGBA{R: r | r<<4, G: g | g<<4, B: b | b<<4, A: 255}, nil
}

func parseHexNibble(s string) (uint8, error) {
	v, err := strconv.ParseUint(s, 16, 4)
	if err != nil {
		return 0, err
	}
	return uint8(v), nil
}

func parseHexPairs(hex6 string, alpha uint8) (RGBA, error) {
	r, err := parseHexByte(hex6[0:2])
	if err != nil {
		return RGBA{}, err
	}
	g, err := parseHexByte(hex6[2:4])
	if err != nil {
		return RGBA{}, err
	}
	b, err := parseHexByte(hex6[4:6])
	if err != nil {
		return RGBA{}, err
	}
	return RGBA{R: r, G: g, B: b, A: alpha}, nil
}

func parseHexByte(pair string) (uint8, error) {
	v, err := strconv.ParseUint(pair, 16, 8)
	if err != nil {
		return 0, err
	}
	return uint8(v), nil
}
