//go:build integration

package availability

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/journey"
)

func TestSeatMapUsesSegmentOverlap(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	runID, passengerID, bookingID := uuid.New(), uuid.New(), uuid.New()
	var seatID uuid.UUID
	if err = pool.QueryRow(ctx, `SELECT s.id FROM seats s JOIN coaches c ON c.id=s.coach_id WHERE c.reservation_type='RESERVED' ORDER BY c.sequence,s.row_number,s.column_code LIMIT 1`).Scan(&seatID); err != nil {
		t.Fatalf("seed data and migrations are required: %v", err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO train_runs(id,train_id,route_id,service_date,departure_at,arrival_at) VALUES($1,'30000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000001',current_date+30,now()+interval '30 days',now()+interval '30 days 11 hours')`, runID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO passengers(id,full_name,email_normalized) VALUES($1,'Seat Map Test',$2)`, passengerID, passengerID.String()+"@example.test"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM train_runs WHERE id=$1`, runID)
		_, _ = pool.Exec(ctx, `DELETE FROM passengers WHERE id=$1`, passengerID)
	}()
	if _, err = pool.Exec(ctx, `INSERT INTO bookings(id,reference,train_run_id,seat_id,passenger_id,origin_station_id,destination_station_id,origin_position,destination_position,status,fare_total_minor,fare_currency,fare_breakdown,confirmed_at) VALUES($1,$2,$3,$4,$5,'20000000-0000-4000-8000-000000000001','20000000-0000-4000-8000-000000000003',0,2,'CONFIRMED',25000,'LKR','{}',now())`, bookingID, "BK-"+bookingID.String()[:12], runID, seatID, passengerID); err != nil {
		t.Fatal(err)
	}

	repository := NewRepository(pool)
	overlapping, err := repository.SeatMap(ctx, runID, journey.Segment{OriginPosition: 1, DestinationPosition: 2}, "")
	if err != nil {
		t.Fatal(err)
	}
	adjacent, err := repository.SeatMap(ctx, runID, journey.Segment{OriginPosition: 2, DestinationPosition: 3}, "")
	if err != nil {
		t.Fatal(err)
	}
	if statusForSeat(overlapping, seatID) != SeatBooked {
		t.Fatal("expected overlapping segment to be booked")
	}
	if statusForSeat(adjacent, seatID) != SeatAvailable {
		t.Fatal("expected adjacent segment to be available")
	}
}

func statusForSeat(items []SeatMapItem, seatID uuid.UUID) SeatStatus {
	for _, item := range items {
		if item.ID == seatID {
			return item.AvailabilityStatus
		}
	}
	return ""
}
