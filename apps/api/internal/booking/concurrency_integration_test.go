//go:build integration

package booking

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestConcurrentOverlappingBookingsExactlyOneSucceeds(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	runID := uuid.New()
	_, err = pool.Exec(ctx, `INSERT INTO train_runs(id,train_id,route_id,service_date,departure_at,arrival_at) VALUES($1,'30000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000001',current_date+30,now()+interval '30 days',now()+interval '30 days 11 hours')`, runID)
	if err != nil {
		t.Fatalf("seed data and migrations are required: %v", err)
	}
	var seatID uuid.UUID
	if err = pool.QueryRow(ctx, `SELECT s.id FROM seats s JOIN coaches c ON c.id=s.coach_id WHERE c.reservation_type='RESERVED' LIMIT 1`).Scan(&seatID); err != nil {
		t.Fatal(err)
	}
	defer pool.Exec(ctx, `DELETE FROM train_runs WHERE id=$1`, runID)

	const contenders = 12
	start := make(chan struct{})
	var wg sync.WaitGroup
	var successes atomic.Int32
	passengers := make([]uuid.UUID, contenders)
	for i := range contenders {
		passengers[i] = uuid.New()
		wg.Add(1)
		go func(passengerID uuid.UUID) {
			defer wg.Done()
			<-start
			tx, e := pool.Begin(ctx)
			if e != nil {
				t.Error(e)
				return
			}
			defer tx.Rollback(ctx)
			if _, e = tx.Exec(ctx, `INSERT INTO passengers(id,full_name,email_normalized) VALUES($1,'Concurrency Test',$2)`, passengerID, passengerID.String()+"@example.test"); e != nil {
				t.Error(e)
				return
			}
			_, e = tx.Exec(ctx, `INSERT INTO bookings(id,reference,train_run_id,seat_id,passenger_id,origin_station_id,destination_station_id,origin_position,destination_position,status,fare_total_minor,fare_currency,fare_breakdown,confirmed_at) VALUES($1,$2,$3,$4,$5,'20000000-0000-4000-8000-000000000001','20000000-0000-4000-8000-000000000005',0,4,'CONFIRMED',46000,'LKR','{}',now())`, uuid.New(), "BK-"+passengerID.String()[:12], runID, seatID, passengerID)
			if e != nil {
				var pgErr *pgconn.PgError
				if !errors.As(e, &pgErr) || pgErr.Code != "23P01" {
					t.Errorf("unexpected insert error: %v", e)
				}
				return
			}
			if e = tx.Commit(ctx); e != nil {
				t.Error(e)
				return
			}
			successes.Add(1)
		}(passengers[i])
	}
	close(start)
	wg.Wait()
	if got := successes.Load(); got != 1 {
		t.Fatalf("got %d successful bookings; want exactly 1", got)
	}
	var count int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM bookings WHERE train_run_id=$1`, runID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("database has %d bookings, err=%v", count, err)
	}
	_, _ = pool.Exec(ctx, `DELETE FROM bookings WHERE train_run_id=$1`, runID)
	_, _ = pool.Exec(ctx, `DELETE FROM passengers WHERE id=ANY($1)`, passengers)
}
