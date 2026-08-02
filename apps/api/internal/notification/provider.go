package notification

import (
	"context"
	"log/slog"
)

type Provider interface {
	Deliver(context.Context, Message) error
}

type LogProvider struct{ logger *slog.Logger }

func NewLogProvider(logger *slog.Logger) *LogProvider { return &LogProvider{logger: logger} }

func (p *LogProvider) Deliver(_ context.Context, message Message) error {
	p.logger.Info("sandbox notification delivered", "message_id", message.ID, "topic", message.Topic, "attempt", message.Attempts)
	return nil
}
