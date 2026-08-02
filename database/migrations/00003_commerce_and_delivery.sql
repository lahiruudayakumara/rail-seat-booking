-- +goose Up
ALTER TABLE bookings DROP CONSTRAINT booking_hold_shape;
ALTER TABLE bookings ADD CONSTRAINT booking_hold_shape CHECK (
  (status = 'HELD' AND hold_expires_at IS NOT NULL) OR status <> 'HELD'
);

CREATE TABLE payments (
  id uuid PRIMARY KEY,
  booking_id uuid NOT NULL REFERENCES bookings(id),
  provider text NOT NULL,
  provider_reference text NOT NULL,
  status text NOT NULL CHECK (status IN ('PENDING','PAID','FAILED','REFUNDED','PARTIALLY_REFUNDED','DISPUTED')),
  amount_minor bigint NOT NULL CHECK (amount_minor >= 0),
  currency char(3) NOT NULL,
  idempotency_key_hash char(64) NOT NULL UNIQUE,
  paid_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (provider, provider_reference)
);

CREATE TABLE refunds (
  id uuid PRIMARY KEY,
  payment_id uuid NOT NULL REFERENCES payments(id),
  amount_minor bigint NOT NULL CHECK (amount_minor > 0),
  status text NOT NULL CHECK (status IN ('PENDING','SUCCEEDED','FAILED')),
  reason text,
  created_at timestamptz NOT NULL DEFAULT now(),
  completed_at timestamptz
);

CREATE TABLE tickets (
  id uuid PRIMARY KEY,
  booking_id uuid NOT NULL UNIQUE REFERENCES bookings(id),
  verification_code_hash char(64) NOT NULL UNIQUE,
  status text NOT NULL CHECK (status IN ('ACTIVE','CANCELLED','USED')),
  issued_at timestamptz NOT NULL DEFAULT now(),
  used_at timestamptz
);

CREATE TABLE outbox_messages (
  id uuid PRIMARY KEY,
  topic text NOT NULL,
  aggregate_id uuid NOT NULL,
  payload jsonb NOT NULL,
  status text NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','PROCESSING','DELIVERED','FAILED')),
  attempts integer NOT NULL DEFAULT 0,
  available_at timestamptz NOT NULL DEFAULT now(),
  delivered_at timestamptz,
  last_error text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX payments_booking_idx ON payments(booking_id, created_at DESC);
CREATE INDEX refunds_payment_idx ON refunds(payment_id, created_at DESC);
CREATE INDEX outbox_delivery_idx ON outbox_messages(status, available_at) WHERE status IN ('PENDING','FAILED');

-- +goose Down
DROP TABLE IF EXISTS outbox_messages, tickets, refunds, payments;
UPDATE bookings SET passenger_id=NULL WHERE status='HELD';
ALTER TABLE bookings DROP CONSTRAINT booking_hold_shape;
ALTER TABLE bookings ADD CONSTRAINT booking_hold_shape CHECK (
  (status = 'HELD' AND passenger_id IS NULL AND hold_expires_at IS NOT NULL) OR status <> 'HELD'
);
