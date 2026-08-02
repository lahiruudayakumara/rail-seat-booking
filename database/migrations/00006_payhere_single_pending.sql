-- +goose Up
DROP INDEX IF EXISTS payments_pending_booking_idx;
WITH duplicate_pending AS (
  SELECT id,
         row_number() OVER (PARTITION BY booking_id ORDER BY created_at DESC, id DESC) AS attempt_number
  FROM payments
  WHERE status = 'PENDING'
)
UPDATE payments
SET status = 'FAILED', updated_at = now()
WHERE id IN (SELECT id FROM duplicate_pending WHERE attempt_number > 1);
CREATE UNIQUE INDEX payments_pending_booking_idx
  ON payments(booking_id)
  WHERE status = 'PENDING';

-- +goose Down
DROP INDEX IF EXISTS payments_pending_booking_idx;
CREATE INDEX payments_pending_booking_idx
  ON payments(booking_id, created_at DESC)
  WHERE status = 'PENDING';
