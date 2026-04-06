package render

import (
	"context"
	"time"

	"hotel-rasch-vacancy-go/internal/vacancy"
)

type SnapshotSource interface {
	Snapshot(now time.Time) vacancy.ViewModel
}

type Runner interface {
	Run(ctx context.Context, source SnapshotSource) error
}
