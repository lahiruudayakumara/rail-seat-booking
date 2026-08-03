-- +goose Up
CREATE TABLE saved_travellers (
  id uuid PRIMARY KEY,
  account_id uuid NOT NULL REFERENCES passenger_accounts(id) ON DELETE CASCADE,
  full_name text NOT NULL CHECK (length(full_name) BETWEEN 2 AND 120),
  email_normalized text,
  phone_e164 text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK (email_normalized IS NOT NULL OR phone_e164 IS NOT NULL)
);

CREATE INDEX saved_travellers_account_idx
  ON saved_travellers(account_id, created_at, id);

CREATE TABLE passenger_preferences (
  account_id uuid PRIMARY KEY REFERENCES passenger_accounts(id) ON DELETE CASCADE,
  preferred_coach_class text NOT NULL DEFAULT 'ANY'
    CHECK (preferred_coach_class IN ('ANY', 'FIRST', 'SECOND')),
  preferred_seat_type text NOT NULL DEFAULT 'ANY'
    CHECK (preferred_seat_type IN ('ANY', 'WINDOW', 'AISLE')),
  language text NOT NULL DEFAULT 'en'
    CHECK (language IN ('en', 'si', 'ta')),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS passenger_preferences;
DROP TABLE IF EXISTS saved_travellers;
