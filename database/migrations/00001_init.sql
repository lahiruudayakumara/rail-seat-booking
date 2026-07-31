-- +goose Up
CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE routes (
  id uuid PRIMARY KEY,
  code varchar(32) NOT NULL UNIQUE,
  name text NOT NULL,
  direction text NOT NULL,
  timezone text NOT NULL DEFAULT 'Asia/Colombo',
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE stations (
  id uuid PRIMARY KEY,
  code varchar(16) NOT NULL UNIQUE,
  name text NOT NULL,
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE route_stations (
  route_id uuid NOT NULL REFERENCES routes(id),
  station_id uuid NOT NULL REFERENCES stations(id),
  position integer NOT NULL CHECK (position >= 0),
  cumulative_distance_m integer NOT NULL CHECK (cumulative_distance_m >= 0),
  PRIMARY KEY (route_id, station_id),
  UNIQUE (route_id, position)
);

CREATE TABLE trains (
  id uuid PRIMARY KEY,
  code varchar(32) NOT NULL UNIQUE,
  name text NOT NULL,
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE train_runs (
  id uuid PRIMARY KEY,
  train_id uuid NOT NULL REFERENCES trains(id),
  route_id uuid NOT NULL REFERENCES routes(id),
  service_date date NOT NULL,
  departure_at timestamptz NOT NULL,
  arrival_at timestamptz NOT NULL,
  status text NOT NULL DEFAULT 'SCHEDULED' CHECK (status IN ('SCHEDULED','BOARDING','DEPARTED','COMPLETED','CANCELLED')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK (departure_at < arrival_at),
  UNIQUE (train_id, route_id, service_date, departure_at)
);

CREATE TABLE coaches (
  id uuid PRIMARY KEY,
  train_id uuid NOT NULL REFERENCES trains(id),
  code varchar(16) NOT NULL,
  sequence smallint NOT NULL CHECK (sequence > 0),
  coach_class text NOT NULL,
  reservation_type text NOT NULL CHECK (reservation_type IN ('RESERVED','UNRESERVED')),
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (train_id, code),
  UNIQUE (train_id, sequence)
);

CREATE TABLE seats (
  id uuid PRIMARY KEY,
  coach_id uuid NOT NULL REFERENCES coaches(id),
  label varchar(16) NOT NULL,
  row_number smallint,
  column_code varchar(8),
  attributes jsonb NOT NULL DEFAULT '[]'::jsonb,
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (coach_id, label)
);

CREATE TABLE passengers (
  id uuid PRIMARY KEY,
  full_name text NOT NULL CHECK (length(full_name) BETWEEN 1 AND 120),
  email_normalized text,
  phone_e164 text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK (email_normalized IS NOT NULL OR phone_e164 IS NOT NULL)
);

CREATE TABLE fare_rules (
  id uuid PRIMARY KEY,
  route_id uuid NOT NULL REFERENCES routes(id),
  coach_class text NOT NULL,
  currency char(3) NOT NULL DEFAULT 'LKR',
  currency_scale smallint NOT NULL DEFAULT 2 CHECK (currency_scale BETWEEN 0 AND 4),
  base_fee_minor bigint NOT NULL CHECK (base_fee_minor >= 0),
  rate_per_km_minor bigint NOT NULL CHECK (rate_per_km_minor >= 0),
  minimum_fare_minor bigint NOT NULL CHECK (minimum_fare_minor >= 0),
  class_multiplier_basis_points integer NOT NULL DEFAULT 10000 CHECK (class_multiplier_basis_points > 0),
  effective_from date NOT NULL,
  effective_to date,
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK (effective_to IS NULL OR effective_to > effective_from)
);

CREATE TABLE fare_quotes (
  id uuid PRIMARY KEY,
  train_run_id uuid NOT NULL REFERENCES train_runs(id),
  seat_id uuid NOT NULL REFERENCES seats(id),
  origin_station_id uuid NOT NULL REFERENCES stations(id),
  destination_station_id uuid NOT NULL REFERENCES stations(id),
  origin_position integer NOT NULL,
  destination_position integer NOT NULL,
  distance_m integer NOT NULL CHECK (distance_m > 0),
  fare_rule_id uuid NOT NULL REFERENCES fare_rules(id),
  amount_minor bigint NOT NULL CHECK (amount_minor >= 0),
  currency char(3) NOT NULL,
  currency_scale smallint NOT NULL,
  breakdown jsonb NOT NULL,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK (origin_position < destination_position)
);

CREATE TABLE bookings (
  id uuid PRIMARY KEY,
  reference varchar(24) NOT NULL UNIQUE,
  train_run_id uuid NOT NULL REFERENCES train_runs(id),
  seat_id uuid NOT NULL REFERENCES seats(id),
  passenger_id uuid NOT NULL REFERENCES passengers(id),
  origin_station_id uuid NOT NULL REFERENCES stations(id),
  destination_station_id uuid NOT NULL REFERENCES stations(id),
  origin_position integer NOT NULL,
  destination_position integer NOT NULL,
  status text NOT NULL CHECK (status IN ('PENDING','HELD','CONFIRMED','CANCELLED','EXPIRED','COMPLETED')),
  fare_rule_id uuid REFERENCES fare_rules(id),
  fare_total_minor bigint NOT NULL CHECK (fare_total_minor >= 0),
  fare_currency char(3) NOT NULL,
  fare_currency_scale smallint NOT NULL DEFAULT 2,
  fare_breakdown jsonb NOT NULL,
  confirmed_at timestamptz,
  cancelled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK (origin_position >= 0 AND origin_position < destination_position)
);

ALTER TABLE bookings ADD CONSTRAINT prevent_overlapping_active_bookings
  EXCLUDE USING gist (
    train_run_id WITH =,
    seat_id WITH =,
    int4range(origin_position, destination_position, '[)') WITH &&
  ) WHERE (status IN ('HELD', 'CONFIRMED'));

CREATE TABLE idempotency_keys (
  scope text NOT NULL,
  key_hash char(64) NOT NULL,
  request_hash char(64) NOT NULL,
  booking_id uuid REFERENCES bookings(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL DEFAULT now() + interval '24 hours',
  PRIMARY KEY (scope, key_hash)
);

CREATE TABLE audit_events (
  id uuid PRIMARY KEY,
  aggregate_type text NOT NULL,
  aggregate_id uuid NOT NULL,
  event_type text NOT NULL,
  request_id uuid,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  occurred_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX route_stations_order_idx ON route_stations(route_id, position);
CREATE INDEX train_runs_search_idx ON train_runs(service_date, route_id, status, departure_at);
CREATE INDEX coaches_inventory_idx ON coaches(train_id, reservation_type, active);
CREATE INDEX seats_coach_active_idx ON seats(coach_id, active);
CREATE INDEX fare_rules_lookup_idx ON fare_rules(route_id, coach_class, effective_from, effective_to) WHERE active;
CREATE INDEX bookings_availability_idx ON bookings(train_run_id, seat_id, status) WHERE status IN ('HELD','CONFIRMED');
CREATE INDEX bookings_passenger_idx ON bookings(passenger_id, created_at DESC);
CREATE INDEX audit_aggregate_idx ON audit_events(aggregate_type, aggregate_id, occurred_at);

-- +goose Down
DROP TABLE IF EXISTS audit_events, idempotency_keys, bookings, fare_quotes, fare_rules, passengers, seats, coaches, train_runs, trains, route_stations, stations, routes CASCADE;
