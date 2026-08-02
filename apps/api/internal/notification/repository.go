package notification

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) Claim(ctx context.Context) (*Message, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var message Message
	err = tx.QueryRow(ctx, `SELECT id,topic,payload,attempts+1 FROM outbox_messages
		WHERE (status IN ('PENDING','FAILED') AND available_at<=now())
		   OR (status='PROCESSING' AND available_at<=now()-interval '5 minutes')
		ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1`).Scan(&message.ID, &message.Topic, &message.Payload, &message.Attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `UPDATE outbox_messages SET status='PROCESSING',attempts=$2,last_error=NULL,available_at=now() WHERE id=$1`, message.ID, message.Attempts); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *Repository) Delivered(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE outbox_messages SET status='DELIVERED',delivered_at=now(),last_error=NULL WHERE id=$1`, id)
	return err
}

func (r *Repository) Failed(ctx context.Context, id uuid.UUID, deliveryErr error, retryAt time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE outbox_messages SET status='FAILED',last_error=$2,available_at=$3 WHERE id=$1`, id, deliveryErr.Error(), retryAt)
	return err
}
