package runtime

import (
	"context"
	"log/slog"
	"time"

	"hotel-rasch-vacancy-go/internal/render"
	"hotel-rasch-vacancy-go/internal/vacancy"
)

type App struct {
	logger            *slog.Logger
	store             *vacancy.Store
	service           *vacancy.Service
	renderer          render.Runner
	pollInterval      time.Duration
	todayDateOverride string
	dayCount          int
	bookzoRequests    int
}

func NewApp(
	logger *slog.Logger,
	store *vacancy.Store,
	service *vacancy.Service,
	renderer render.Runner,
	pollInterval time.Duration,
	todayDateOverride string,
	dayCount int,
	bookzoRequests int,
) *App {
	return &App{
		logger:            logger,
		store:             store,
		service:           service,
		renderer:          renderer,
		pollInterval:      pollInterval,
		todayDateOverride: todayDateOverride,
		dayCount:          dayCount,
		bookzoRequests:    bookzoRequests,
	}
}

func (a *App) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go a.pollLoop(runCtx)
	return a.renderer.Run(runCtx, a.store)
}

func (a *App) pollLoop(ctx context.Context) {
	a.refresh(ctx)
	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()
	for {
		if done := waitPoll(ctx, ticker); done {
			return
		}
		a.refresh(ctx)
	}
}

func waitPoll(ctx context.Context, ticker *time.Ticker) bool {
	select {
	case <-ctx.Done():
		return true
	case <-ticker.C:
		return false
	}
}

func (a *App) refresh(ctx context.Context) {
	result, err := a.fetch(ctx)
	a.store.UpdateLoad(result)
	a.logRefresh(err, result)
}

func (a *App) fetch(ctx context.Context) (vacancy.LoadResult, error) {
	start := time.Now()
	refDate, _ := vacancy.EffectiveToday(time.Now(), a.todayDateOverride)
	a.logFetchStart(refDate)
	requestCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	availability, err := a.service.CheckInAvailability(requestCtx, refDate, a.dayCount)
	a.logFetchDone(err, start)
	return loadResult(availability, err), err
}

func (a *App) logFetchStart(refDate time.Time) {
	a.logger.Debug(
		"bookzo fetch",
		"referenceDate", vacancy.FormatYMD(refDate),
		"dayCount", a.dayCount,
		"bookzoRequests", a.bookzoRequests,
	)
}

func (a *App) logFetchDone(err error, start time.Time) {
	elapsed := time.Since(start)
	n := a.bookzoRequests
	if err != nil {
		a.logger.Debug("bookzo fetch done", "elapsed", elapsed, "bookzoRequests", n, "err", err)
		return
	}
	a.logger.Debug("bookzo fetch done", "elapsed", elapsed, "bookzoRequests", n)
}

func loadResult(
	availability vacancy.CheckInAvailability,
	err error,
) vacancy.LoadResult {
	if err != nil {
		return failedResult()
	}
	return okResult(availability)
}

func failedResult() vacancy.LoadResult {
	return vacancy.LoadResult{AnyRoomAvailable: true, FetchFailed: true, FetchedAt: time.Now()}
}

func okResult(availability vacancy.CheckInAvailability) vacancy.LoadResult {
	return vacancy.LoadResult{
		AnyRoomAvailable:   availability.AnyRoomAvailable,
		FirstAvailableDate: availability.FirstAvailableDate,
		FetchedAt:          time.Now(),
	}
}

func (a *App) logRefresh(err error, result vacancy.LoadResult) {
	if err != nil {
		a.logger.Warn("bookzo refresh failed", "error", err)
		return
	}
	a.logger.Info("bookzo refresh ok", "available", result.AnyRoomAvailable, "firstAvailableDate", result.FirstAvailableDate)
}
