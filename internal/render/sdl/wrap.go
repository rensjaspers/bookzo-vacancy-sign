//go:build sdl

package sdlrender

import (
	"strings"

	"github.com/veandco/go-sdl2/ttf"
)

func wrapLines(font *ttf.Font, text string, maxWidth int) ([]string, error) {
	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return nil, nil
	}
	var lines []string
	cur := ""
	for _, word := range words {
		next, flushed, err := lineAfterWord(font, cur, word, maxWidth)
		if err != nil {
			return nil, err
		}
		lines = append(lines, flushed...)
		cur = next
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines, nil
}

func lineAfterWord(font *ttf.Font, cur, word string, maxWidth int) (string, []string, error) {
	w, _, err := font.SizeUTF8(word)
	if err != nil {
		return "", nil, err
	}
	if w > maxWidth {
		return flushOversizedWord(font, cur, word, maxWidth)
	}
	trial := word
	if cur != "" {
		trial = cur + " " + word
	}
	tw, _, err := font.SizeUTF8(trial)
	if err != nil {
		return "", nil, err
	}
	if tw <= maxWidth {
		return trial, nil, nil
	}
	if cur == "" {
		return word, nil, nil
	}
	return word, []string{cur}, nil
}

func flushOversizedWord(font *ttf.Font, cur, word string, maxWidth int) (string, []string, error) {
	var flushed []string
	if cur != "" {
		flushed = append(flushed, cur)
		cur = ""
	}
	parts, err := breakLongWord(font, word, maxWidth)
	if err != nil {
		return "", nil, err
	}
	for i := 0; i < len(parts)-1; i++ {
		flushed = append(flushed, parts[i])
	}
	return parts[len(parts)-1], flushed, nil
}

func breakLongWord(font *ttf.Font, word string, maxWidth int) ([]string, error) {
	r := []rune(word)
	if len(r) == 0 {
		return nil, nil
	}
	var out []string
	start := 0
	for start < len(r) {
		end, err := longestFit(font, r, start, maxWidth)
		if err != nil {
			return nil, err
		}
		out = append(out, string(r[start:end]))
		start = end
	}
	return out, nil
}

func longestFit(font *ttf.Font, r []rune, start, maxWidth int) (int, error) {
	lo := start + 1
	hi := len(r) + 1
	best := start
	for lo < hi {
		mid := (lo + hi) / 2
		w, _, err := font.SizeUTF8(string(r[start:mid]))
		if err != nil {
			return 0, err
		}
		if w <= maxWidth {
			best = mid
			lo = mid + 1
			continue
		}
		hi = mid
	}
	if best == start {
		return start + 1, nil
	}
	return best, nil
}
