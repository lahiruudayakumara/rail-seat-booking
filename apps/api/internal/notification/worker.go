package notification

import (
	"context"
	"log/slog"
	"time"
)

type Worker struct {
	repo     *Repository
	provider Provider
	logger   *slog.Logger
	interval time.Duration
}

func NewWorker(repo *Repository, provider Provider, logger *slog.Logger, interval time.Duration) *Worker {
	return &Worker{repo: repo, provider: provider, logger: logger, interval: interval}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		w.drain(ctx, 20)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (w *Worker) drain(ctx context.Context, limit int) {
	for range limit {
		message, err := w.repo.Claim(ctx)
		if err != nil {
			if ctx.Err() == nil {
				w.logger.Error("claim notification", "error", err)
			}
			return
		}
		if message == nil {
			return
		}
		if err = w.provider.Deliver(ctx, *message); err != nil {
			retryAt := time.Now().Add(retryDelay(message.Attempts))
			if markErr := w.repo.Failed(ctx, message.ID, err, retryAt); markErr != nil && ctx.Err() == nil {
				w.logger.Error("mark notification failed", "message_id", message.ID, "error", markErr)
			}
			continue
		}
		if err = w.repo.Delivered(ctx, message.ID); err != nil && ctx.Err() == nil {
			w.logger.Error("mark notification delivered", "message_id", message.ID, "error", err)
		}
	}
}

func retryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 8 {
		attempt = 8
	}
	return time.Duration(1<<(attempt-1)) * 5 * time.Second
}
