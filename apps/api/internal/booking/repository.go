package booking

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/database"
)

type Repository struct{}

func NewRepository() *Repository { return &Repository{} }
func (r *Repository) ReserveIdempotency(ctx context.Context, db database.DBTX, keyHash, requestHash string) (bool, error) {
	tag, err := db.Exec(ctx, `INSERT INTO idempotency_keys(scope,key_hash,request_hash) VALUES('create-booking',$1,$2) ON CONFLICT DO NOTHING`, keyHash, requestHash)
	return tag.RowsAffected() == 1, err
}
func (r *Repository) Idempotency(ctx context.Context, db database.DBTX, keyHash string) (string, *uuid.UUID, error) {
	var requestHash string
	var bookingID *uuid.UUID
	err := db.QueryRow(ctx, `SELECT request_hash,booking_id FROM idempotency_keys WHERE scope='create-booking' AND key_hash=$1 FOR UPDATE`, keyHash).Scan(&requestHash, &bookingID)
	return requestHash, bookingID, err
}
func (r *Repository) AttachIdempotency(ctx context.Context, db database.DBTX, keyHash string, bookingID uuid.UUID) error {
	_, err := db.Exec(ctx, `UPDATE idempotency_keys SET booking_id=$1 WHERE scope='create-booking' AND key_hash=$2`, bookingID, keyHash)
	return err
}
func (r *Repository) Quote(ctx context.Context, db database.DBTX, id uuid.UUID) (QuoteSnapshot, error) {
	var q QuoteSnapshot
	err := db.QueryRow(ctx, `SELECT train_run_id,seat_id,origin_station_id,destination_station_id,fare_rule_id,origin_position,destination_position,amount_minor,currency,currency_scale,breakdown FROM fare_quotes WHERE id=$1 AND expires_at>now()`, id).Scan(&q.TrainRunID, &q.SeatID, &q.OriginStationID, &q.DestinationStationID, &q.FareRuleID, &q.OriginPosition, &q.DestinationPosition, &q.AmountMinor, &q.Currency, &q.CurrencyScale, &q.Breakdown)
	return q, err
}
func (r *Repository) InsertPassenger(ctx context.Context, db database.DBTX, id uuid.UUID, input PassengerInput) error {
	_, err := db.Exec(ctx, `INSERT INTO passengers(id,full_name,email_normalized,phone_e164) VALUES($1,$2,NULLIF(lower($3),''),NULLIF($4,''))`, id, input.FullName, input.Email, input.Phone)
	return err
}
func (r *Repository) InsertBooking(ctx context.Context, db database.DBTX, id uuid.UUID, reference string, passengerID uuid.UUID, q QuoteSnapshot) error {
	_, err := db.Exec(ctx, `INSERT INTO bookings(id,reference,train_run_id,seat_id,passenger_id,origin_station_id,destination_station_id,origin_position,destination_position,status,fare_rule_id,fare_total_minor,fare_currency,fare_currency_scale,fare_breakdown,confirmed_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,'CONFIRMED',$10,$11,$12,$13,$14,now())`, id, reference, q.TrainRunID, q.SeatID, passengerID, q.OriginStationID, q.DestinationStationID, q.OriginPosition, q.DestinationPosition, q.FareRuleID, q.AmountMinor, q.Currency, q.CurrencyScale, q.Breakdown)
	return err
}
func (r *Repository) InsertAudit(ctx context.Context, db database.DBTX, eventID, aggregateID uuid.UUID, eventType, requestID string) error {
	var parsed *uuid.UUID
	if id, err := uuid.Parse(requestID); err == nil {
		parsed = &id
	}
	metadata, _ := json.Marshal(map[string]string{})
	_, err := db.Exec(ctx, `INSERT INTO audit_events(id,aggregate_type,aggregate_id,event_type,request_id,metadata) VALUES($1,'BOOKING',$2,$3,$4,$5)`, eventID, aggregateID, eventType, parsed, metadata)
	return err
}
func (r *Repository) Load(ctx context.Context, db database.DBTX, id uuid.UUID) (Booking, error) {
	var b Booking
	var attributes []byte
	err := db.QueryRow(ctx, `SELECT b.id,b.reference,b.status,b.train_run_id,s.id,s.label,c.id,c.code,c.coach_class,s.attributes,b.origin_station_id,b.destination_station_id,b.fare_total_minor,b.fare_currency,b.fare_currency_scale,b.created_at,b.confirmed_at,b.cancelled_at FROM bookings b JOIN seats s ON s.id=b.seat_id JOIN coaches c ON c.id=s.coach_id WHERE b.id=$1`, id).Scan(&b.ID, &b.Reference, &b.Status, &b.TrainRunID, &b.Seat.ID, &b.Seat.Label, &b.Seat.CoachID, &b.Seat.CoachCode, &b.Seat.CoachClass, &attributes, &b.OriginStationID, &b.DestinationStationID, &b.Fare.AmountMinor, &b.Fare.Currency, &b.Fare.CurrencyScale, &b.CreatedAt, &b.ConfirmedAt, &b.CancelledAt)
	if err == nil {
		err = json.Unmarshal(attributes, &b.Seat.Attributes)
	}
	return b, err
}
func (r *Repository) FindIDByReferenceAndContact(ctx context.Context, db database.DBTX, reference, contact string) (uuid.UUID, error) {
	var id uuid.UUID
	err := db.QueryRow(ctx, `SELECT b.id FROM bookings b JOIN passengers p ON p.id=b.passenger_id WHERE b.reference=$1 AND (p.email_normalized=lower($2) OR p.phone_e164=$2)`, reference, contact).Scan(&id)
	return id, err
}
func (r *Repository) LockStatus(ctx context.Context, db database.DBTX, id uuid.UUID) (string, error) {
	var status string
	err := db.QueryRow(ctx, `SELECT status FROM bookings WHERE id=$1 FOR UPDATE`, id).Scan(&status)
	return status, err
}
func (r *Repository) Cancel(ctx context.Context, db database.DBTX, id uuid.UUID) error {
	_, err := db.Exec(ctx, `UPDATE bookings SET status='CANCELLED',cancelled_at=now(),updated_at=now() WHERE id=$1`, id)
	return err
}
func IsExclusionViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23P01"
}
