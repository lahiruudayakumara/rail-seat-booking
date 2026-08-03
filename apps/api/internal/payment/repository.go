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

type payableGroup struct {
	AmountMinor int64
	Currency    string
	Status      string
	ExpiresAt   time.Time
	LeadBooking uuid.UUID
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

func (r *Repository) FindGroupCheckout(ctx context.Context, db database.DBTX, keyHash string) (groupCheckoutRecord, error) {
	var item groupCheckoutRecord
	err := db.QueryRow(ctx, `SELECT p.booking_group_id,p.id,p.status,p.provider,p.provider_reference,p.amount_minor,p.currency,p.paid_at FROM payments p WHERE p.idempotency_key_hash=$1 AND p.booking_group_id IS NOT NULL`, keyHash).Scan(&item.GroupID, &item.Payment.ID, &item.Payment.Status, &item.Payment.Provider, &item.Payment.ProviderReference, &item.Payment.AmountMinor, &item.Payment.Currency, &item.Payment.PaidAt)
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

func (r *Repository) LockGroupPayable(ctx context.Context, db database.DBTX, groupID uuid.UUID) (payableGroup, []uuid.UUID, error) {
	var item payableGroup
	err := db.QueryRow(ctx, `SELECT g.fare_total_minor,g.fare_currency,g.status,g.hold_expires_at,g.lead_booking_id,p.full_name,COALESCE(p.email_normalized,''),COALESCE(p.phone_e164,'') FROM booking_groups g JOIN bookings b ON b.id=g.lead_booking_id JOIN passengers p ON p.id=b.passenger_id WHERE g.id=$1 FOR UPDATE OF g`, groupID).Scan(&item.AmountMinor, &item.Currency, &item.Status, &item.ExpiresAt, &item.LeadBooking, &item.FullName, &item.Email, &item.Phone)
	if err != nil {
		return payableGroup{}, nil, err
	}
	rows, err := db.Query(ctx, `SELECT id FROM bookings WHERE booking_group_id=$1 ORDER BY created_at,id FOR UPDATE`, groupID)
	if err != nil {
		return payableGroup{}, nil, err
	}
	defer rows.Close()
	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			return payableGroup{}, nil, err
		}
		ids = append(ids, id)
	}
	return item, ids, rows.Err()
}

func (r *Repository) LoadGroupTickets(ctx context.Context, db database.DBTX, groupID uuid.UUID) ([]Ticket, error) {
	rows, err := db.Query(ctx, `SELECT t.id,t.status FROM tickets t JOIN bookings b ON b.id=t.booking_id WHERE b.booking_group_id=$1 ORDER BY b.created_at,b.id`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Ticket, 0)
	for rows.Next() {
		var item Ticket
		if err = rows.Scan(&item.ID, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CompleteGroup(ctx context.Context, db database.DBTX, payment Payment, groupID, leadBookingID uuid.UUID, issues []ticketIssue, keyHash, requestID string) error {
	if _, err := db.Exec(ctx, `INSERT INTO payments(id,booking_id,booking_group_id,provider,provider_reference,status,amount_minor,currency,idempotency_key_hash,paid_at) VALUES($1,$2,$3,$4,$5,'PAID',$6,$7,$8,$9)`, payment.ID, leadBookingID, groupID, payment.Provider, payment.ProviderReference, payment.AmountMinor, payment.Currency, keyHash, payment.PaidAt); err != nil {
		return err
	}
	now := time.Now()
	if _, err := db.Exec(ctx, `UPDATE booking_groups SET status='CONFIRMED',confirmed_at=$2,hold_expires_at=NULL,updated_at=$2 WHERE id=$1 AND status='HELD'`, groupID, now); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, `UPDATE bookings SET status='CONFIRMED',confirmed_at=$2,hold_expires_at=NULL,updated_at=$2 WHERE booking_group_id=$1 AND status='HELD'`, groupID, now); err != nil {
		return err
	}
	for _, issue := range issues {
		if _, err := db.Exec(ctx, `INSERT INTO tickets(id,booking_id,verification_code_hash,status) VALUES($1,$2,$3,'ACTIVE')`, issue.ID, issue.BookingID, issue.CodeHash); err != nil {
			return err
		}
		payload, _ := json.Marshal(map[string]any{"bookingGroupId": groupID, "bookingId": issue.BookingID, "paymentId": payment.ID, "ticketId": issue.ID})
		if _, err := db.Exec(ctx, `INSERT INTO outbox_messages(id,topic,aggregate_id,payload) VALUES($1,'BOOKING_CONFIRMED',$2,$3)`, uuid.New(), issue.BookingID, payload); err != nil {
			return err
		}
	}
	return r.insertGroupPaymentAudit(ctx, db, groupID, payment.ID, requestID)
}

func (r *Repository) insertGroupPaymentAudit(ctx context.Context, db database.DBTX, groupID, paymentID uuid.UUID, requestID string) error {
	var parsedRequestID *uuid.UUID
	if parsed, err := uuid.Parse(requestID); err == nil {
		parsedRequestID = &parsed
	}
	metadata, _ := json.Marshal(map[string]any{"paymentId": paymentID})
	_, err := db.Exec(ctx, `INSERT INTO audit_events(id,aggregate_type,aggregate_id,event_type,request_id,metadata) VALUES($1,'BOOKING_GROUP',$2,'PAYMENT_CAPTURED',$3,$4)`, uuid.New(), groupID, parsedRequestID, metadata)
	return err
}

func (r *Repository) FindPayHereCheckoutByKey(ctx context.Context, db database.DBTX, keyHash string) (payHereCheckoutRecord, error) {
	var item payHereCheckoutRecord
	err := db.QueryRow(ctx, `SELECT pay.id,pay.booking_id,pay.booking_group_id,pay.status,b.status,pay.amount_minor,pay.currency,COALESCE(b.hold_expires_at,pay.updated_at),p.full_name,COALESCE(p.email_normalized,''),COALESCE(p.phone_e164,'') FROM payments pay JOIN bookings b ON b.id=pay.booking_id JOIN passengers p ON p.id=b.passenger_id WHERE pay.provider='PAYHERE' AND pay.idempotency_key_hash=$1`, keyHash).Scan(&item.PaymentID, &item.BookingID, &item.GroupID, &item.Status, &item.BookingStatus, &item.AmountMinor, &item.Currency, &item.ExpiresAt, &item.FullName, &item.Email, &item.Phone)
	return item, err
}

func (r *Repository) FindPendingPayHereByBooking(ctx context.Context, db database.DBTX, bookingID uuid.UUID) (payHereCheckoutRecord, error) {
	var item payHereCheckoutRecord
	err := db.QueryRow(ctx, `SELECT pay.id,pay.booking_id,pay.booking_group_id,pay.status,b.status,pay.amount_minor,pay.currency,COALESCE(b.hold_expires_at,pay.updated_at),p.full_name,COALESCE(p.email_normalized,''),COALESCE(p.phone_e164,'') FROM payments pay JOIN bookings b ON b.id=pay.booking_id JOIN passengers p ON p.id=b.passenger_id WHERE pay.provider='PAYHERE' AND pay.booking_id=$1 AND pay.booking_group_id IS NULL AND pay.status='PENDING'`, bookingID).Scan(&item.PaymentID, &item.BookingID, &item.GroupID, &item.Status, &item.BookingStatus, &item.AmountMinor, &item.Currency, &item.ExpiresAt, &item.FullName, &item.Email, &item.Phone)
	return item, err
}

func (r *Repository) FindPendingPayHereByGroup(ctx context.Context, db database.DBTX, groupID uuid.UUID) (payHereCheckoutRecord, error) {
	var item payHereCheckoutRecord
	err := db.QueryRow(ctx, `SELECT pay.id,pay.booking_id,pay.booking_group_id,pay.status,g.status,pay.amount_minor,pay.currency,COALESCE(g.hold_expires_at,pay.updated_at),p.full_name,COALESCE(p.email_normalized,''),COALESCE(p.phone_e164,'') FROM payments pay JOIN booking_groups g ON g.id=pay.booking_group_id JOIN bookings b ON b.id=g.lead_booking_id JOIN passengers p ON p.id=b.passenger_id WHERE pay.provider='PAYHERE' AND pay.booking_group_id=$1 AND pay.status='PENDING'`, groupID).Scan(&item.PaymentID, &item.BookingID, &item.GroupID, &item.Status, &item.BookingStatus, &item.AmountMinor, &item.Currency, &item.ExpiresAt, &item.FullName, &item.Email, &item.Phone)
	return item, err
}

func (r *Repository) InsertPayHereGroupPending(ctx context.Context, db database.DBTX, paymentID, groupID, leadBookingID uuid.UUID, amountMinor int64, currency, keyHash string, expiresAt time.Time) error {
	if _, err := db.Exec(ctx, `INSERT INTO payments(id,booking_id,booking_group_id,provider,provider_reference,status,amount_minor,currency,idempotency_key_hash) VALUES($1,$2,$3,'PAYHERE',$4,'PENDING',$5,$6,$7)`, paymentID, leadBookingID, groupID, paymentID.String(), amountMinor, currency, keyHash); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, `UPDATE booking_groups SET hold_expires_at=$2,updated_at=now() WHERE id=$1 AND status='HELD'`, groupID, expiresAt); err != nil {
		return err
	}
	_, err := db.Exec(ctx, `UPDATE bookings SET hold_expires_at=$2,updated_at=now() WHERE booking_group_id=$1 AND status='HELD'`, groupID, expiresAt)
	return err
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
	err := db.QueryRow(ctx, `SELECT pay.id,pay.booking_id,pay.booking_group_id,pay.status,COALESCE(g.status,b.status),pay.amount_minor,pay.currency,COALESCE(g.hold_expires_at,b.hold_expires_at,pay.updated_at),p.full_name,COALESCE(p.email_normalized,''),COALESCE(p.phone_e164,'') FROM payments pay JOIN bookings b ON b.id=pay.booking_id JOIN passengers p ON p.id=b.passenger_id LEFT JOIN booking_groups g ON g.id=pay.booking_group_id WHERE pay.provider='PAYHERE' AND pay.id=$1 FOR UPDATE OF pay,b`, orderID).Scan(&item.PaymentID, &item.BookingID, &item.GroupID, &item.Status, &item.BookingStatus, &item.AmountMinor, &item.Currency, &item.ExpiresAt, &item.FullName, &item.Email, &item.Phone)
	return item, err
}

func (r *Repository) CompletePayHereGroup(ctx context.Context, db database.DBTX, paymentID, groupID uuid.UUID, issues []ticketIssue, providerReference, requestID string) error {
	now := time.Now()
	tag, err := db.Exec(ctx, `UPDATE payments SET status='PAID',provider_reference=$2,paid_at=$3,updated_at=$3 WHERE id=$1 AND status='PENDING'`, paymentID, providerReference, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("pending PayHere group payment state changed")
	}
	if _, err = db.Exec(ctx, `UPDATE booking_groups SET status='CONFIRMED',confirmed_at=$2,hold_expires_at=NULL,updated_at=$2 WHERE id=$1 AND status='HELD'`, groupID, now); err != nil {
		return err
	}
	if _, err = db.Exec(ctx, `UPDATE bookings SET status='CONFIRMED',confirmed_at=$2,hold_expires_at=NULL,updated_at=$2 WHERE booking_group_id=$1 AND status='HELD'`, groupID, now); err != nil {
		return err
	}
	for _, issue := range issues {
		if _, err = db.Exec(ctx, `INSERT INTO tickets(id,booking_id,verification_code_hash,status) VALUES($1,$2,$3,'ACTIVE') ON CONFLICT(booking_id) DO NOTHING`, issue.ID, issue.BookingID, issue.CodeHash); err != nil {
			return err
		}
		payload, _ := json.Marshal(map[string]any{"bookingGroupId": groupID, "bookingId": issue.BookingID, "paymentId": paymentID, "ticketId": issue.ID})
		if _, err = db.Exec(ctx, `INSERT INTO outbox_messages(id,topic,aggregate_id,payload) VALUES($1,'BOOKING_CONFIRMED',$2,$3)`, uuid.New(), issue.BookingID, payload); err != nil {
			return err
		}
	}
	return r.insertGroupPaymentAudit(ctx, db, groupID, paymentID, requestID)
}

func (r *Repository) UpdatePayHereGroupStatus(ctx context.Context, db database.DBTX, paymentID, groupID uuid.UUID, paymentStatus, providerReference string, cancelGroup bool) error {
	if _, err := db.Exec(ctx, `UPDATE payments SET status=$2,provider_reference=COALESCE(NULLIF($3,''),provider_reference),updated_at=now() WHERE id=$1`, paymentID, paymentStatus, providerReference); err != nil {
		return err
	}
	if cancelGroup {
		if _, err := db.Exec(ctx, `UPDATE booking_groups SET status='CANCELLED',cancelled_at=now(),updated_at=now() WHERE id=$1 AND status IN ('HELD','CONFIRMED')`, groupID); err != nil {
			return err
		}
		if _, err := db.Exec(ctx, `UPDATE bookings SET status='CANCELLED',cancelled_at=now(),updated_at=now() WHERE booking_group_id=$1 AND status IN ('HELD','CONFIRMED')`, groupID); err != nil {
			return err
		}
		_, err := db.Exec(ctx, `UPDATE tickets SET status='CANCELLED' FROM bookings b WHERE tickets.booking_id=b.id AND b.booking_group_id=$1 AND tickets.status='ACTIVE'`, groupID)
		return err
	}
	return nil
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
	err := db.QueryRow(ctx, `SELECT pay.id,pay.booking_id,pay.booking_group_id,pay.status,pay.provider_reference,pay.amount_minor,pay.currency,pay.paid_at,t.id,t.status FROM payments pay LEFT JOIN tickets t ON t.booking_id=pay.booking_id WHERE pay.id=$1 AND pay.provider='PAYHERE'`, paymentID).Scan(&item.PaymentID, &item.BookingID, &item.GroupID, &item.Status, &item.ProviderReference, &item.AmountMinor, &item.Currency, &item.PaidAt, &item.TicketID, &item.TicketStatus)
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
