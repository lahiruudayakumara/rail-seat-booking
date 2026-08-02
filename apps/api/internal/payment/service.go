package payment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/booking"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
)

type Service struct {
	pool     *pgxpool.Pool
	repo     *Repository
	bookings *booking.Repository
	access   *booking.AccessSigner
	tickets  *TicketSigner
	provider Provider
}

func NewService(pool *pgxpool.Pool, repo *Repository, bookings *booking.Repository, access *booking.AccessSigner, tickets *TicketSigner, provider Provider) *Service {
	return &Service{pool: pool, repo: repo, bookings: bookings, access: access, tickets: tickets, provider: provider}
}

func (s *Service) Checkout(ctx context.Context, request CheckoutRequest, idempotencyKey, requestID string) (CheckoutResult, error) {
	if len(idempotencyKey) < 16 || len(idempotencyKey) > 128 {
		return CheckoutResult{}, apperror.Validation("Idempotency-Key", "Header must contain 16 to 128 characters.")
	}
	if err := s.access.Verify(request.BookingToken, request.BookingID); err != nil {
		return CheckoutResult{}, apperror.New(401, "BOOKING_ACCESS_DENIED", "Booking access verification is required.", nil)
	}
	keyDigest := sha256.Sum256([]byte(idempotencyKey))
	keyHash := hex.EncodeToString(keyDigest[:])
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CheckoutResult{}, apperror.Wrap(err)
	}
	defer tx.Rollback(ctx)
	if existing, findErr := s.repo.FindCheckout(ctx, tx, keyHash); findErr == nil {
		if existing.BookingID != request.BookingID {
			return CheckoutResult{}, apperror.New(409, "IDEMPOTENCY_KEY_REUSED", "The idempotency key belongs to a different checkout.", nil)
		}
		return s.checkoutResult(ctx, tx, existing)
	} else if !errors.Is(findErr, pgx.ErrNoRows) {
		return CheckoutResult{}, apperror.Wrap(findErr)
	}
	payable, err := s.repo.LockPayable(ctx, tx, request.BookingID)
	if errors.Is(err, pgx.ErrNoRows) {
		return CheckoutResult{}, apperror.New(404, "BOOKING_NOT_FOUND", "Booking was not found.", nil)
	}
	if err != nil {
		return CheckoutResult{}, apperror.Wrap(err)
	}
	// A concurrent request may have completed while this transaction waited for
	// the booking lock. Re-read the idempotent result before evaluating status.
	if existing, findErr := s.repo.FindCheckout(ctx, tx, keyHash); findErr == nil {
		if existing.BookingID != request.BookingID {
			return CheckoutResult{}, apperror.New(409, "IDEMPOTENCY_KEY_REUSED", "The idempotency key belongs to a different checkout.", nil)
		}
		return s.checkoutResult(ctx, tx, existing)
	} else if !errors.Is(findErr, pgx.ErrNoRows) {
		return CheckoutResult{}, apperror.Wrap(findErr)
	}
	if payable.Status != "HELD" || !time.Now().Before(payable.ExpiresAt) {
		return CheckoutResult{}, apperror.New(409, "HOLD_EXPIRED", "The seat hold has expired.", nil)
	}
	paymentID := uuid.New()
	providerResult, err := s.provider.Charge(ctx, paymentID, payable.AmountMinor, payable.Currency)
	if err != nil || providerResult.Status != "PAID" {
		return CheckoutResult{}, apperror.New(502, "PAYMENT_FAILED", "Payment could not be completed.", nil)
	}
	ticketID := uuid.New()
	verificationCode := s.tickets.Code(ticketID)
	codeHash := sha256.Sum256([]byte(verificationCode))
	payment := Payment{ID: paymentID, Status: "PAID", Provider: "SANDBOX", ProviderReference: providerResult.Reference, AmountMinor: payable.AmountMinor, Currency: payable.Currency, PaidAt: time.Now()}
	ticket := Ticket{ID: ticketID, VerificationCode: verificationCode, Status: "ACTIVE"}
	if err = s.repo.Complete(ctx, tx, payment, request.BookingID, ticket.ID, keyHash, hex.EncodeToString(codeHash[:]), requestID); err != nil {
		return CheckoutResult{}, apperror.Wrap(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return CheckoutResult{}, apperror.Wrap(err)
	}
	confirmed, err := s.bookings.Load(ctx, s.pool, request.BookingID)
	if err != nil {
		return CheckoutResult{}, apperror.Wrap(err)
	}
	confirmed.ManagementToken = s.access.Sign(confirmed.ID)
	return CheckoutResult{Payment: payment, Ticket: ticket, Booking: confirmed}, nil
}

func (s *Service) checkoutResult(ctx context.Context, tx pgx.Tx, existing checkoutRecord) (CheckoutResult, error) {
	confirmed, err := s.bookings.Load(ctx, tx, existing.BookingID)
	if err != nil {
		return CheckoutResult{}, apperror.Wrap(err)
	}
	confirmed.ManagementToken = s.access.Sign(confirmed.ID)
	if err = tx.Commit(ctx); err != nil {
		return CheckoutResult{}, apperror.Wrap(err)
	}
	ticket := Ticket{ID: existing.TicketID, VerificationCode: s.tickets.Code(existing.TicketID), Status: "ACTIVE"}
	return CheckoutResult{Payment: existing.Payment, Ticket: ticket, Booking: confirmed}, nil
}
