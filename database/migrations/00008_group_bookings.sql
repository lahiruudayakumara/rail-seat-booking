-- +goose Up
CREATE TABLE booking_groups (
  id uuid PRIMARY KEY,
  reference varchar(24) NOT NULL UNIQUE,
  account_id uuid REFERENCES passenger_accounts(id) ON DELETE SET NULL,
  lead_booking_id uuid,
  status text NOT NULL CHECK (status IN ('HELD','CONFIRMED','CANCELLED','EXPIRED','COMPLETED')),
  fare_total_minor bigint NOT NULL CHECK (fare_total_minor >= 0),
  fare_currency char(3) NOT NULL,
  fare_currency_scale smallint NOT NULL DEFAULT 2,
  hold_expires_at timestamptz,
  confirmed_at timestamptz,
  cancelled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK ((status = 'HELD' AND hold_expires_at IS NOT NULL) OR status <> 'HELD')
);

ALTER TABLE bookings
  ADD COLUMN booking_group_id uuid REFERENCES booking_groups(id) ON DELETE SET NULL;
ALTER TABLE booking_groups
  ADD CONSTRAINT booking_groups_lead_booking_fk
  FOREIGN KEY (lead_booking_id) REFERENCES bookings(id) ON DELETE SET NULL;

ALTER TABLE payments
  ADD COLUMN booking_group_id uuid REFERENCES booking_groups(id) ON DELETE SET NULL;

CREATE TABLE group_idempotency_keys (
  scope text NOT NULL,
  key_hash char(64) NOT NULL,
  request_hash char(64) NOT NULL,
  booking_group_id uuid REFERENCES booking_groups(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL DEFAULT now() + interval '24 hours',
  PRIMARY KEY (scope, key_hash)
);

CREATE INDEX bookings_group_idx ON bookings(booking_group_id, created_at);
CREATE INDEX booking_groups_account_idx ON booking_groups(account_id, created_at DESC);
CREATE UNIQUE INDEX payments_pending_group_idx
  ON payments(booking_group_id)
  WHERE status = 'PENDING' AND booking_group_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS payments_pending_group_idx;
DROP TABLE IF EXISTS group_idempotency_keys;
ALTER TABLE payments DROP COLUMN IF EXISTS booking_group_id;
ALTER TABLE booking_groups DROP CONSTRAINT IF EXISTS booking_groups_lead_booking_fk;
ALTER TABLE bookings DROP COLUMN IF EXISTS booking_group_id;
DROP TABLE IF EXISTS booking_groups;
