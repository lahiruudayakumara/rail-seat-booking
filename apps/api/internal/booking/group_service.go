package booking

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/passengerauth"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
)

func (s *Service) CreateGroup(ctx context.Context, request CreateGroupRequest, idempotencyKey string, payload []byte, requestID string) (BookingGroup, bool, error) {
	if len(request.Members) < 2 || len(request.Members) > 6 {
		return BookingGroup{}, false, apperror.Validation("members", "A group booking must contain 2 to 6 seats.")
	}
	if len(idempotencyKey) < 16 || len(idempotencyKey) > 128 {
		return BookingGroup{}, false, apperror.Validation("Idempotency-Key", "Header must contain 16 to 128 characters.")
	}
	seenHolds := make(map[uuid.UUID]bool, len(request.Members))
	seenSeats := make(map[uuid.UUID]bool, len(request.Members))
	for index := range request.Members {
		member := &request.Members[index]
		member.Passenger.FullName = strings.TrimSpace(member.Passenger.FullName)
		member.Passenger.Email = strings.ToLower(strings.TrimSpace(member.Passenger.Email))
		member.Passenger.Phone = strings.TrimSpace(member.Passenger.Phone)
		if member.Passenger.FullName == "" || (member.Passenger.Email == "" && member.Passenger.Phone == "") {
			return BookingGroup{}, false, apperror.Validation("members.passenger", "Every seat requires a passenger name and email or phone.")
		}
		if member.HoldID == uuid.Nil || seenHolds[member.HoldID] || member.SeatID == uuid.Nil || seenSeats[member.SeatID] {
			return BookingGroup{}, false, apperror.Validation("members", "Every group member must use a distinct seat hold and seat.")
		}
		if s.access.Verify(member.HoldToken, member.HoldID) != nil {
			return BookingGroup{}, false, apperror.New(401, "HOLD_ACCESS_DENIED", "Every group seat requires a valid hold.", nil)
		}
		seenHolds[member.HoldID], seenSeats[member.SeatID] = true, true
	}

	keySum := sha256.Sum256([]byte(idempotencyKey))
	payloadSum := sha256.Sum256(payload)
	keyHash, requestHash := hex.EncodeToString(keySum[:]), hex.EncodeToString(payloadSum[:])
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return BookingGroup{}, false, apperror.Wrap(err)
	}
	defer tx.Rollback(ctx)
	reserved, err := s.repo.ReserveGroupIdempotency(ctx, tx, keyHash, requestHash)
	if err != nil {
		return BookingGroup{}, false, apperror.Wrap(err)
	}
	if !reserved {
		existingHash, groupID, lookupErr := s.repo.GroupIdempotency(ctx, tx, keyHash)
		if lookupErr != nil {
			return BookingGroup{}, false, apperror.Wrap(lookupErr)
		}
		if existingHash != requestHash {
			return BookingGroup{}, false, apperror.New(409, "IDEMPOTENCY_CONFLICT", "This idempotency key was used with a different group request.", nil)
		}
		if groupID != nil {
			if err = tx.Commit(ctx); err != nil {
				return BookingGroup{}, false, apperror.Wrap(err)
			}
			group, loadErr := s.GetGroup(ctx, *groupID)
			if loadErr == nil {
				group.ManagementToken = s.access.Sign(group.ID)
			}
			return group, true, loadErr
		}
	}

	// Lock holds in deterministic order to avoid deadlocks between competing group requests.
	members := append([]GroupMemberRequest(nil), request.Members...)
	sort.Slice(members, func(i, j int) bool { return members[i].HoldID.String() < members[j].HoldID.String() })
	var firstQuote QuoteSnapshot
	var total int64
	var currency string
	var scale int16
	var earliestExpiry time.Time
	for index, member := range members {
		quote, status, expiresAt, lockErr := s.repo.LockHold(ctx, tx, member.HoldID)
		if errors.Is(lockErr, pgx.ErrNoRows) {
			return BookingGroup{}, false, apperror.New(404, "HOLD_NOT_FOUND", "A group seat hold was not found.", nil)
		}
		if lockErr != nil {
			return BookingGroup{}, false, apperror.Wrap(lockErr)
		}
		if status != "HELD" || !time.Now().Before(expiresAt) {
			return BookingGroup{}, false, apperror.New(409, "HOLD_EXPIRED", "A group seat hold expired. Select the seats again.", nil)
		}
		if quote.TrainRunID != member.TrainRunID || quote.SeatID != member.SeatID || quote.OriginStationID != member.OriginStationID || quote.DestinationStationID != member.DestinationStationID {
			return BookingGroup{}, false, apperror.Validation("members.holdId", "A seat hold does not match its group member.")
		}
		if index > 0 && (quote.TrainRunID != firstQuote.TrainRunID || quote.OriginStationID != firstQuote.OriginStationID || quote.DestinationStationID != firstQuote.DestinationStationID) {
			return BookingGroup{}, false, apperror.Validation("members", "All group seats must be on the same train and journey segment.")
		}
		if index == 0 {
			currency, scale, earliestExpiry = quote.Currency, quote.CurrencyScale, expiresAt
			firstQuote = quote
		} else if quote.Currency != currency || quote.CurrencyScale != scale {
			return BookingGroup{}, false, apperror.Validation("members", "All group fares must use the same currency.")
		}
		if expiresAt.Before(earliestExpiry) {
			earliestExpiry = expiresAt
		}
		total += quote.AmountMinor
	}

	groupID := uuid.New()
	accountID := passengerauth.AccountID(ctx)
	if err = s.repo.InsertGroup(ctx, tx, groupID, groupReference(groupID), accountID, total, currency, scale, earliestExpiry); err != nil {
		return BookingGroup{}, false, apperror.Wrap(err)
	}
	for _, member := range members {
		passengerID := uuid.New()
		if err = s.repo.InsertPassenger(ctx, tx, passengerID, accountID, member.Passenger); err != nil {
			return BookingGroup{}, false, apperror.Wrap(err)
		}
		if err = s.repo.PrepareHeldBookingGroup(ctx, tx, member.HoldID, passengerID, groupID, bookingReference(member.HoldID)); err != nil {
			return BookingGroup{}, false, apperror.Wrap(err)
		}
	}
	// The first requested member is the explicit lead passenger. Lock ordering must
	// never change who supplies the payment-provider contact details.
	if err = s.repo.SetGroupLead(ctx, tx, groupID, request.Members[0].HoldID); err != nil {
		return BookingGroup{}, false, apperror.Wrap(err)
	}
	if err = s.repo.InsertGroupAudit(ctx, tx, uuid.New(), groupID, "GROUP_PENDING_PAYMENT", requestID); err != nil {
		return BookingGroup{}, false, apperror.Wrap(err)
	}
	if err = s.repo.AttachGroupIdempotency(ctx, tx, keyHash, groupID); err != nil {
		return BookingGroup{}, false, apperror.Wrap(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return BookingGroup{}, false, apperror.Wrap(err)
	}
	group, err := s.GetGroup(ctx, groupID)
	if err == nil {
		group.ManagementToken = s.access.Sign(group.ID)
	}
	return group, false, err
}

func (s *Service) GetGroup(ctx context.Context, groupID uuid.UUID) (BookingGroup, error) {
	group, err := s.repo.LoadGroupHeader(ctx, s.pool, groupID)
	if errors.Is(err, pgx.ErrNoRows) {
		return BookingGroup{}, apperror.New(404, "BOOKING_GROUP_NOT_FOUND", "Booking group was not found.", nil)
	}
	if err != nil {
		return BookingGroup{}, apperror.Wrap(err)
	}
	ids, err := s.repo.GroupBookingIDs(ctx, s.pool, groupID)
	if err != nil {
		return BookingGroup{}, apperror.Wrap(err)
	}
	group.Members = make([]GroupMember, 0, len(ids))
	for _, bookingID := range ids {
		booked, loadErr := s.repo.Load(ctx, s.pool, bookingID)
		if loadErr != nil {
			return BookingGroup{}, apperror.Wrap(loadErr)
		}
		passenger, passengerErr := s.repo.BookingPassenger(ctx, s.pool, bookingID)
		if passengerErr != nil {
			return BookingGroup{}, apperror.Wrap(passengerErr)
		}
		group.Members = append(group.Members, GroupMember{Booking: booked, Passenger: passenger})
	}
	return group, nil
}

func groupReference(id uuid.UUID) string {
	return "GR-" + strings.TrimRight(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(id[:7]), "=")
}
