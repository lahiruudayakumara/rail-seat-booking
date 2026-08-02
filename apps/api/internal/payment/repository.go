package payment

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/database"
)

type payableBooking struct {
	AmountMinor int64
	Currency    string
	Status      string
	ExpiresAt   time.Time
}

type Repository struct{}

func NewRepository() *Repository { return &Repository{} }

func (r *Repository) FindCheckout(ctx context.Context, db database.DBTX, keyHash string) (checkoutRecord, error) {
	var item checkoutRecord
	err := db.QueryRow(ctx, `SELECT p.booking_id,p.id,p.status,p.provider,p.provider_reference,p.amount_minor,p.currency,p.paid_at,t.id
		FROM payments p JOIN tickets t ON t.booking_id=p.booking_id
		WHERE p.idempotency_key_hash=$1`, keyHash).Scan(
		&item.BookingID, &item.Payment.ID, &item.Payment.Status, &item.Payment.Provider,
		&item.Payment.ProviderReference, &item.Payment.AmountMinor, &item.Payment.Currency,
		&item.Payment.PaidAt, &item.TicketID,
	)
	return item, err
}

func (r *Repository) VerifyTicket(ctx context.Context, db database.DBTX, codeHash string) (TicketVerification, error) {
	var item TicketVerification
	err := db.QueryRow(ctx, `SELECT t.id,t.status,b.reference,b.status,b.train_run_id,c.code,s.label,b.origin_station_id,b.destination_station_id
		FROM tickets t
		JOIN bookings b ON b.id=t.booking_id
		JOIN seats s ON s.id=b.seat_id
		JOIN coaches c ON c.id=s.coach_id
		WHERE t.verification_code_hash=$1`, codeHash).Scan(
		&item.TicketID, &item.TicketStatus, &item.BookingReference, &item.BookingStatus,
		&item.TrainRunID, &item.CoachCode, &item.SeatLabel,
		&item.OriginStationID, &item.DestinationStationID,
	)
	return item, err
}

func (r *Repository) LockPayable(ctx context.Context, db database.DBTX, bookingID uuid.UUID) (payableBooking, error) {
	var item payableBooking
	err := db.QueryRow(ctx, `SELECT fare_total_minor,fare_currency,status,hold_expires_at FROM bookings WHERE id=$1 AND passenger_id IS NOT NULL FOR UPDATE`, bookingID).Scan(&item.AmountMinor, &item.Currency, &item.Status, &item.ExpiresAt)
	return item, err
}

func (r *Repository) Complete(ctx context.Context, db database.DBTX, payment Payment, bookingID, ticketID uuid.UUID, keyHash, ticketHash, requestID string) error {
	if _, err := db.Exec(ctx, `INSERT INTO payments(id,booking_id,provider,provider_reference,status,amount_minor,currency,idempotency_key_hash,paid_at) VALUES($1,$2,$3,$4,'PAID',$5,$6,$7,$8)`, payment.ID, bookingID, payment.Provider, payment.ProviderReference, payment.AmountMinor, payment.Currency, keyHash, payment.PaidAt); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, `UPDATE bookings SET status='CONFIRMED',confirmed_at=now(),hold_expires_at=NULL,updated_at=now() WHERE id=$1`, bookingID); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, `INSERT INTO tickets(id,booking_id,verification_code_hash,status) VALUES($1,$2,$3,'ACTIVE')`, ticketID, bookingID, ticketHash); err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]any{"bookingId": bookingID, "paymentId": payment.ID, "ticketId": ticketID})
	if _, err := db.Exec(ctx, `INSERT INTO outbox_messages(id,topic,aggregate_id,payload) VALUES($1,'BOOKING_CONFIRMED',$2,$3)`, uuid.New(), bookingID, payload); err != nil {
		return err
	}
	var parsedRequestID *uuid.UUID
	if parsed, err := uuid.Parse(requestID); err == nil {
		parsedRequestID = &parsed
	}
	metadata, _ := json.Marshal(map[string]any{"paymentId": payment.ID, "ticketId": ticketID})
	_, err := db.Exec(ctx, `INSERT INTO audit_events(id,aggregate_type,aggregate_id,event_type,request_id,metadata) VALUES($1,'BOOKING',$2,'PAYMENT_CAPTURED',$3,$4)`, uuid.New(), bookingID, parsedRequestID, metadata)
	return err
}
