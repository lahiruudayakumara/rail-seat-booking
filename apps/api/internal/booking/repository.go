package booking

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

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
func (r *Repository) InsertPassenger(ctx context.Context, db database.DBTX, id uuid.UUID, accountID *uuid.UUID, input PassengerInput) error {
	if accountID != nil {
		_, err := db.Exec(ctx, `INSERT INTO passengers(id,account_id,full_name,email_normalized,phone_e164) VALUES($1,$2,$3,NULLIF(lower($4),''),NULLIF($5,''))`, id, *accountID, input.FullName, input.Email, input.Phone)
		return err
	}
	_, err := db.Exec(ctx, `INSERT INTO passengers(id,full_name,email_normalized,phone_e164) VALUES($1,$2,NULLIF(lower($3),''),NULLIF($4,''))`, id, input.FullName, input.Email, input.Phone)
	return err
}

func (r *Repository) PrepareHeldBookingGroup(ctx context.Context, db database.DBTX, id, passengerID, groupID uuid.UUID, reference string) error {
	_, err := db.Exec(ctx, `UPDATE bookings SET reference=$2,passenger_id=$3,booking_group_id=$4,updated_at=now() WHERE id=$1`, id, reference, passengerID, groupID)
	return err
}

func (r *Repository) ReserveGroupIdempotency(ctx context.Context, db database.DBTX, keyHash, requestHash string) (bool, error) {
	tag, err := db.Exec(ctx, `INSERT INTO group_idempotency_keys(scope,key_hash,request_hash) VALUES('create-group',$1,$2) ON CONFLICT DO NOTHING`, keyHash, requestHash)
	return tag.RowsAffected() == 1, err
}

func (r *Repository) GroupIdempotency(ctx context.Context, db database.DBTX, keyHash string) (string, *uuid.UUID, error) {
	var requestHash string
	var groupID *uuid.UUID
	err := db.QueryRow(ctx, `SELECT request_hash,booking_group_id FROM group_idempotency_keys WHERE scope='create-group' AND key_hash=$1 FOR UPDATE`, keyHash).Scan(&requestHash, &groupID)
	return requestHash, groupID, err
}

func (r *Repository) AttachGroupIdempotency(ctx context.Context, db database.DBTX, keyHash string, groupID uuid.UUID) error {
	_, err := db.Exec(ctx, `UPDATE group_idempotency_keys SET booking_group_id=$1 WHERE scope='create-group' AND key_hash=$2`, groupID, keyHash)
	return err
}

func (r *Repository) InsertGroup(ctx context.Context, db database.DBTX, id uuid.UUID, reference string, accountID *uuid.UUID, amountMinor int64, currency string, currencyScale int16, expiresAt time.Time) error {
	_, err := db.Exec(ctx, `INSERT INTO booking_groups(id,reference,account_id,status,fare_total_minor,fare_currency,fare_currency_scale,hold_expires_at) VALUES($1,$2,$3,'HELD',$4,$5,$6,$7)`, id, reference, accountID, amountMinor, currency, currencyScale, expiresAt)
	return err
}

func (r *Repository) SetGroupLead(ctx context.Context, db database.DBTX, groupID, bookingID uuid.UUID) error {
	_, err := db.Exec(ctx, `UPDATE booking_groups SET lead_booking_id=$2 WHERE id=$1`, groupID, bookingID)
	return err
}

func (r *Repository) InsertGroupAudit(ctx context.Context, db database.DBTX, eventID, groupID uuid.UUID, eventType, requestID string) error {
	var parsed *uuid.UUID
	if id, err := uuid.Parse(requestID); err == nil {
		parsed = &id
	}
	_, err := db.Exec(ctx, `INSERT INTO audit_events(id,aggregate_type,aggregate_id,event_type,request_id,metadata) VALUES($1,'BOOKING_GROUP',$2,$3,$4,'{}'::jsonb)`, eventID, groupID, eventType, parsed)
	return err
}

func (r *Repository) LoadGroupHeader(ctx context.Context, db database.DBTX, groupID uuid.UUID) (BookingGroup, error) {
	var group BookingGroup
	err := db.QueryRow(ctx, `SELECT id,reference,status,fare_total_minor,fare_currency,fare_currency_scale,created_at,confirmed_at FROM booking_groups WHERE id=$1`, groupID).Scan(&group.ID, &group.Reference, &group.Status, &group.Fare.AmountMinor, &group.Fare.Currency, &group.Fare.CurrencyScale, &group.CreatedAt, &group.ConfirmedAt)
	return group, err
}

func (r *Repository) GroupBookingIDs(ctx context.Context, db database.DBTX, groupID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := db.Query(ctx, `SELECT id FROM bookings WHERE booking_group_id=$1 ORDER BY created_at,id`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *Repository) BookingPassenger(ctx context.Context, db database.DBTX, bookingID uuid.UUID) (PassengerInput, error) {
	var passenger PassengerInput
	err := db.QueryRow(ctx, `SELECT p.full_name,COALESCE(p.email_normalized,''),COALESCE(p.phone_e164,'') FROM bookings b JOIN passengers p ON p.id=b.passenger_id WHERE b.id=$1`, bookingID).Scan(&passenger.FullName, &passenger.Email, &passenger.Phone)
	return passenger, err
}
func (r *Repository) InsertBooking(ctx context.Context, db database.DBTX, id uuid.UUID, reference string, passengerID uuid.UUID, q QuoteSnapshot) error {
	_, err := db.Exec(ctx, `INSERT INTO bookings(id,reference,train_run_id,seat_id,passenger_id,origin_station_id,destination_station_id,origin_position,destination_position,status,fare_rule_id,fare_total_minor,fare_currency,fare_currency_scale,fare_breakdown,confirmed_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,'CONFIRMED',$10,$11,$12,$13,$14,now())`, id, reference, q.TrainRunID, q.SeatID, passengerID, q.OriginStationID, q.DestinationStationID, q.OriginPosition, q.DestinationPosition, q.FareRuleID, q.AmountMinor, q.Currency, q.CurrencyScale, q.Breakdown)
	return err
}
func (r *Repository) ExpireHolds(ctx context.Context, db database.DBTX) error {
	_, err := db.Exec(ctx, `UPDATE bookings SET status='EXPIRED',updated_at=now() WHERE status='HELD' AND hold_expires_at<=now()`)
	return err
}
func (r *Repository) InsertHold(ctx context.Context, db database.DBTX, id uuid.UUID, q QuoteSnapshot, expiresAt time.Time) error {
	_, err := db.Exec(ctx, `INSERT INTO bookings(id,reference,train_run_id,seat_id,origin_station_id,destination_station_id,origin_position,destination_position,status,fare_rule_id,fare_total_minor,fare_currency,fare_currency_scale,fare_breakdown,hold_expires_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,'HELD',$9,$10,$11,$12,$13,$14)`, id, "HLD-"+strings.ToUpper(id.String()[:12]), q.TrainRunID, q.SeatID, q.OriginStationID, q.DestinationStationID, q.OriginPosition, q.DestinationPosition, q.FareRuleID, q.AmountMinor, q.Currency, q.CurrencyScale, q.Breakdown, expiresAt)
	return err
}
func (r *Repository) ReleaseUnusedHold(ctx context.Context, db database.DBTX, id uuid.UUID) (bool, error) {
	tag, err := db.Exec(ctx, `UPDATE bookings SET status='EXPIRED',hold_expires_at=NULL,updated_at=now() WHERE id=$1 AND status='HELD' AND passenger_id IS NULL AND booking_group_id IS NULL`, id)
	return tag.RowsAffected() == 1, err
}
func (r *Repository) LockHold(ctx context.Context, db database.DBTX, id uuid.UUID) (QuoteSnapshot, string, time.Time, error) {
	var q QuoteSnapshot
	var status string
	var expiresAt time.Time
	err := db.QueryRow(ctx, `SELECT train_run_id,seat_id,origin_station_id,destination_station_id,fare_rule_id,origin_position,destination_position,fare_total_minor,fare_currency,fare_currency_scale,fare_breakdown,status,hold_expires_at FROM bookings WHERE id=$1 FOR UPDATE`, id).Scan(&q.TrainRunID, &q.SeatID, &q.OriginStationID, &q.DestinationStationID, &q.FareRuleID, &q.OriginPosition, &q.DestinationPosition, &q.AmountMinor, &q.Currency, &q.CurrencyScale, &q.Breakdown, &status, &expiresAt)
	return q, status, expiresAt, err
}
func (r *Repository) PrepareHeldBooking(ctx context.Context, db database.DBTX, id, passengerID uuid.UUID, reference string) error {
	_, err := db.Exec(ctx, `UPDATE bookings SET reference=$2,passenger_id=$3,updated_at=now() WHERE id=$1`, id, reference, passengerID)
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
func (r *Repository) ListByAccount(ctx context.Context, db database.DBTX, accountID uuid.UUID) ([]Booking, error) {
	rows, err := db.Query(ctx, `SELECT b.id,b.reference,b.status,b.train_run_id,s.id,s.label,c.id,c.code,c.coach_class,s.attributes,b.origin_station_id,b.destination_station_id,b.fare_total_minor,b.fare_currency,b.fare_currency_scale,b.created_at,b.confirmed_at,b.cancelled_at FROM bookings b JOIN passengers p ON p.id=b.passenger_id JOIN seats s ON s.id=b.seat_id JOIN coaches c ON c.id=s.coach_id WHERE p.account_id=$1 ORDER BY b.created_at DESC`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Booking, 0)
	for rows.Next() {
		var item Booking
		var attributes []byte
		if err = rows.Scan(&item.ID, &item.Reference, &item.Status, &item.TrainRunID, &item.Seat.ID, &item.Seat.Label, &item.Seat.CoachID, &item.Seat.CoachCode, &item.Seat.CoachClass, &attributes, &item.OriginStationID, &item.DestinationStationID, &item.Fare.AmountMinor, &item.Fare.Currency, &item.Fare.CurrencyScale, &item.CreatedAt, &item.ConfirmedAt, &item.CancelledAt); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(attributes, &item.Seat.Attributes); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (r *Repository) BelongsToAccount(ctx context.Context, db database.DBTX, bookingID, accountID uuid.UUID) (bool, error) {
	var belongs bool
	err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM bookings b JOIN passengers p ON p.id=b.passenger_id WHERE b.id=$1 AND p.account_id=$2)`, bookingID, accountID).Scan(&belongs)
	return belongs, err
}
func (r *Repository) FindIDByReferenceAndContact(ctx context.Context, db database.DBTX, reference, contact string) (uuid.UUID, error) {
	var id uuid.UUID
	err := db.QueryRow(ctx, `SELECT b.id FROM bookings b JOIN passengers p ON p.id=b.passenger_id WHERE b.reference=$1 AND (p.email_normalized=lower($2) OR p.phone_e164=$2)`, reference, contact).Scan(&id)
	return id, err
}
func (r *Repository) LockStatus(ctx context.Context, db database.DBTX, id uuid.UUID) (string, time.Time, error) {
	var status string
	var departureAt time.Time
	err := db.QueryRow(ctx, `SELECT b.status,tr.departure_at FROM bookings b JOIN train_runs tr ON tr.id=b.train_run_id WHERE b.id=$1 FOR UPDATE OF b`, id).Scan(&status, &departureAt)
	return status, departureAt, err
}
func (r *Repository) Cancel(ctx context.Context, db database.DBTX, id uuid.UUID) error {
	_, err := db.Exec(ctx, `UPDATE bookings SET status='CANCELLED',cancelled_at=now(),updated_at=now() WHERE id=$1`, id)
	return err
}
func IsExclusionViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23P01"
}
