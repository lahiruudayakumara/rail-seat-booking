package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/database"
)

type payableBooking struct {
	AmountMinor int64
	Currency    string
	Status      string
	ExpiresAt   time.Time
	FullName    string
	Email       string
	Phone       string
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
	err := db.QueryRow(ctx, `SELECT b.fare_total_minor,b.fare_currency,b.status,b.hold_expires_at,p.full_name,COALESCE(p.email_normalized,''),COALESCE(p.phone_e164,'') FROM bookings b JOIN passengers p ON p.id=b.passenger_id WHERE b.id=$1 FOR UPDATE OF b`, bookingID).Scan(&item.AmountMinor, &item.Currency, &item.Status, &item.ExpiresAt, &item.FullName, &item.Email, &item.Phone)
	return item, err
}

func (r *Repository) FindPayHereCheckoutByKey(ctx context.Context, db database.DBTX, keyHash string) (payHereCheckoutRecord, error) {
	var item payHereCheckoutRecord
	err := db.QueryRow(ctx, `SELECT pay.id,pay.booking_id,pay.status,b.status,pay.amount_minor,pay.currency,b.hold_expires_at,p.full_name,COALESCE(p.email_normalized,''),COALESCE(p.phone_e164,'') FROM payments pay JOIN bookings b ON b.id=pay.booking_id JOIN passengers p ON p.id=b.passenger_id WHERE pay.provider='PAYHERE' AND pay.idempotency_key_hash=$1`, keyHash).Scan(&item.PaymentID, &item.BookingID, &item.Status, &item.BookingStatus, &item.AmountMinor, &item.Currency, &item.ExpiresAt, &item.FullName, &item.Email, &item.Phone)
	return item, err
}

func (r *Repository) InsertPayHerePending(ctx context.Context, db database.DBTX, paymentID, bookingID uuid.UUID, amountMinor int64, currency, keyHash string, expiresAt time.Time) error {
	if _, err := db.Exec(ctx, `INSERT INTO payments(id,booking_id,provider,provider_reference,status,amount_minor,currency,idempotency_key_hash) VALUES($1,$2,'PAYHERE',$3,'PENDING',$4,$5,$6)`, paymentID, bookingID, paymentID.String(), amountMinor, currency, keyHash); err != nil {
		return err
	}
	_, err := db.Exec(ctx, `UPDATE bookings SET hold_expires_at=$2,updated_at=now() WHERE id=$1 AND status='HELD'`, bookingID, expiresAt)
	return err
}

func (r *Repository) LockPayHereOrder(ctx context.Context, db database.DBTX, orderID uuid.UUID) (payHereCheckoutRecord, error) {
	var item payHereCheckoutRecord
	err := db.QueryRow(ctx, `SELECT pay.id,pay.booking_id,pay.status,b.status,pay.amount_minor,pay.currency,b.hold_expires_at,p.full_name,COALESCE(p.email_normalized,''),COALESCE(p.phone_e164,'') FROM payments pay JOIN bookings b ON b.id=pay.booking_id JOIN passengers p ON p.id=b.passenger_id WHERE pay.provider='PAYHERE' AND pay.id=$1 FOR UPDATE OF pay,b`, orderID).Scan(&item.PaymentID, &item.BookingID, &item.Status, &item.BookingStatus, &item.AmountMinor, &item.Currency, &item.ExpiresAt, &item.FullName, &item.Email, &item.Phone)
	return item, err
}

func (r *Repository) InsertWebhookEvent(ctx context.Context, db database.DBTX, notification PayHereNotification) (bool, error) {
	payload, _ := json.Marshal(map[string]string{"merchantId": notification.MerchantID, "orderId": notification.OrderID, "paymentId": notification.PaymentID, "amount": notification.Amount, "currency": notification.Currency, "statusCode": notification.StatusCode, "method": notification.Method, "statusMessage": notification.StatusMessage})
	tag, err := db.Exec(ctx, `INSERT INTO payment_webhook_events(id,provider,event_fingerprint,provider_payment_id,order_id,status_code,payload) VALUES($1,'PAYHERE',$2,NULLIF($3,''),$4,$5,$6) ON CONFLICT(provider,event_fingerprint) DO NOTHING`, uuid.New(), notification.EventFingerprint, notification.PaymentID, notification.OrderID, notification.StatusCode, payload)
	return tag.RowsAffected() == 1, err
}

func (r *Repository) MarkWebhookProcessed(ctx context.Context, db database.DBTX, fingerprint, processingError string) error {
	_, err := db.Exec(ctx, `UPDATE payment_webhook_events SET processed_at=now(),processing_error=NULLIF($2,'') WHERE provider='PAYHERE' AND event_fingerprint=$1`, fingerprint, processingError)
	return err
}

func (r *Repository) CompletePayHere(ctx context.Context, db database.DBTX, paymentID, bookingID, ticketID uuid.UUID, providerReference, ticketHash, requestID string) error {
	now := time.Now()
	tag, err := db.Exec(ctx, `UPDATE payments SET status='PAID',provider_reference=$2,paid_at=$3,updated_at=$3 WHERE id=$1 AND status='PENDING'`, paymentID, providerReference, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("pending PayHere payment state changed")
	}
	tag, err = db.Exec(ctx, `UPDATE bookings SET status='CONFIRMED',confirmed_at=$2,hold_expires_at=NULL,updated_at=$2 WHERE id=$1 AND status='HELD'`, bookingID, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("held booking state changed")
	}
	if _, err := db.Exec(ctx, `INSERT INTO tickets(id,booking_id,verification_code_hash,status) VALUES($1,$2,$3,'ACTIVE') ON CONFLICT(booking_id) DO NOTHING`, ticketID, bookingID, ticketHash); err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]any{"bookingId": bookingID, "paymentId": paymentID, "ticketId": ticketID})
	if _, err := db.Exec(ctx, `INSERT INTO outbox_messages(id,topic,aggregate_id,payload) VALUES($1,'BOOKING_CONFIRMED',$2,$3)`, uuid.New(), bookingID, payload); err != nil {
		return err
	}
	var parsedRequestID *uuid.UUID
	if parsed, err := uuid.Parse(requestID); err == nil {
		parsedRequestID = &parsed
	}
	metadata, _ := json.Marshal(map[string]any{"paymentId": paymentID, "ticketId": ticketID, "provider": "PAYHERE"})
	_, err = db.Exec(ctx, `INSERT INTO audit_events(id,aggregate_type,aggregate_id,event_type,request_id,metadata) VALUES($1,'BOOKING',$2,'PAYMENT_CAPTURED',$3,$4)`, uuid.New(), bookingID, parsedRequestID, metadata)
	return err
}

func (r *Repository) UpdatePayHereStatus(ctx context.Context, db database.DBTX, paymentID, bookingID uuid.UUID, paymentStatus, providerReference string, cancelBooking bool) error {
	if _, err := db.Exec(ctx, `UPDATE payments SET status=$2,provider_reference=COALESCE(NULLIF($3,''),provider_reference),updated_at=now() WHERE id=$1`, paymentID, paymentStatus, providerReference); err != nil {
		return err
	}
	if cancelBooking {
		if _, err := db.Exec(ctx, `UPDATE bookings SET status='CANCELLED',cancelled_at=now(),updated_at=now() WHERE id=$1 AND status IN ('HELD','CONFIRMED')`, bookingID); err != nil {
			return err
		}
		_, err := db.Exec(ctx, `UPDATE tickets SET status='CANCELLED' WHERE booking_id=$1 AND status='ACTIVE'`, bookingID)
		return err
	}
	return nil
}

func (r *Repository) LoadPayHereStatus(ctx context.Context, db database.DBTX, paymentID uuid.UUID) (payHerePaymentRecord, error) {
	var item payHerePaymentRecord
	err := db.QueryRow(ctx, `SELECT pay.id,pay.booking_id,pay.status,pay.provider_reference,pay.amount_minor,pay.currency,pay.paid_at,t.id,t.status FROM payments pay LEFT JOIN tickets t ON t.booking_id=pay.booking_id WHERE pay.id=$1 AND pay.provider='PAYHERE'`, paymentID).Scan(&item.PaymentID, &item.BookingID, &item.Status, &item.ProviderReference, &item.AmountMinor, &item.Currency, &item.PaidAt, &item.TicketID, &item.TicketStatus)
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
