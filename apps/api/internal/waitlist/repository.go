package waitlist

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/database"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/journey"
)

type Repository struct{ db database.DBTX }

func NewRepository(db database.DBTX) *Repository { return &Repository{db: db} }

func (r *Repository) HasAvailable(ctx context.Context, runID uuid.UUID, segment journey.Segment, coachClass string) (bool, error) {
	var available bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1 FROM train_runs tr
		JOIN coaches c ON c.train_id=tr.train_id AND c.active AND c.reservation_type='RESERVED'
		JOIN seats s ON s.coach_id=c.id AND s.active
		WHERE tr.id=$1 AND ($4='ANY' OR c.coach_class=$4)
		AND NOT EXISTS (
			SELECT 1 FROM bookings b WHERE b.train_run_id=tr.id AND b.seat_id=s.id
			AND (b.status='CONFIRMED' OR (b.status='HELD' AND b.hold_expires_at>now()))
			AND int4range(b.origin_position,b.destination_position,'[)') && int4range($2,$3,'[)')
		)
	)`, runID, segment.OriginPosition, segment.DestinationPosition, coachClass).Scan(&available)
	return available, err
}

func (r *Repository) Insert(ctx context.Context, entry Entry, accountID *uuid.UUID) (Entry, error) {
	err := r.db.QueryRow(ctx, `INSERT INTO waitlist_entries(
		id,reference,account_id,train_run_id,origin_station_id,destination_station_id,
		origin_position,destination_position,full_name,email_normalized,phone_e164,preferred_coach_class
	) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,NULLIF($10,''),NULLIF($11,''),$12)
	RETURNING created_at`, entry.ID, entry.Reference, accountID, entry.TrainRunID, entry.OriginStationID,
		entry.DestinationStationID, entry.OriginPosition, entry.DestinationPosition, entry.FullName,
		entry.Email, entry.Phone, entry.PreferredCoachClass).Scan(&entry.CreatedAt)
	return entry, err
}

func (r *Repository) FindByReferenceAndContact(ctx context.Context, reference, contact string) (Entry, error) {
	var entry Entry
	err := r.db.QueryRow(ctx, `SELECT id,reference,train_run_id,origin_station_id,destination_station_id,
		origin_position,destination_position,full_name,COALESCE(email_normalized,''),COALESCE(phone_e164,''),
		preferred_coach_class,status,created_at,notified_at
		FROM waitlist_entries WHERE reference=$1 AND (email_normalized=lower($2) OR phone_e164=$2)`, reference, contact).
		Scan(&entry.ID, &entry.Reference, &entry.TrainRunID, &entry.OriginStationID, &entry.DestinationStationID,
			&entry.OriginPosition, &entry.DestinationPosition, &entry.FullName, &entry.Email, &entry.Phone,
			&entry.PreferredCoachClass, &entry.Status, &entry.CreatedAt, &entry.NotifiedAt)
	return entry, err
}

func (r *Repository) Cancel(ctx context.Context, id uuid.UUID) (Entry, error) {
	var entry Entry
	err := r.db.QueryRow(ctx, `UPDATE waitlist_entries SET status='CANCELLED',cancelled_at=now(),updated_at=now()
		WHERE id=$1 AND status='WAITING'
		RETURNING id,reference,train_run_id,origin_station_id,destination_station_id,full_name,
		COALESCE(email_normalized,''),COALESCE(phone_e164,''),preferred_coach_class,status,created_at`, id).
		Scan(&entry.ID, &entry.Reference, &entry.TrainRunID, &entry.OriginStationID, &entry.DestinationStationID,
			&entry.FullName, &entry.Email, &entry.Phone, &entry.PreferredCoachClass, &entry.Status, &entry.CreatedAt)
	return entry, err
}

func (r *Repository) NotifyNext(ctx context.Context, db database.DBTX, cancelledBookingID uuid.UUID, requestID string) error {
	var entry Entry
	err := db.QueryRow(ctx, `SELECT w.id,w.reference,w.train_run_id,w.origin_station_id,w.destination_station_id,
		w.full_name,COALESCE(w.email_normalized,''),COALESCE(w.phone_e164,''),w.preferred_coach_class
		FROM bookings cancelled
		JOIN waitlist_entries w ON w.train_run_id=cancelled.train_run_id AND w.status='WAITING'
		WHERE cancelled.id=$1
		AND int4range(w.origin_position,w.destination_position,'[)') && int4range(cancelled.origin_position,cancelled.destination_position,'[)')
		AND EXISTS (
			SELECT 1 FROM train_runs tr
			JOIN coaches c ON c.train_id=tr.train_id AND c.active AND c.reservation_type='RESERVED'
			JOIN seats s ON s.coach_id=c.id AND s.active
			WHERE tr.id=w.train_run_id AND (w.preferred_coach_class='ANY' OR c.coach_class=w.preferred_coach_class)
			AND NOT EXISTS (
				SELECT 1 FROM bookings active WHERE active.train_run_id=w.train_run_id AND active.seat_id=s.id
				AND (active.status='CONFIRMED' OR (active.status='HELD' AND active.hold_expires_at>now()))
				AND int4range(active.origin_position,active.destination_position,'[)') && int4range(w.origin_position,w.destination_position,'[)')
			)
		)
		ORDER BY w.created_at,w.id FOR UPDATE OF w SKIP LOCKED LIMIT 1`, cancelledBookingID).
		Scan(&entry.ID, &entry.Reference, &entry.TrainRunID, &entry.OriginStationID, &entry.DestinationStationID,
			&entry.FullName, &entry.Email, &entry.Phone, &entry.PreferredCoachClass)
	if err != nil {
		return err
	}
	if _, err = db.Exec(ctx, `UPDATE waitlist_entries SET status='NOTIFIED',notified_at=now(),updated_at=now() WHERE id=$1`, entry.ID); err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]any{
		"waitlistEntryId": entry.ID, "reference": entry.Reference, "trainRunId": entry.TrainRunID,
		"originStationId": entry.OriginStationID, "destinationStationId": entry.DestinationStationID,
		"email": entry.Email, "phone": entry.Phone, "requestId": requestID,
	})
	_, err = db.Exec(ctx, `INSERT INTO outbox_messages(id,topic,aggregate_id,payload) VALUES($1,'WAITLIST_SEAT_AVAILABLE',$2,$3)`, uuid.New(), entry.ID, payload)
	return err
}
