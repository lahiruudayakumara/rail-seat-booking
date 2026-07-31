package booking

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
)

type Service struct {
	pool *pgxpool.Pool
	repo *Repository
}

func NewService(pool *pgxpool.Pool, repo *Repository) *Service {
	return &Service{pool: pool, repo: repo}
}
func (s *Service) Create(ctx context.Context, request CreateRequest, idempotencyKey string, payload []byte, requestID string) (Booking, bool, error) {
	if len(idempotencyKey) < 16 || len(idempotencyKey) > 128 {
		return Booking{}, false, apperror.Validation("Idempotency-Key", "Header must contain 16 to 128 characters.")
	}
	request.Passenger.FullName = strings.TrimSpace(request.Passenger.FullName)
	request.Passenger.Email = strings.TrimSpace(request.Passenger.Email)
	request.Passenger.Phone = strings.TrimSpace(request.Passenger.Phone)
	if request.Passenger.FullName == "" || (request.Passenger.Email == "" && request.Passenger.Phone == "") {
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
			return item, true, err
		}
	}
	quote, err := s.repo.Quote(ctx, tx, request.FareQuoteID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Booking{}, false, apperror.New(422, "VALIDATION_ERROR", "Fare quote is invalid or expired.", nil)
	}
	if err != nil {
		return Booking{}, false, apperror.Wrap(err)
	}
	if quote.TrainRunID != request.TrainRunID || quote.SeatID != request.SeatID || quote.OriginStationID != request.OriginStationID || quote.DestinationStationID != request.DestinationStationID {
		return Booking{}, false, apperror.Validation("fareQuoteId", "Fare quote does not match the booking.")
	}
	passengerID, bookingID := uuid.New(), uuid.New()
	if err = s.repo.InsertPassenger(ctx, tx, passengerID, request.Passenger); err != nil {
		return Booking{}, false, apperror.Wrap(err)
	}
	if err = s.repo.InsertBooking(ctx, tx, bookingID, bookingReference(bookingID), passengerID, quote); err != nil {
		if IsExclusionViolation(err) {
			return Booking{}, false, apperror.New(409, "SEAT_NO_LONGER_AVAILABLE", "This seat is no longer available for the selected journey segment.", map[string]any{"seatId": request.SeatID, "trainRunId": request.TrainRunID})
		}
		return Booking{}, false, apperror.Wrap(err)
	}
	if err = s.repo.InsertAudit(ctx, tx, uuid.New(), bookingID, "BOOKING_CONFIRMED", requestID); err != nil {
		return Booking{}, false, apperror.Wrap(err)
	}
	if err = s.repo.AttachIdempotency(ctx, tx, keyHash, bookingID); err != nil {
		return Booking{}, false, apperror.Wrap(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Booking{}, false, apperror.Wrap(err)
	}
	item, err := s.Get(ctx, bookingID)
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
func (s *Service) GetByReference(ctx context.Context, reference string) (Booking, error) {
	id, err := s.repo.FindIDByReference(ctx, s.pool, strings.ToUpper(strings.TrimSpace(reference)))
	if errors.Is(err, pgx.ErrNoRows) {
		return Booking{}, apperror.New(404, "BOOKING_NOT_FOUND", "Booking was not found.", nil)
	}
	if err != nil {
		return Booking{}, apperror.Wrap(err)
	}
	return s.Get(ctx, id)
}
func (s *Service) Cancel(ctx context.Context, id uuid.UUID, requestID string) (Booking, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Booking{}, apperror.Wrap(err)
	}
	defer tx.Rollback(ctx)
	status, err := s.repo.LockStatus(ctx, tx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Booking{}, apperror.New(404, "BOOKING_NOT_FOUND", "Booking was not found.", nil)
	}
	if err != nil {
		return Booking{}, apperror.Wrap(err)
	}
	if status == "COMPLETED" || status == "EXPIRED" {
		return Booking{}, apperror.New(422, "BOOKING_CANNOT_BE_CANCELLED", "This booking can no longer be cancelled.", nil)
	}
	if status != "CANCELLED" {
		if err = s.repo.Cancel(ctx, tx, id); err != nil {
			return Booking{}, apperror.Wrap(err)
		}
		if err = s.repo.InsertAudit(ctx, tx, uuid.New(), id, "BOOKING_CANCELLED", requestID); err != nil {
			return Booking{}, apperror.Wrap(err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return Booking{}, apperror.Wrap(err)
	}
	return s.Get(ctx, id)
}
func bookingReference(id uuid.UUID) string {
	return "BK-" + strings.TrimRight(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(id[:6]), "=")
}
