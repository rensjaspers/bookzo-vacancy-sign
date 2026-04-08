//go:build !sdl

package sdlrender

import (
	"context"
	"fmt"

	"github.com/rensjaspers/bookzo-vacancy-sign/internal/config"
	"github.com/rensjaspers/bookzo-vacancy-sign/internal/render"
)

type Config struct {
	FontPath   string
	Fullscreen bool
	Title      string
	Width      int
	Height     int
	LightTheme config.ThemePalette
	DarkTheme  config.ThemePalette
}

type Renderer struct {
	config Config
}

func New(config Config) *Renderer {
	return &Renderer{config: config}
}

func (r *Renderer) Run(ctx context.Context, source render.SnapshotSource) error {
	_ = ctx
	_ = source
	return fmt.Errorf("build with -tags sdl to enable native rendering")
}
