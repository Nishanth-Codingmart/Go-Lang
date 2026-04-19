package cleanup

import (
	"context"
	"log/slog"
	"time"

	"github.com/user/kvstore/internal/store"
)

type Worker struct {
	store    store.Store
	interval time.Duration
	logger   *slog.Logger
}

func NewWorker(s store.Store, interval time.Duration, l *slog.Logger) *Worker {
	return &Worker{
		store:    s,
		interval: interval,
		logger:   l,
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Info("Background cleanup worker started", "interval", w.interval)

	for {
		select {
		case <-ticker.C:
			count := w.store.Cleanup()
			if count > 0 {
				w.logger.Info("Cleanup cycle completed", "expired_keys_removed", count)
			}
		case <-ctx.Done():
			w.logger.Info("Background cleanup worker stopping")
			return
		}
	}
}
