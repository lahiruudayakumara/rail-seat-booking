-- +goose Up
CREATE TABLE passenger_accounts (
  id uuid PRIMARY KEY,
  full_name text NOT NULL CHECK (length(full_name) BETWEEN 2 AND 120),
  email_normalized text NOT NULL UNIQUE,
  phone_e164 text UNIQUE,
  password_hash text NOT NULL,
  email_verified_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE passenger_sessions (
  id uuid PRIMARY KEY,
  account_id uuid NOT NULL REFERENCES passenger_accounts(id) ON DELETE CASCADE,
  token_hash char(64) NOT NULL UNIQUE,
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  last_seen_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE passengers
  ADD COLUMN account_id uuid REFERENCES passenger_accounts(id) ON DELETE SET NULL;

CREATE INDEX passenger_sessions_active_idx
  ON passenger_sessions(token_hash, expires_at)
  WHERE revoked_at IS NULL;
CREATE INDEX passengers_account_idx ON passengers(account_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS passengers_account_idx;
ALTER TABLE passengers DROP COLUMN IF EXISTS account_id;
DROP TABLE IF EXISTS passenger_sessions, passenger_accounts;
