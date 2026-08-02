package payment

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/booking"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/database"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
)

type CancellationProcessor struct {
	provider Provider
}

func NewCancellationProcessor(provider Provider) *CancellationProcessor {
	return &CancellationProcessor{provider: provider}
}

func (p *CancellationProcessor) Process(ctx context.Context, db database.DBTX, bookingID uuid.UUID, reason, requestID string) (*booking.RefundSummary, error) {
	var paymentID uuid.UUID
	var providerReference, status, currency string
	var amountMinor int64
	err := db.QueryRow(ctx, `SELECT id,provider_reference,status,amount_minor,currency FROM payments WHERE booking_id=$1 ORDER BY created_at DESC LIMIT 1 FOR UPDATE`, bookingID).Scan(&paymentID, &providerReference, &status, &amountMinor, &currency)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, apperror.Wrap(err)
	}
	if status != "PAID" {
		return nil, nil
	}
	refundID := uuid.New()
	providerResult, err := p.provider.Refund(ctx, refundID, providerReference, amountMinor, currency)
	if err != nil || providerResult.Status != "SUCCEEDED" {
		return nil, apperror.New(502, "REFUND_FAILED", "The payment refund could not be completed.", nil)
	}
	if _, err = db.Exec(ctx, `INSERT INTO refunds(id,payment_id,amount_minor,status,reason,completed_at) VALUES($1,$2,$3,'SUCCEEDED',NULLIF($4,''),now())`, refundID, paymentID, amountMinor, reason); err != nil {
		return nil, apperror.Wrap(err)
	}
	if _, err = db.Exec(ctx, `UPDATE payments SET status='REFUNDED',updated_at=now() WHERE id=$1`, paymentID); err != nil {
		return nil, apperror.Wrap(err)
	}
	if _, err = db.Exec(ctx, `UPDATE tickets SET status='CANCELLED' WHERE booking_id=$1 AND status='ACTIVE'`, bookingID); err != nil {
		return nil, apperror.Wrap(err)
	}
	payload, _ := json.Marshal(map[string]any{"bookingId": bookingID, "paymentId": paymentID, "refundId": refundID, "requestId": requestID})
	if _, err = db.Exec(ctx, `INSERT INTO outbox_messages(id,topic,aggregate_id,payload) VALUES($1,'BOOKING_CANCELLED',$2,$3)`, uuid.New(), bookingID, payload); err != nil {
		return nil, apperror.Wrap(err)
	}
	return &booking.RefundSummary{ID: refundID, Status: "SUCCEEDED", AmountMinor: amountMinor, Currency: currency}, nil
}
