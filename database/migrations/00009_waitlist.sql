-- +goose Up
CREATE TABLE waitlist_entries (
  id uuid PRIMARY KEY,
  reference varchar(24) NOT NULL UNIQUE,
  account_id uuid REFERENCES passenger_accounts(id) ON DELETE SET NULL,
  train_run_id uuid NOT NULL REFERENCES train_runs(id) ON DELETE CASCADE,
  origin_station_id uuid NOT NULL REFERENCES stations(id),
  destination_station_id uuid NOT NULL REFERENCES stations(id),
  origin_position integer NOT NULL,
  destination_position integer NOT NULL,
  full_name text NOT NULL CHECK (length(full_name) BETWEEN 2 AND 120),
  email_normalized text,
  phone_e164 text,
  preferred_coach_class text NOT NULL DEFAULT 'ANY'
    CHECK (preferred_coach_class IN ('ANY','FIRST','SECOND')),
  status text NOT NULL DEFAULT 'WAITING'
    CHECK (status IN ('WAITING','NOTIFIED','CANCELLED','EXPIRED','FULFILLED')),
  notified_at timestamptz,
  cancelled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK (origin_position >= 0 AND origin_position < destination_position),
  CHECK (email_normalized IS NOT NULL OR phone_e164 IS NOT NULL)
);

CREATE INDEX waitlist_match_idx
  ON waitlist_entries(train_run_id, status, created_at)
  WHERE status = 'WAITING';
CREATE UNIQUE INDEX waitlist_active_email_idx
  ON waitlist_entries(train_run_id, origin_station_id, destination_station_id, email_normalized)
  WHERE status = 'WAITING' AND email_normalized IS NOT NULL;
CREATE UNIQUE INDEX waitlist_active_phone_idx
  ON waitlist_entries(train_run_id, origin_station_id, destination_station_id, phone_e164)
  WHERE status = 'WAITING' AND phone_e164 IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS waitlist_entries;
