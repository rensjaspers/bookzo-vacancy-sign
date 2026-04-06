//go:build sdl

package sdlrender

import (
	"github.com/veandco/go-sdl2/sdl"
)

var measureColor = sdl.Color{R: 0, G: 0, B: 0, A: 255}

func (r *Renderer) fittingFontSizeForOnePhrase(
	text string,
	startSize int,
	minSize int,
	maxWidth int32,
	wrap bool,
) (int, error) {
	current := startSize
	for {
		surface, err := r.singleLineSurface(text, current, measureColor)
		if err != nil {
			return 0, err
		}
		w := surface.W
		surface.Free()
		if w <= maxWidth || current <= minSize {
			if w <= maxWidth || !wrap {
				return current, nil
			}
			return current, nil
		}
		current = nextSize(current, minSize, startSize)
	}
}

func (r *Renderer) minUnifyFontSize(
	texts []string,
	startSize int,
	minSize int,
	maxWidth int32,
	wrap bool,
) (int, error) {
	if len(texts) == 0 {
		return startSize, nil
	}
	smallest := startSize
	for _, t := range texts {
		s, err := r.fittingFontSizeForOnePhrase(t, startSize, minSize, maxWidth, wrap)
		if err != nil {
			return 0, err
		}
		if s < smallest {
			smallest = s
		}
	}
	return smallest, nil
}

func (r *Renderer) headlineSurfaceAtFixed(
	text string,
	box layout,
	color sdl.Color,
	size int,
) (*sdl.Surface, error) {
	headlineMin := maxInt(box.headlineSize*3/5, box.sublineSize)
	if size < headlineMin {
		size = headlineMin
	}
	maxW := box.width * 9 / 10
	surface, err := r.singleLineSurface(text, size, color)
	if err != nil {
		return nil, err
	}
	if surface.W <= maxW {
		return surface, nil
	}
	surface.Free()
	return r.centeredWrappedSurface(text, size, maxW, color)
}

func (r *Renderer) hotelSurfaceAtFixed(
	text string,
	box layout,
	color sdl.Color,
	size int,
) (*sdl.Surface, error) {
	hotelMin := box.hotelSize * 3 / 4
	if size < hotelMin {
		size = hotelMin
	}
	return r.singleLineSurface(text, size, color)
}

func (r *Renderer) sublineSurfaceAtFixed(
	text string,
	box layout,
	color sdl.Color,
	size int,
) (*sdl.Surface, error) {
	subMin := box.sublineSize * 4 / 5
	if size < subMin {
		size = subMin
	}
	return r.singleLineSurface(text, size, color)
}
