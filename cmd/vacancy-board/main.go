package main

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/rensjaspers/bookzo-vacancy-sign/internal/bookzo"
	"github.com/rensjaspers/bookzo-vacancy-sign/internal/config"
	sdlrender "github.com/rensjaspers/bookzo-vacancy-sign/internal/render/sdl"
	appruntime "github.com/rensjaspers/bookzo-vacancy-sign/internal/runtime"
	"github.com/rensjaspers/bookzo-vacancy-sign/internal/vacancy"
)

var embeddedConfigBase64 string

func main() {
	os.Exit(run())
}

func run() int {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	defer recoverMain(logger)
	if err := execute(logger); err != nil {
		logger.Error("startup failed", "error", err)
		return 1
	}
	return 0
}

func execute(logger *slog.Logger) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	return runApp(loggerFor(cfg), cfg)
}

func loadConfig() (config.Config, error) {
	path := configPath()
	if path != "" {
		return config.Load(path)
	}
	return autoConfig()
}

func configPath() string {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	path := fs.String("config", "", "")
	_ = fs.Parse(os.Args[1:])
	return *path
}

func loggerFor(cfg config.Config) *slog.Logger {
	options := &slog.HandlerOptions{Level: parseLevel(cfg.LogLevel)}
	return slog.New(slog.NewTextHandler(os.Stdout, options))
}

func parseLevel(value string) slog.Level {
	switch value {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func runApp(logger *slog.Logger, cfg config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	logger.Debug("startup config", debugStartupAttrs(cfg)...)
	store := vacancy.NewStore(settingsFromConfig(cfg))
	service := vacancy.NewService(bookzo.New(cfg.APIKey, cfg.RequestOrigin, logger), cfg.RoomObjectIDs, cfg.RequestConcurrency)
	renderer := sdlrender.New(rendererConfig(cfg))
	app := appruntime.NewApp(
		logger,
		store,
		service,
		renderer,
		cfg.PollInterval,
		cfg.TodayDateOverride,
		cfg.DayCount,
		len(cfg.RoomObjectIDs),
	)
	return app.Run(ctx)
}

func settingsFromConfig(cfg config.Config) vacancy.Settings {
	return vacancy.Settings{
		HotelName:         cfg.HotelName,
		TodayDateOverride: cfg.TodayDateOverride,
		DayCount:          cfg.DayCount,
		OverrideMode:      cfg.OverrideMode,
		ThemeMode:         cfg.ThemeMode,
		LightWindow:       cfg.LightWindow,
		PhraseRotate:      cfg.PhraseRotate,
		HeadlineScale:     cfg.HeadlineScale,
		Languages:         cfg.Languages,
		ShowErrorHint:     cfg.ShowErrorHint,
	}
}

func rendererConfig(cfg config.Config) sdlrender.Config {
	return sdlrender.Config{
		FontPath:   cfg.FontPath,
		Fullscreen: cfg.Fullscreen,
		Title:      cfg.HotelName,
		Width:      cfg.WindowWidth,
		Height:     cfg.WindowHeight,
	}
}

func recoverMain(logger *slog.Logger) {
	if value := recover(); value != nil {
		logger.Error("panic recovered", "value", value)
	}
}

func autoConfig() (config.Config, error) {
	if path, ok := firstExistingConfig(); ok {
		return config.Load(path)
	}
	return embeddedConfig()
}

func firstExistingConfig() (string, bool) {
	for _, path := range configCandidates() {
		if fileExists(path) {
			return path, true
		}
	}
	return "", false
}

func configCandidates() []string {
	dir, err := executableDir()
	if err != nil {
		return []string{"config.json", "config.pi.json"}
	}
	return []string{
		filepath.Join(dir, "config.json"),
		filepath.Join(dir, "config.pi.json"),
		"config.json",
		"config.pi.json",
	}
}

func executableDir() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(path), nil
}

func embeddedConfig() (config.Config, error) {
	data, err := embeddedConfigBytes()
	if err != nil {
		return config.Config{}, err
	}
	dir, _ := executableDir()
	return config.LoadBytes(data, dir)
}

func embeddedConfigBytes() ([]byte, error) {
	if embeddedConfigBase64 == "" {
		return nil, errors.New("no config.json or embedded config found")
	}
	return base64.StdEncoding.DecodeString(embeddedConfigBase64)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
