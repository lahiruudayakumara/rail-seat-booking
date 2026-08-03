//go:build integration

package waitlist

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/journey"
)

func TestCancellationNotifiesOldestEligibleWaitlistEntry(t *testing.T) {
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

	runID, passengerID := uuid.New(), uuid.New()
	originID := uuid.MustParse("20000000-0000-4000-8000-000000000001")
	destinationID := uuid.MustParse("20000000-0000-4000-8000-000000000003")
	if _, err = pool.Exec(ctx, `INSERT INTO train_runs(id,train_id,route_id,service_date,departure_at,arrival_at)
		VALUES($1,'30000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000001',current_date+30,now()+interval '30 days',now()+interval '30 days 11 hours')`, runID); err != nil {
		t.Fatalf("seed data and migrations are required: %v", err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO passengers(id,full_name,email_normalized) VALUES($1,'Waitlist Test',$2)`, passengerID, passengerID.String()+"@example.test"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM outbox_messages WHERE aggregate_id IN (SELECT id FROM waitlist_entries WHERE train_run_id=$1)`, runID)
		_, _ = pool.Exec(ctx, `DELETE FROM train_runs WHERE id=$1`, runID)
		_, _ = pool.Exec(ctx, `DELETE FROM passengers WHERE id=$1`, passengerID)
	}()

	rows, err := pool.Query(ctx, `SELECT s.id FROM seats s JOIN coaches c ON c.id=s.coach_id
		WHERE c.reservation_type='RESERVED' AND c.coach_class='FIRST' AND c.active AND s.active`)
	if err != nil {
		t.Fatal(err)
	}
	var seatIDs []uuid.UUID
	for rows.Next() {
		var seatID uuid.UUID
		if err = rows.Scan(&seatID); err != nil {
			t.Fatal(err)
		}
		seatIDs = append(seatIDs, seatID)
	}
	rows.Close()
	if len(seatIDs) == 0 {
		t.Fatal("seed data has no active first-class reserved seats")
	}

	bookingIDs := make([]uuid.UUID, 0, len(seatIDs))
	for index, seatID := range seatIDs {
		bookingID := uuid.New()
		bookingIDs = append(bookingIDs, bookingID)
		if _, err = pool.Exec(ctx, `INSERT INTO bookings(id,reference,train_run_id,seat_id,passenger_id,
			origin_station_id,destination_station_id,origin_position,destination_position,status,
			fare_total_minor,fare_currency,fare_breakdown,confirmed_at)
			VALUES($1,$2,$3,$4,$5,$6,$7,0,2,'CONFIRMED',25000,'LKR','{}',now())`,
			bookingID, fmt.Sprintf("BK-WAIT%07d", index), runID, seatID, passengerID, originID, destinationID); err != nil {
			t.Fatal(err)
		}
	}

	repository := NewRepository(pool)
	available, err := repository.HasAvailable(ctx, runID, journey.Segment{OriginPosition: 0, DestinationPosition: 2}, "FIRST")
	if err != nil {
		t.Fatal(err)
	}
	if available {
		t.Fatal("expected the first-class segment to be sold out")
	}

	older := Entry{ID: uuid.New(), Reference: "WL-OLDESTTEST1", TrainRunID: runID, OriginStationID: originID,
		DestinationStationID: destinationID, OriginPosition: 0, DestinationPosition: 2, FullName: "Oldest Passenger",
		Email: "oldest@example.test", PreferredCoachClass: "FIRST", Status: "WAITING"}
	older, err = repository.Insert(ctx, older, nil)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)
	newer := older
	newer.ID, newer.Reference, newer.Email = uuid.New(), "WL-NEWESTTEST1", "newest@example.test"
	if _, err = repository.Insert(ctx, newer, nil); err != nil {
		t.Fatal(err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `UPDATE bookings SET status='CANCELLED',cancelled_at=now() WHERE id=$1`, bookingIDs[0]); err != nil {
		t.Fatal(err)
	}
	if err = repository.NotifyNext(ctx, tx, bookingIDs[0], "integration-request"); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	var olderStatus, newerStatus string
	if err = pool.QueryRow(ctx, `SELECT status FROM waitlist_entries WHERE id=$1`, older.ID).Scan(&olderStatus); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT status FROM waitlist_entries WHERE id=$1`, newer.ID).Scan(&newerStatus); err != nil {
		t.Fatal(err)
	}
	if olderStatus != "NOTIFIED" || newerStatus != "WAITING" {
		t.Fatalf("got oldest=%s newest=%s; want NOTIFIED and WAITING", olderStatus, newerStatus)
	}
	var topic string
	if err = pool.QueryRow(ctx, `SELECT topic FROM outbox_messages WHERE aggregate_id=$1`, older.ID).Scan(&topic); err != nil {
		t.Fatal(err)
	}
	if topic != "WAITLIST_SEAT_AVAILABLE" {
		t.Fatalf("got outbox topic %q", topic)
	}
}
