package waitlist

import (
	"context"
	"encoding/base32"
	"errors"
	"net/mail"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/booking"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/database"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/journey"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
)

var phonePattern = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)

type Service struct {
	repo     *Repository
	journeys *journey.Service
	access   *booking.AccessSigner
}

func NewService(repo *Repository, journeys *journey.Service, access *booking.AccessSigner) *Service {
	return &Service{repo: repo, journeys: journeys, access: access}
}

func (s *Service) Create(ctx context.Context, request CreateRequest, accountID *uuid.UUID) (Entry, error) {
	segment, err := s.journeys.Resolve(ctx, request.TrainRunID, request.OriginStationID, request.DestinationStationID)
	if err != nil {
		return Entry{}, err
	}
	entry, err := validateCreate(request)
	if err != nil {
		return Entry{}, err
	}
	available, err := s.repo.HasAvailable(ctx, request.TrainRunID, segment, entry.PreferredCoachClass)
	if err != nil {
		return Entry{}, apperror.Wrap(err)
	}
	if available {
		return Entry{}, apperror.New(409, "SEATS_CURRENTLY_AVAILABLE", "Seats are currently available for this segment. Book a seat instead of joining the waitlist.", nil)
	}
	entry.ID = uuid.New()
	entry.Reference = waitlistReference(entry.ID)
	entry.TrainRunID = request.TrainRunID
	entry.OriginStationID = request.OriginStationID
	entry.DestinationStationID = request.DestinationStationID
	entry.OriginPosition = segment.OriginPosition
	entry.DestinationPosition = segment.DestinationPosition
	entry.Status = "WAITING"
	entry, err = s.repo.Insert(ctx, entry, accountID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Entry{}, apperror.New(409, "WAITLIST_ENTRY_EXISTS", "This contact is already waiting for the selected journey.", nil)
		}
		return Entry{}, apperror.Wrap(err)
	}
	entry.ManagementToken = s.access.Sign(entry.ID)
	return entry, nil
}

func (s *Service) Access(ctx context.Context, request AccessRequest) (Entry, error) {
	reference := strings.ToUpper(strings.TrimSpace(request.Reference))
	contact := normalizeContact(request.Contact)
	if reference == "" || contact == "" {
		return Entry{}, apperror.Validation("access", "Waitlist reference and email or phone are required.")
	}
	entry, err := s.repo.FindByReferenceAndContact(ctx, reference, contact)
	if errors.Is(err, pgx.ErrNoRows) {
		return Entry{}, apperror.New(404, "WAITLIST_ENTRY_NOT_FOUND", "Waitlist entry was not found.", nil)
	}
	if err != nil {
		return Entry{}, apperror.Wrap(err)
	}
	entry.ManagementToken = s.access.Sign(entry.ID)
	return entry, nil
}

func (s *Service) Cancel(ctx context.Context, id uuid.UUID, token string) (Entry, error) {
	if s.access.Verify(token, id) != nil {
		return Entry{}, apperror.New(401, "WAITLIST_ACCESS_DENIED", "Waitlist access verification is required.", nil)
	}
	entry, err := s.repo.Cancel(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Entry{}, apperror.New(409, "WAITLIST_NOT_ACTIVE", "This waitlist entry is no longer active.", nil)
	}
	if err != nil {
		return Entry{}, apperror.Wrap(err)
	}
	return entry, nil
}

func (s *Service) NotifyNext(ctx context.Context, db database.DBTX, bookingID uuid.UUID, requestID string) error {
	err := s.repo.NotifyNext(ctx, db, bookingID, requestID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	return err
}

func validateCreate(request CreateRequest) (Entry, error) {
	entry := Entry{
		FullName: strings.TrimSpace(request.FullName), Email: strings.ToLower(strings.TrimSpace(request.Email)),
		Phone: normalizePhone(request.Phone), PreferredCoachClass: strings.ToUpper(strings.TrimSpace(request.PreferredCoachClass)),
	}
	if len(entry.FullName) < 2 || len(entry.FullName) > 120 {
		return Entry{}, apperror.Validation("fullName", "Full name must contain 2 to 120 characters.")
	}
	if entry.Email != "" {
		address, err := mail.ParseAddress(entry.Email)
		if err != nil || address.Address != entry.Email || len(entry.Email) > 254 {
			return Entry{}, apperror.Validation("email", "Enter a valid email address.")
		}
	}
	if entry.Phone != "" && !phonePattern.MatchString(entry.Phone) {
		return Entry{}, apperror.Validation("phone", "Enter a valid international phone number.")
	}
	if entry.Email == "" && entry.Phone == "" {
		return Entry{}, apperror.Validation("contact", "Enter an email address or phone number.")
	}
	if entry.PreferredCoachClass == "" {
		entry.PreferredCoachClass = "ANY"
	}
	if entry.PreferredCoachClass != "ANY" && entry.PreferredCoachClass != "FIRST" && entry.PreferredCoachClass != "SECOND" {
		return Entry{}, apperror.Validation("preferredCoachClass", "Choose ANY, FIRST, or SECOND.")
	}
	return entry, nil
}

func normalizeContact(value string) string {
	value = strings.TrimSpace(value)
	if strings.Contains(value, "@") {
		return strings.ToLower(value)
	}
	return normalizePhone(value)
}

func normalizePhone(value string) string {
	compact := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(strings.TrimSpace(value))
	switch {
	case strings.HasPrefix(compact, "0094"):
		return "+" + strings.TrimPrefix(compact, "00")
	case strings.HasPrefix(compact, "94") && len(compact) == 11:
		return "+" + compact
	case strings.HasPrefix(compact, "0") && len(compact) == 10:
		return "+94" + compact[1:]
	case len(compact) == 9:
		return "+94" + compact
	default:
		return compact
	}
}

func waitlistReference(id uuid.UUID) string {
	return "WL-" + base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(id[:])[:12]
}
