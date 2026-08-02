package admin

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) Dashboard(ctx context.Context, trainRunID uuid.UUID) (Dashboard, error) {
	var item Dashboard
	err := r.pool.QueryRow(ctx, `
		WITH run AS (
			SELECT id,route_id FROM train_runs WHERE id=$1
		), capacity AS (
			SELECT count(*)::bigint * GREATEST((SELECT count(*)-1 FROM route_stations rs JOIN run ON run.route_id=rs.route_id),0) AS total
			FROM seats s JOIN coaches c ON c.id=s.coach_id JOIN run ON run.id=$1
			WHERE s.active AND c.active AND c.reservation_type='RESERVED'
			  AND c.train_id=(SELECT train_id FROM train_runs WHERE id=run.id)
		), occupancy AS (
			SELECT COALESCE(sum(destination_position-origin_position),0)::bigint AS occupied
			FROM bookings WHERE train_run_id=$1 AND status='CONFIRMED'
		), booking_counts AS (
			SELECT count(*) FILTER (WHERE status='CONFIRMED')::bigint AS confirmed,
			       count(*) FILTER (WHERE status='CANCELLED')::bigint AS cancelled,
			       count(*) FILTER (WHERE status='HELD')::bigint AS held
			FROM bookings WHERE train_run_id=$1
		), money AS (
			SELECT COALESCE(sum(p.amount_minor) FILTER (WHERE p.status IN ('PAID','REFUNDED','PARTIALLY_REFUNDED')),0)::bigint AS gross,
			       COALESCE(sum(rf.amount_minor) FILTER (WHERE rf.status='SUCCEEDED'),0)::bigint AS refunded,
			       COALESCE(max(p.currency),'LKR') AS currency
			FROM bookings b LEFT JOIN payments p ON p.booking_id=b.id LEFT JOIN refunds rf ON rf.payment_id=p.id
			WHERE b.train_run_id=$1
		), delivery AS (
			SELECT count(*)::bigint AS pending FROM outbox_messages o JOIN bookings b ON b.id=o.aggregate_id
			WHERE b.train_run_id=$1 AND o.status IN ('PENDING','FAILED')
		)
		SELECT $1,capacity.total,occupancy.occupied,
		       CASE WHEN capacity.total=0 THEN 0 ELSE round((occupancy.occupied::numeric/capacity.total::numeric)*100,2) END,
		       booking_counts.confirmed,booking_counts.cancelled,booking_counts.held,
		       money.gross,money.refunded,money.gross-money.refunded,money.currency,delivery.pending
		FROM run,capacity,occupancy,booking_counts,money,delivery`, trainRunID).Scan(
		&item.TrainRunID, &item.SellableSeatSegments, &item.OccupiedSeatSegments,
		&item.SegmentUtilizationPercent, &item.ConfirmedBookings, &item.CancelledBookings,
		&item.HeldBookings, &item.GrossRevenueMinor, &item.RefundedMinor,
		&item.NetRevenueMinor, &item.Currency, &item.PendingDeliveries,
	)
	return item, err
}
