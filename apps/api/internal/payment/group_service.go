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
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/booking"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/database"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
)

func (s *Service) CheckoutGroup(ctx context.Context, request GroupCheckoutRequest, idempotencyKey, requestID string) (GroupCheckoutResult, error) {
	if len(idempotencyKey) < 16 || len(idempotencyKey) > 128 {
		return GroupCheckoutResult{}, apperror.Validation("Idempotency-Key", "Header must contain 16 to 128 characters.")
	}
	if err := s.access.Verify(request.GroupToken, request.GroupID); err != nil {
		return GroupCheckoutResult{}, apperror.New(401, "BOOKING_GROUP_ACCESS_DENIED", "Group booking access verification is required.", nil)
	}
	keyDigest := sha256.Sum256([]byte(idempotencyKey))
	keyHash := hex.EncodeToString(keyDigest[:])
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return GroupCheckoutResult{}, apperror.Wrap(err)
	}
	defer tx.Rollback(ctx)
	if existing, findErr := s.repo.FindGroupCheckout(ctx, tx, keyHash); findErr == nil {
		if existing.GroupID != request.GroupID {
			return GroupCheckoutResult{}, apperror.New(409, "IDEMPOTENCY_KEY_REUSED", "The idempotency key belongs to a different group checkout.", nil)
		}
		return s.groupCheckoutResult(ctx, tx, existing)
	} else if !errors.Is(findErr, pgx.ErrNoRows) {
		return GroupCheckoutResult{}, apperror.Wrap(findErr)
	}
	payable, bookingIDs, err := s.repo.LockGroupPayable(ctx, tx, request.GroupID)
	if errors.Is(err, pgx.ErrNoRows) {
		return GroupCheckoutResult{}, apperror.New(404, "BOOKING_GROUP_NOT_FOUND", "Booking group was not found.", nil)
	}
	if err != nil {
		return GroupCheckoutResult{}, apperror.Wrap(err)
	}
	if payable.Status != "HELD" || !time.Now().Before(payable.ExpiresAt) {
		return GroupCheckoutResult{}, apperror.New(409, "HOLD_EXPIRED", "One or more group seat holds expired. Select the seats again.", nil)
	}
	if len(bookingIDs) < 2 {
		return GroupCheckoutResult{}, apperror.New(409, "GROUP_INCOMPLETE", "The group booking does not contain enough seats.", nil)
	}
	paymentID := uuid.New()
	providerResult, err := s.provider.Charge(ctx, paymentID, payable.AmountMinor, payable.Currency)
	if err != nil || providerResult.Status != "PAID" {
		return GroupCheckoutResult{}, apperror.New(502, "PAYMENT_FAILED", "The group payment could not be completed.", nil)
	}
	issues := make([]ticketIssue, 0, len(bookingIDs))
	tickets := make([]Ticket, 0, len(bookingIDs))
	for _, bookingID := range bookingIDs {
		ticketID := uuid.New()
		verificationCode := s.tickets.Code(ticketID)
		codeHash := sha256.Sum256([]byte(verificationCode))
		issues = append(issues, ticketIssue{ID: ticketID, BookingID: bookingID, CodeHash: hex.EncodeToString(codeHash[:])})
		tickets = append(tickets, Ticket{ID: ticketID, VerificationCode: verificationCode, Status: "ACTIVE"})
	}
	payment := Payment{ID: paymentID, Status: "PAID", Provider: "SANDBOX", ProviderReference: providerResult.Reference, AmountMinor: payable.AmountMinor, Currency: payable.Currency, PaidAt: time.Now()}
	if err = s.repo.CompleteGroup(ctx, tx, payment, request.GroupID, payable.LeadBooking, issues, keyHash, requestID); err != nil {
		return GroupCheckoutResult{}, apperror.Wrap(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return GroupCheckoutResult{}, apperror.Wrap(err)
	}
	group, err := s.loadGroup(ctx, s.pool, request.GroupID)
	if err != nil {
		return GroupCheckoutResult{}, apperror.Wrap(err)
	}
	group.ManagementToken = s.access.Sign(group.ID)
	return GroupCheckoutResult{Payment: payment, Tickets: tickets, Group: group}, nil
}

func (s *Service) StartPayHereGroupCheckout(ctx context.Context, request PayHereGroupCheckoutRequest, idempotencyKey string) (PayHereCheckoutResponse, error) {
	if s.payHere == nil || !s.payHere.Enabled() {
		return PayHereCheckoutResponse{}, apperror.New(503, "PAYHERE_NOT_CONFIGURED", "PayHere checkout is not configured.", nil)
	}
	if err := s.payHere.ValidateNotifyURL(ctx); err != nil {
		return PayHereCheckoutResponse{}, apperror.New(503, "PAYHERE_CALLBACK_UNREACHABLE", "PayHere checkout is temporarily unavailable because its verified callback cannot reach this server. Start or repair the public HTTPS tunnel, then try again; no payment was created.", nil)
	}
	request.BillingAddress = strings.TrimSpace(request.BillingAddress)
	request.City = strings.TrimSpace(request.City)
	if len(request.BillingAddress) < 3 || len(request.BillingAddress) > 200 || len(request.City) < 2 || len(request.City) > 80 {
		return PayHereCheckoutResponse{}, apperror.Validation("billingAddress", "Billing address and city are required for PayHere checkout.")
	}
	if len(idempotencyKey) < 16 || len(idempotencyKey) > 128 {
		return PayHereCheckoutResponse{}, apperror.Validation("Idempotency-Key", "Header must contain 16 to 128 characters.")
	}
	if err := s.access.Verify(request.GroupToken, request.GroupID); err != nil {
		return PayHereCheckoutResponse{}, apperror.New(401, "BOOKING_GROUP_ACCESS_DENIED", "Group booking access verification is required.", nil)
	}
	keyDigest := sha256.Sum256([]byte(idempotencyKey))
	keyHash := hex.EncodeToString(keyDigest[:])
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return PayHereCheckoutResponse{}, apperror.Wrap(err)
	}
	defer tx.Rollback(ctx)
	if existing, findErr := s.repo.FindPayHereCheckoutByKey(ctx, tx, keyHash); findErr == nil {
		if existing.GroupID == nil || *existing.GroupID != request.GroupID {
			return PayHereCheckoutResponse{}, apperror.New(409, "IDEMPOTENCY_KEY_REUSED", "The idempotency key belongs to a different checkout.", nil)
		}
		if err = tx.Commit(ctx); err != nil {
			return PayHereCheckoutResponse{}, apperror.Wrap(err)
		}
		return s.payHereGroupCheckoutResponse(existing, request)
	} else if !errors.Is(findErr, pgx.ErrNoRows) {
		return PayHereCheckoutResponse{}, apperror.Wrap(findErr)
	}
	payable, bookingIDs, err := s.repo.LockGroupPayable(ctx, tx, request.GroupID)
	if errors.Is(err, pgx.ErrNoRows) {
		return PayHereCheckoutResponse{}, apperror.New(404, "BOOKING_GROUP_NOT_FOUND", "Booking group was not found.", nil)
	}
	if err != nil {
		return PayHereCheckoutResponse{}, apperror.Wrap(err)
	}
	if payable.Status != "HELD" || !time.Now().Before(payable.ExpiresAt) || len(bookingIDs) < 2 {
		return PayHereCheckoutResponse{}, apperror.New(409, "HOLD_EXPIRED", "One or more group seat holds expired. Select the seats again.", nil)
	}
	if payable.Email == "" || payable.Phone == "" {
		return PayHereCheckoutResponse{}, apperror.Validation("passenger", "PayHere requires the lead passenger to have both an email address and phone number.")
	}
	if existing, findErr := s.repo.FindPendingPayHereByGroup(ctx, tx, request.GroupID); findErr == nil {
		if err = tx.Commit(ctx); err != nil {
			return PayHereCheckoutResponse{}, apperror.Wrap(err)
		}
		return s.payHereGroupCheckoutResponse(existing, request)
	} else if !errors.Is(findErr, pgx.ErrNoRows) {
		return PayHereCheckoutResponse{}, apperror.Wrap(findErr)
	}
	paymentID := uuid.New()
	expiresAt := time.Now().Add(s.paymentPendingTTL)
	groupID := request.GroupID
	record := payHereCheckoutRecord{PaymentID: paymentID, BookingID: payable.LeadBooking, GroupID: &groupID, Status: "PENDING", BookingStatus: "HELD", AmountMinor: payable.AmountMinor, Currency: payable.Currency, ExpiresAt: expiresAt, FullName: payable.FullName, Email: payable.Email, Phone: payable.Phone}
	if err = s.repo.InsertPayHereGroupPending(ctx, tx, paymentID, request.GroupID, payable.LeadBooking, payable.AmountMinor, payable.Currency, keyHash, expiresAt); err != nil {
		return PayHereCheckoutResponse{}, apperror.Wrap(err)
	}
	response, err := s.payHereGroupCheckoutResponse(record, request)
	if err != nil {
		return PayHereCheckoutResponse{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return PayHereCheckoutResponse{}, apperror.Wrap(err)
	}
	return response, nil
}

func (s *Service) payHereGroupCheckoutResponse(record payHereCheckoutRecord, request PayHereGroupCheckoutRequest) (PayHereCheckoutResponse, error) {
	session, err := s.payHere.CreateSession(record.PaymentID, request.GroupID, PayHereCustomer{FullName: record.FullName, Email: record.Email, Phone: record.Phone, Address: request.BillingAddress, City: request.City}, record.AmountMinor, record.Currency)
	if err != nil {
		return PayHereCheckoutResponse{}, apperror.New(422, "PAYHERE_CHECKOUT_INVALID", err.Error(), nil)
	}
	groupID := request.GroupID
	return PayHereCheckoutResponse{PaymentID: record.PaymentID, BookingID: request.GroupID, GroupID: &groupID, Status: record.Status, ExpiresAt: record.ExpiresAt, ActionURL: session.ActionURL, Fields: session.Fields}, nil
}

func (s *Service) groupCheckoutResult(ctx context.Context, tx pgx.Tx, existing groupCheckoutRecord) (GroupCheckoutResult, error) {
	group, err := s.loadGroup(ctx, tx, existing.GroupID)
	if err != nil {
		return GroupCheckoutResult{}, apperror.Wrap(err)
	}
	tickets, err := s.repo.LoadGroupTickets(ctx, tx, existing.GroupID)
	if err != nil {
		return GroupCheckoutResult{}, apperror.Wrap(err)
	}
	for index := range tickets {
		tickets[index].VerificationCode = s.tickets.Code(tickets[index].ID)
	}
	group.ManagementToken = s.access.Sign(group.ID)
	if err = tx.Commit(ctx); err != nil {
		return GroupCheckoutResult{}, apperror.Wrap(err)
	}
	return GroupCheckoutResult{Payment: existing.Payment, Tickets: tickets, Group: group}, nil
}

func (s *Service) loadGroup(ctx context.Context, db database.DBTX, groupID uuid.UUID) (booking.BookingGroup, error) {
	group, err := s.bookings.LoadGroupHeader(ctx, db, groupID)
	if err != nil {
		return booking.BookingGroup{}, err
	}
	ids, err := s.bookings.GroupBookingIDs(ctx, db, groupID)
	if err != nil {
		return booking.BookingGroup{}, err
	}
	group.Members = make([]booking.GroupMember, 0, len(ids))
	for _, bookingID := range ids {
		booked, loadErr := s.bookings.Load(ctx, db, bookingID)
		if loadErr != nil {
			return booking.BookingGroup{}, loadErr
		}
		passenger, passengerErr := s.bookings.BookingPassenger(ctx, db, bookingID)
		if passengerErr != nil {
			return booking.BookingGroup{}, passengerErr
		}
		group.Members = append(group.Members, booking.GroupMember{Booking: booked, Passenger: passenger})
	}
	return group, nil
}
