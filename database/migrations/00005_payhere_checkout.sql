-- +goose Up
CREATE TABLE payment_webhook_events (
  id uuid PRIMARY KEY,
  provider text NOT NULL,
  event_fingerprint char(64) NOT NULL,
  provider_payment_id text,
  order_id text NOT NULL,
  status_code text NOT NULL,
  payload jsonb NOT NULL,
  received_at timestamptz NOT NULL DEFAULT now(),
  processed_at timestamptz,
  processing_error text,
  UNIQUE (provider, event_fingerprint)
);

CREATE INDEX payment_webhook_order_idx
  ON payment_webhook_events(provider, order_id, received_at DESC);
CREATE INDEX payments_pending_booking_idx
  ON payments(booking_id, created_at DESC)
  WHERE status = 'PENDING';

-- +goose Down
DROP INDEX IF EXISTS payments_pending_booking_idx;
DROP TABLE IF EXISTS payment_webhook_events;
