//go:build integration

package booking

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestGroupLookupUsesLeadPassengerContact(t *testing.T) {
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

	runID, groupID := uuid.New(), uuid.New()
	leadPassengerID, secondPassengerID := uuid.New(), uuid.New()
	leadBookingID, secondBookingID := uuid.New(), uuid.New()
	originID := uuid.MustParse("20000000-0000-4000-8000-000000000001")
	destinationID := uuid.MustParse("20000000-0000-4000-8000-000000000003")
	if _, err = pool.Exec(ctx, `INSERT INTO train_runs(id,train_id,route_id,service_date,departure_at,arrival_at)
		VALUES($1,'30000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000001',current_date+30,now()+interval '30 days',now()+interval '30 days 11 hours')`, runID); err != nil {
		t.Fatalf("seed data and migrations are required: %v", err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `UPDATE booking_groups SET lead_booking_id=NULL WHERE id=$1`, groupID)
		_, _ = pool.Exec(ctx, `DELETE FROM bookings WHERE booking_group_id=$1`, groupID)
		_, _ = pool.Exec(ctx, `DELETE FROM booking_groups WHERE id=$1`, groupID)
		_, _ = pool.Exec(ctx, `DELETE FROM train_runs WHERE id=$1`, runID)
		_, _ = pool.Exec(ctx, `DELETE FROM passengers WHERE id IN ($1,$2)`, leadPassengerID, secondPassengerID)
	}()
	if _, err = pool.Exec(ctx, `INSERT INTO passengers(id,full_name,email_normalized,phone_e164) VALUES
		($1,'Lead Passenger','lead.lookup@example.test','+94770000001'),
		($2,'Second Passenger','second.lookup@example.test','+94770000002')`, leadPassengerID, secondPassengerID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO booking_groups(id,reference,status,fare_total_minor,fare_currency,fare_currency_scale,confirmed_at)
		VALUES($1,'GR-LOOKUPTEST12','CONFIRMED',50000,'LKR',2,now())`, groupID); err != nil {
		t.Fatal(err)
	}
	var seatIDs []uuid.UUID
	rows, err := pool.Query(ctx, `SELECT s.id FROM seats s JOIN coaches c ON c.id=s.coach_id WHERE c.reservation_type='RESERVED' ORDER BY c.sequence,s.row_number,s.column_code LIMIT 2`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var seatID uuid.UUID
		if err = rows.Scan(&seatID); err != nil {
			t.Fatal(err)
		}
		seatIDs = append(seatIDs, seatID)
	}
	rows.Close()
	if len(seatIDs) != 2 {
		t.Fatal("seed data must contain at least two reserved seats")
	}
	if _, err = pool.Exec(ctx, `INSERT INTO bookings(id,reference,train_run_id,seat_id,passenger_id,booking_group_id,
		origin_station_id,destination_station_id,origin_position,destination_position,status,fare_total_minor,fare_currency,fare_breakdown,confirmed_at) VALUES
		($1,'BK-GRPLOOKUP1',$2,$3,$4,$5,$6,$7,0,2,'CONFIRMED',25000,'LKR','{}',now()),
		($8,'BK-GRPLOOKUP2',$2,$9,$10,$5,$6,$7,0,2,'CONFIRMED',25000,'LKR','{}',now())`,
		leadBookingID, runID, seatIDs[0], leadPassengerID, groupID, originID, destinationID,
		secondBookingID, seatIDs[1], secondPassengerID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `UPDATE booking_groups SET lead_booking_id=$2 WHERE id=$1`, groupID, leadBookingID); err != nil {
		t.Fatal(err)
	}

	service := NewService(pool, NewRepository(), NewAccessSigner("group-lookup-test-secret", time.Hour), nil, nil, time.Minute)
	group, err := service.AccessGroup(ctx, AccessRequest{Reference: " gr-lookuptest12 ", Contact: "Lead.Lookup@Example.Test"})
	if err != nil {
		t.Fatal(err)
	}
	if group.ID != groupID || len(group.Members) != 2 || service.access.Verify(group.ManagementToken, groupID) != nil {
		t.Fatalf("unexpected accessible group: %#v", group)
	}

	_, err = service.repo.FindGroupIDByReferenceAndContact(ctx, pool, "GR-LOOKUPTEST12", "second.lookup@example.test")
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("non-lead passenger contact unexpectedly accessed the group: %v", err)
	}
}
