-- +goose Up
ALTER TABLE bookings ALTER COLUMN passenger_id DROP NOT NULL;
ALTER TABLE bookings ADD COLUMN hold_expires_at timestamptz;
ALTER TABLE bookings ADD CONSTRAINT booking_hold_shape CHECK (
  (status = 'HELD' AND passenger_id IS NULL AND hold_expires_at IS NOT NULL)
  OR (status <> 'HELD')
);
CREATE INDEX bookings_expiring_holds_idx ON bookings(hold_expires_at) WHERE status = 'HELD';

-- +goose Down
DROP INDEX IF EXISTS bookings_expiring_holds_idx;
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS booking_hold_shape;
ALTER TABLE bookings DROP COLUMN IF EXISTS hold_expires_at;
DELETE FROM bookings WHERE passenger_id IS NULL;
ALTER TABLE bookings ALTER COLUMN passenger_id SET NOT NULL;
