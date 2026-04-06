package render

import (
	"context"
	"time"

	"github.com/rensjaspers/bookzo-vacancy-sign/internal/vacancy"
)

type SnapshotSource interface {
	Snapshot(now time.Time) vacancy.ViewModel
}

type Runner interface {
	Run(ctx context.Context, source SnapshotSource) error
}
