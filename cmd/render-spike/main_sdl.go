//go:build sdl

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	sdlrender "hotel-rasch-vacancy-go/internal/render/sdl"
	"hotel-rasch-vacancy-go/internal/vacancy"
)

type spikeSource struct{}

func main() {
	os.Exit(run())
}

func run() int {
	fontPath := flag.String("font", "", "")
	fullscreen := flag.Bool("fullscreen", false, "")
	flag.Parse()
	if *fontPath == "" {
		fmt.Fprintln(os.Stderr, "-font is required")
		return 1
	}
	if err := execute(*fontPath, *fullscreen); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func execute(fontPath string, fullscreen bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	renderer := sdlrender.New(config(fontPath, fullscreen))
	return renderer.Run(ctx, spikeSource{})
}

func config(fontPath string, fullscreen bool) sdlrender.Config {
	return sdlrender.Config{FontPath: fontPath, Fullscreen: fullscreen, Title: "Render Spike", Width: 1280, Height: 720}
}

func (spikeSource) Snapshot(now time.Time) vacancy.ViewModel {
	return vacancy.ViewModel{
		HotelName:     "Hotel Rasch",
		Headline:      "ROOMS AVAILABLE",
		Subline:       "from 5 April",
		Dark:          now.Second()%2 == 0,
		Available:     true,
		ShowSubline:   true,
		ShowErrorHint: now.Second()%10 < 5,
		HeadlineScale: 80,
	}
}
