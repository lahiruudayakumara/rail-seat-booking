package payment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/booking"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
)

type Service struct {
	pool              *pgxpool.Pool
	repo              *Repository
	bookings          *booking.Repository
	access            *booking.AccessSigner
	tickets           *TicketSigner
	provider          Provider
	payHere           *PayHereProvider
	paymentPendingTTL time.Duration
}

func (s *Service) VerifyTicket(ctx context.Context, request VerifyTicketRequest) (TicketVerification, error) {
	code := strings.TrimSpace(request.VerificationCode)
	if len(code) < 32 || len(code) > 128 {
		return TicketVerification{}, apperror.Validation("verificationCode", "A valid ticket verification code is required.")
	}
	digest := sha256.Sum256([]byte(code))
	item, err := s.repo.VerifyTicket(ctx, s.pool, hex.EncodeToString(digest[:]))
	if errors.Is(err, pgx.ErrNoRows) {
		return TicketVerification{}, apperror.New(404, "TICKET_NOT_FOUND", "Ticket was not found.", nil)
	}
	if err != nil {
		return TicketVerification{}, apperror.Wrap(err)
	}
	item.Valid = item.TicketStatus == "ACTIVE" && item.BookingStatus == "CONFIRMED"
	item.VerifiedAt = time.Now()
	return item, nil
}

func NewService(pool *pgxpool.Pool, repo *Repository, bookings *booking.Repository, access *booking.AccessSigner, tickets *TicketSigner, provider Provider, payHere *PayHereProvider, paymentPendingTTL time.Duration) *Service {
	return &Service{pool: pool, repo: repo, bookings: bookings, access: access, tickets: tickets, provider: provider, payHere: payHere, paymentPendingTTL: paymentPendingTTL}
}

func (s *Service) StartPayHereCheckout(ctx context.Context, request PayHereCheckoutRequest, idempotencyKey string) (PayHereCheckoutResponse, error) {
	if s.payHere == nil || !s.payHere.Enabled() {
		return PayHereCheckoutResponse{}, apperror.New(503, "PAYHERE_NOT_CONFIGURED", "PayHere checkout is not configured.", nil)
	}
	request.BillingAddress = strings.TrimSpace(request.BillingAddress)
	request.City = strings.TrimSpace(request.City)
	if len(request.BillingAddress) < 3 || len(request.BillingAddress) > 200 || len(request.City) < 2 || len(request.City) > 80 {
		return PayHereCheckoutResponse{}, apperror.Validation("billingAddress", "Billing address and city are required for PayHere checkout.")
	}
	if len(idempotencyKey) < 16 || len(idempotencyKey) > 128 {
		return PayHereCheckoutResponse{}, apperror.Validation("Idempotency-Key", "Header must contain 16 to 128 characters.")
	}
	if err := s.access.Verify(request.BookingToken, request.BookingID); err != nil {
		return PayHereCheckoutResponse{}, apperror.New(401, "BOOKING_ACCESS_DENIED", "Booking access verification is required.", nil)
	}
	keyDigest := sha256.Sum256([]byte(idempotencyKey))
	keyHash := hex.EncodeToString(keyDigest[:])
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return PayHereCheckoutResponse{}, apperror.Wrap(err)
	}
	defer tx.Rollback(ctx)
	record, err := s.repo.FindPayHereCheckoutByKey(ctx, tx, keyHash)
	if err == nil {
		if record.BookingID != request.BookingID {
			return PayHereCheckoutResponse{}, apperror.New(409, "IDEMPOTENCY_KEY_REUSED", "The idempotency key belongs to a different checkout.", nil)
		}
		if err = tx.Commit(ctx); err != nil {
			return PayHereCheckoutResponse{}, apperror.Wrap(err)
		}
		return s.payHereCheckoutResponse(record, request)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return PayHereCheckoutResponse{}, apperror.Wrap(err)
	}
	payable, err := s.repo.LockPayable(ctx, tx, request.BookingID)
	if errors.Is(err, pgx.ErrNoRows) {
		return PayHereCheckoutResponse{}, apperror.New(404, "BOOKING_NOT_FOUND", "Booking was not found.", nil)
	}
	if err != nil {
		return PayHereCheckoutResponse{}, apperror.Wrap(err)
	}
	if payable.Status != "HELD" || !time.Now().Before(payable.ExpiresAt) {
		return PayHereCheckoutResponse{}, apperror.New(409, "HOLD_EXPIRED", "The seat hold has expired.", nil)
	}
	if payable.Email == "" || payable.Phone == "" {
		return PayHereCheckoutResponse{}, apperror.Validation("passenger", "PayHere requires both an email address and phone number.")
	}
	if existing, findErr := s.repo.FindPendingPayHereByBooking(ctx, tx, request.BookingID); findErr == nil {
		if err = tx.Commit(ctx); err != nil {
			return PayHereCheckoutResponse{}, apperror.Wrap(err)
		}
		return s.payHereCheckoutResponse(existing, request)
	} else if !errors.Is(findErr, pgx.ErrNoRows) {
		return PayHereCheckoutResponse{}, apperror.Wrap(findErr)
	}
	paymentID := uuid.New()
	expiresAt := time.Now().Add(s.paymentPendingTTL)
	record = payHereCheckoutRecord{PaymentID: paymentID, BookingID: request.BookingID, Status: "PENDING", BookingStatus: "HELD", AmountMinor: payable.AmountMinor, Currency: payable.Currency, ExpiresAt: expiresAt, FullName: payable.FullName, Email: payable.Email, Phone: payable.Phone}
	if err = s.repo.InsertPayHerePending(ctx, tx, paymentID, request.BookingID, payable.AmountMinor, payable.Currency, keyHash, expiresAt); err != nil {
		return PayHereCheckoutResponse{}, apperror.Wrap(err)
	}
	response, err := s.payHereCheckoutResponse(record, request)
	if err != nil {
		return PayHereCheckoutResponse{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return PayHereCheckoutResponse{}, apperror.Wrap(err)
	}
	return response, nil
}

func (s *Service) payHereCheckoutResponse(record payHereCheckoutRecord, request PayHereCheckoutRequest) (PayHereCheckoutResponse, error) {
	session, err := s.payHere.CreateSession(record.PaymentID, record.BookingID, PayHereCustomer{FullName: record.FullName, Email: record.Email, Phone: record.Phone, Address: request.BillingAddress, City: request.City}, record.AmountMinor, record.Currency)
	if err != nil {
		return PayHereCheckoutResponse{}, apperror.New(422, "PAYHERE_CHECKOUT_INVALID", err.Error(), nil)
	}
	return PayHereCheckoutResponse{PaymentID: record.PaymentID, BookingID: record.BookingID, Status: record.Status, ExpiresAt: record.ExpiresAt, ActionURL: session.ActionURL, Fields: session.Fields}, nil
}

func (s *Service) HandlePayHereWebhook(ctx context.Context, values map[string][]string, requestID string) error {
	if s.payHere == nil || !s.payHere.Enabled() {
		return apperror.New(503, "PAYHERE_NOT_CONFIGURED", "PayHere checkout is not configured.", nil)
	}
	notification, err := s.payHere.VerifyNotification(values)
	if err != nil {
		return apperror.New(401, "INVALID_PAYHERE_SIGNATURE", "The PayHere notification signature is invalid.", nil)
	}
	orderID, err := uuid.Parse(notification.OrderID)
	if err != nil {
		return apperror.Validation("order_id", "PayHere order ID is invalid.")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return apperror.Wrap(err)
	}
	defer tx.Rollback(ctx)
	record, err := s.repo.LockPayHereOrder(ctx, tx, orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.New(404, "PAYMENT_NOT_FOUND", "Payment was not found.", nil)
	}
	if err != nil {
		return apperror.Wrap(err)
	}
	inserted, err := s.repo.InsertWebhookEvent(ctx, tx, notification)
	if err != nil {
		return apperror.Wrap(err)
	}
	if !inserted {
		if err = tx.Commit(ctx); err != nil {
			return apperror.Wrap(err)
		}
		return nil
	}
	if notification.Amount != formatPayHereAmount(record.AmountMinor) || notification.Currency != record.Currency {
		_ = s.repo.MarkWebhookProcessed(ctx, tx, notification.EventFingerprint, "amount or currency mismatch")
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return apperror.Wrap(commitErr)
		}
		return apperror.New(422, "PAYMENT_AMOUNT_MISMATCH", "Payment amount or currency does not match the booking.", nil)
	}
	switch notification.StatusCode {
	case "2":
		if record.Status == "PAID" {
			break
		}
		if record.Status != "PENDING" || record.BookingStatus != "HELD" || !time.Now().Before(record.ExpiresAt) {
			if record.GroupID != nil {
				err = s.repo.UpdatePayHereGroupStatus(ctx, tx, record.PaymentID, *record.GroupID, "DISPUTED", notification.PaymentID, false)
			} else {
				err = s.repo.UpdatePayHereStatus(ctx, tx, record.PaymentID, record.BookingID, "DISPUTED", notification.PaymentID, false)
			}
			if err != nil {
				return apperror.Wrap(err)
			}
			_ = s.repo.MarkWebhookProcessed(ctx, tx, notification.EventFingerprint, "payment received after reservation expiry")
			if commitErr := tx.Commit(ctx); commitErr != nil {
				return apperror.Wrap(commitErr)
			}
			return apperror.New(409, "PAYMENT_REQUIRES_REVIEW", "Payment arrived after the reservation expired and requires review.", nil)
		}
		if record.GroupID != nil {
			_, bookingIDs, groupErr := s.repo.LockGroupPayable(ctx, tx, *record.GroupID)
			if groupErr != nil {
				return apperror.Wrap(groupErr)
			}
			issues := make([]ticketIssue, 0, len(bookingIDs))
			for _, bookingID := range bookingIDs {
				ticketID := uuid.New()
				codeHash := sha256.Sum256([]byte(s.tickets.Code(ticketID)))
				issues = append(issues, ticketIssue{ID: ticketID, BookingID: bookingID, CodeHash: hex.EncodeToString(codeHash[:])})
			}
			if err = s.repo.CompletePayHereGroup(ctx, tx, record.PaymentID, *record.GroupID, issues, notification.PaymentID, requestID); err != nil {
				return apperror.Wrap(err)
			}
		} else {
			ticketID := uuid.New()
			verificationCode := s.tickets.Code(ticketID)
			codeHash := sha256.Sum256([]byte(verificationCode))
			if err = s.repo.CompletePayHere(ctx, tx, record.PaymentID, record.BookingID, ticketID, notification.PaymentID, hex.EncodeToString(codeHash[:]), requestID); err != nil {
				return apperror.Wrap(err)
			}
		}
	case "0":
		// PayHere is still processing the payment; preserve PENDING.
	case "-1", "-2":
		if record.GroupID != nil {
			err = s.repo.UpdatePayHereGroupStatus(ctx, tx, record.PaymentID, *record.GroupID, "FAILED", notification.PaymentID, true)
		} else {
			err = s.repo.UpdatePayHereStatus(ctx, tx, record.PaymentID, record.BookingID, "FAILED", notification.PaymentID, true)
		}
		if err != nil {
			return apperror.Wrap(err)
		}
	case "-3":
		if record.GroupID != nil {
			err = s.repo.UpdatePayHereGroupStatus(ctx, tx, record.PaymentID, *record.GroupID, "DISPUTED", notification.PaymentID, true)
		} else {
			err = s.repo.UpdatePayHereStatus(ctx, tx, record.PaymentID, record.BookingID, "DISPUTED", notification.PaymentID, true)
		}
		if err != nil {
			return apperror.Wrap(err)
		}
	default:
		return apperror.Validation("status_code", "Unsupported PayHere status code.")
	}
	if err = s.repo.MarkWebhookProcessed(ctx, tx, notification.EventFingerprint, ""); err != nil {
		return apperror.Wrap(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return apperror.Wrap(err)
	}
	return nil
}

func (s *Service) PayHereStatus(ctx context.Context, paymentID uuid.UUID, bookingToken string) (PayHerePaymentStatus, error) {
	record, err := s.repo.LoadPayHereStatus(ctx, s.pool, paymentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return PayHerePaymentStatus{}, apperror.New(404, "PAYMENT_NOT_FOUND", "Payment was not found.", nil)
	}
	if err != nil {
		return PayHerePaymentStatus{}, apperror.Wrap(err)
	}
	tokenTarget := record.BookingID
	if record.GroupID != nil {
		tokenTarget = *record.GroupID
	}
	if err = s.access.Verify(bookingToken, tokenTarget); err != nil {
		return PayHerePaymentStatus{}, apperror.New(401, "BOOKING_ACCESS_DENIED", "Booking access verification is required.", nil)
	}
	result := PayHerePaymentStatus{PaymentID: record.PaymentID, BookingID: record.BookingID, Status: record.Status, ProviderReference: record.ProviderReference, AmountMinor: record.AmountMinor, Currency: record.Currency, PaidAt: record.PaidAt}
	if record.GroupID != nil {
		group, groupErr := s.loadGroup(ctx, s.pool, *record.GroupID)
		if groupErr != nil {
			return PayHerePaymentStatus{}, apperror.Wrap(groupErr)
		}
		group.ManagementToken = s.access.Sign(group.ID)
		result.Group = &group
		result.BookingID = group.ID
		tickets, ticketErr := s.repo.LoadGroupTickets(ctx, s.pool, group.ID)
		if ticketErr != nil {
			return PayHerePaymentStatus{}, apperror.Wrap(ticketErr)
		}
		for index := range tickets {
			tickets[index].VerificationCode = s.tickets.Code(tickets[index].ID)
		}
		result.Tickets = tickets
		return result, nil
	}
	booked, err := s.bookings.Load(ctx, s.pool, record.BookingID)
	if err != nil {
		return PayHerePaymentStatus{}, apperror.Wrap(err)
	}
	booked.ManagementToken = s.access.Sign(booked.ID)
	result.Booking = &booked
	if record.TicketID != nil {
		status := "ACTIVE"
		if record.TicketStatus != nil {
			status = *record.TicketStatus
		}
		result.Ticket = &Ticket{ID: *record.TicketID, VerificationCode: s.tickets.Code(*record.TicketID), Status: status}
	}
	return result, nil
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
