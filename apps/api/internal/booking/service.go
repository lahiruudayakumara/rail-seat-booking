package booking

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/passengerauth"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
)

type Service struct {
	pool         *pgxpool.Pool
	repo         *Repository
	access       *AccessSigner
	holdTTL      time.Duration
	cancellation CancellationProcessor
}

func NewService(pool *pgxpool.Pool, repo *Repository, access *AccessSigner, cancellation CancellationProcessor, holdTTL time.Duration) *Service {
	return &Service{pool: pool, repo: repo, access: access, cancellation: cancellation, holdTTL: holdTTL}
}

func (s *Service) CreateHold(ctx context.Context, request HoldRequest, requestID string) (Hold, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Hold{}, apperror.Wrap(err)
	}
	defer tx.Rollback(ctx)
	if err = s.repo.ExpireHolds(ctx, tx); err != nil {
		return Hold{}, apperror.Wrap(err)
	}
	quote, err := s.repo.Quote(ctx, tx, request.FareQuoteID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Hold{}, apperror.New(422, "VALIDATION_ERROR", "Fare quote is invalid or expired.", nil)
	}
	if err != nil {
		return Hold{}, apperror.Wrap(err)
	}
	holdID, expiresAt := uuid.New(), time.Now().Add(s.holdTTL)
	if err = s.repo.InsertHold(ctx, tx, holdID, quote, expiresAt); err != nil {
		if IsExclusionViolation(err) {
			return Hold{}, apperror.New(409, "SEAT_NO_LONGER_AVAILABLE", "This seat is no longer available for the selected journey segment.", nil)
		}
		return Hold{}, apperror.Wrap(err)
	}
	if err = s.repo.InsertAudit(ctx, tx, uuid.New(), holdID, "SEAT_HELD", requestID); err != nil {
		return Hold{}, apperror.Wrap(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Hold{}, apperror.Wrap(err)
	}
	return Hold{ID: holdID, Status: "HELD", ExpiresAt: expiresAt, ManagementToken: s.access.Sign(holdID)}, nil
}

func (s *Service) ReleaseHold(ctx context.Context, holdID uuid.UUID, token, requestID string) error {
	if holdID == uuid.Nil || s.access.Verify(token, holdID) != nil {
		return apperror.New(401, "HOLD_ACCESS_DENIED", "A valid seat hold is required.", nil)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return apperror.Wrap(err)
	}
	defer tx.Rollback(ctx)
	released, err := s.repo.ReleaseUnusedHold(ctx, tx, holdID)
	if err != nil {
		return apperror.Wrap(err)
	}
	if released {
		if err = s.repo.InsertAudit(ctx, tx, uuid.New(), holdID, "SEAT_HOLD_RELEASED", requestID); err != nil {
			return apperror.Wrap(err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return apperror.Wrap(err)
	}
	return nil
}

func (s *Service) Create(ctx context.Context, request CreateRequest, idempotencyKey string, payload []byte, requestID string) (Booking, bool, error) {
	if len(idempotencyKey) < 16 || len(idempotencyKey) > 128 {
		return Booking{}, false, apperror.Validation("Idempotency-Key", "Header must contain 16 to 128 characters.")
	}
	request.Passenger.FullName = strings.TrimSpace(request.Passenger.FullName)
	request.Passenger.Email = strings.TrimSpace(request.Passenger.Email)
	request.Passenger.Phone = strings.TrimSpace(request.Passenger.Phone)
	accountID := passengerauth.AccountID(ctx)
	if accountID == nil && (request.Passenger.FullName == "" || (request.Passenger.Email == "" && request.Passenger.Phone == "")) {
		return Booking{}, false, apperror.Validation("passenger", "Name and email or phone are required.")
	}
	keySum := sha256.Sum256([]byte(idempotencyKey))
	payloadSum := sha256.Sum256(payload)
	keyHash, requestHash := hex.EncodeToString(keySum[:]), hex.EncodeToString(payloadSum[:])
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Booking{}, false, apperror.Wrap(err)
	}
	defer tx.Rollback(ctx)
	reserved, err := s.repo.ReserveIdempotency(ctx, tx, keyHash, requestHash)
	if err != nil {
		return Booking{}, false, apperror.Wrap(err)
	}
	if !reserved {
		existingHash, bookingID, err := s.repo.Idempotency(ctx, tx, keyHash)
		if err != nil {
			return Booking{}, false, apperror.Wrap(err)
		}
		if existingHash != requestHash {
			return Booking{}, false, apperror.New(409, "IDEMPOTENCY_CONFLICT", "This idempotency key was used with a different request.", nil)
		}
		if bookingID != nil {
			if err = tx.Commit(ctx); err != nil {
				return Booking{}, false, apperror.Wrap(err)
			}
			item, err := s.Get(ctx, *bookingID)
			if err == nil {
				item.ManagementToken = s.access.Sign(item.ID)
			}
			return item, true, err
		}
	}
	if request.HoldID == uuid.Nil || s.access.Verify(request.HoldToken, request.HoldID) != nil {
		return Booking{}, false, apperror.New(401, "HOLD_ACCESS_DENIED", "A valid seat hold is required.", nil)
	}
	quote, status, expiresAt, err := s.repo.LockHold(ctx, tx, request.HoldID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Booking{}, false, apperror.New(404, "HOLD_NOT_FOUND", "Seat hold was not found.", nil)
	}
	if err != nil {
		return Booking{}, false, apperror.Wrap(err)
	}
	if status != "HELD" || !time.Now().Before(expiresAt) {
		return Booking{}, false, apperror.New(409, "HOLD_EXPIRED", "The seat hold has expired. Select the seat again.", nil)
	}
	if quote.TrainRunID != request.TrainRunID || quote.SeatID != request.SeatID || quote.OriginStationID != request.OriginStationID || quote.DestinationStationID != request.DestinationStationID {
		return Booking{}, false, apperror.Validation("holdId", "Seat hold does not match the booking.")
	}
	passengerID, bookingID := uuid.New(), request.HoldID
	if err = s.repo.InsertPassenger(ctx, tx, passengerID, accountID, request.Passenger); err != nil {
		return Booking{}, false, apperror.Wrap(err)
	}
	if err = s.repo.PrepareHeldBooking(ctx, tx, bookingID, passengerID, bookingReference(bookingID)); err != nil {
		return Booking{}, false, apperror.Wrap(err)
	}
	if err = s.repo.InsertAudit(ctx, tx, uuid.New(), bookingID, "BOOKING_PENDING_PAYMENT", requestID); err != nil {
		return Booking{}, false, apperror.Wrap(err)
	}
	if err = s.repo.AttachIdempotency(ctx, tx, keyHash, bookingID); err != nil {
		return Booking{}, false, apperror.Wrap(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Booking{}, false, apperror.Wrap(err)
	}
	item, err := s.Get(ctx, bookingID)
	if err == nil {
		item.ManagementToken = s.access.Sign(item.ID)
	}
	return item, false, err
}
func (s *Service) Get(ctx context.Context, id uuid.UUID) (Booking, error) {
	item, err := s.repo.Load(ctx, s.pool, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Booking{}, apperror.New(404, "BOOKING_NOT_FOUND", "Booking was not found.", nil)
	}
	if err != nil {
		return Booking{}, apperror.Wrap(err)
	}
	return item, nil
}
func (s *Service) ListForAccount(ctx context.Context, accountID uuid.UUID) ([]Booking, error) {
	items, err := s.repo.ListByAccount(ctx, s.pool, accountID)
	if err != nil {
		return nil, apperror.Wrap(err)
	}
	return items, nil
}
func (s *Service) Access(ctx context.Context, request AccessRequest) (Booking, error) {
	reference := strings.ToUpper(strings.TrimSpace(request.Reference))
	contact := normalizeLookupContact(request.Contact)
	if reference == "" || contact == "" {
		return Booking{}, apperror.Validation("access", "Booking reference and email or phone are required.")
	}
	id, err := s.repo.FindIDByReferenceAndContact(ctx, s.pool, reference, contact)
	if errors.Is(err, pgx.ErrNoRows) {
		return Booking{}, apperror.New(404, "BOOKING_NOT_FOUND", "Booking was not found.", nil)
	}
	if err != nil {
		return Booking{}, apperror.Wrap(err)
	}
	item, err := s.Get(ctx, id)
	if err == nil {
		item.ManagementToken = s.access.Sign(item.ID)
	}
	return item, err
}

func normalizeLookupContact(value string) string {
	value = strings.TrimSpace(value)
	if strings.Contains(value, "@") {
		return strings.ToLower(value)
	}

	compact := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(value)
	switch {
	case strings.HasPrefix(compact, "0094"):
		return "+" + strings.TrimPrefix(compact, "00")
	case strings.HasPrefix(compact, "94") && len(compact) == 11:
		return "+" + compact
	case strings.HasPrefix(compact, "0") && len(compact) == 10:
		return "+94" + compact[1:]
	case len(compact) == 9 && compact[0] != '+':
		return "+94" + compact
	default:
		return compact
	}
}
func (s *Service) Cancel(ctx context.Context, id uuid.UUID, token, reason, requestID string) (Booking, error) {
	if err := s.access.Verify(token, id); err != nil {
		accountID := passengerauth.AccountID(ctx)
		if accountID == nil {
			return Booking{}, apperror.New(401, "BOOKING_ACCESS_DENIED", "Booking access verification is required.", nil)
		}
		belongs, ownershipErr := s.repo.BelongsToAccount(ctx, s.pool, id, *accountID)
		if ownershipErr != nil {
			return Booking{}, apperror.Wrap(ownershipErr)
		}
		if !belongs {
			return Booking{}, apperror.New(404, "BOOKING_NOT_FOUND", "Booking was not found.", nil)
		}
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Booking{}, apperror.Wrap(err)
	}
	defer tx.Rollback(ctx)
	status, departureAt, err := s.repo.LockStatus(ctx, tx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Booking{}, apperror.New(404, "BOOKING_NOT_FOUND", "Booking was not found.", nil)
	}
	if err != nil {
		return Booking{}, apperror.Wrap(err)
	}
	if status == "COMPLETED" || status == "EXPIRED" {
		return Booking{}, apperror.New(422, "BOOKING_CANNOT_BE_CANCELLED", "This booking can no longer be cancelled.", nil)
	}
	if status != "CANCELLED" && !time.Now().Before(departureAt) {
		return Booking{}, apperror.New(422, "DEPARTURE_PASSED", "Bookings cannot be cancelled after departure.", nil)
	}
	if status != "CANCELLED" {
		var refund *RefundSummary
		if s.cancellation != nil {
			refund, err = s.cancellation.Process(ctx, tx, id, strings.TrimSpace(reason), requestID)
			if err != nil {
				return Booking{}, err
			}
		}
		if err = s.repo.Cancel(ctx, tx, id); err != nil {
			return Booking{}, apperror.Wrap(err)
		}
		if err = s.repo.InsertAudit(ctx, tx, uuid.New(), id, "BOOKING_CANCELLED", requestID); err != nil {
			return Booking{}, apperror.Wrap(err)
		}
		if err = tx.Commit(ctx); err != nil {
			return Booking{}, apperror.Wrap(err)
		}
		item, loadErr := s.Get(ctx, id)
		item.Refund = refund
		return item, loadErr
	}
	if err = tx.Commit(ctx); err != nil {
		return Booking{}, apperror.Wrap(err)
	}
	return s.Get(ctx, id)
}
func bookingReference(id uuid.UUID) string {
	return "BK-" + strings.TrimRight(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(id[:6]), "=")
}
