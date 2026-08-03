package passengerauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
	"golang.org/x/crypto/bcrypt"
)

var phonePattern = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
var dummyPasswordHash = func() []byte {
	hash, _ := bcrypt.GenerateFromPassword([]byte("InvalidPassengerPassword1"), bcrypt.DefaultCost)
	return hash
}()

type Service struct {
	repo       *Repository
	sessionTTL time.Duration
}

func NewService(repo *Repository, sessionTTL time.Duration) *Service {
	return &Service{repo: repo, sessionTTL: sessionTTL}
}

func (s *Service) Register(ctx context.Context, request RegisterRequest) (Account, Session, error) {
	request.FullName = strings.TrimSpace(request.FullName)
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Phone = strings.TrimSpace(request.Phone)
	if len(request.FullName) < 2 || len(request.FullName) > 120 {
		return Account{}, Session{}, apperror.Validation("fullName", "Full name must contain 2 to 120 characters.")
	}
	address, emailErr := mail.ParseAddress(request.Email)
	if emailErr != nil || address.Address != request.Email || len(request.Email) > 254 {
		return Account{}, Session{}, apperror.Validation("email", "Enter a valid email address.")
	}
	if request.Phone != "" && !phonePattern.MatchString(request.Phone) {
		return Account{}, Session{}, apperror.Validation("phone", "Use international phone format, for example +94770000000.")
	}
	if err := validatePassword(request.Password); err != nil {
		return Account{}, Session{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return Account{}, Session{}, apperror.Wrap(err)
	}
	account := Account{ID: uuid.New(), FullName: request.FullName, Email: request.Email, Phone: request.Phone, CreatedAt: time.Now().UTC()}
	if err = s.repo.InsertAccount(ctx, account, string(hash)); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Account{}, Session{}, apperror.New(409, "ACCOUNT_ALREADY_EXISTS", "An account already exists for that email or phone.", nil)
		}
		return Account{}, Session{}, apperror.Wrap(err)
	}
	session, err := s.newSession(ctx, account.ID)
	return account, session, err
}

func (s *Service) Login(ctx context.Context, request LoginRequest) (Account, Session, error) {
	email := strings.ToLower(strings.TrimSpace(request.Email))
	account, hash, err := s.repo.FindByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(request.Password))
		return Account{}, Session{}, apperror.New(401, "INVALID_CREDENTIALS", "Email or password is incorrect.", nil)
	}
	if err == nil && bcrypt.CompareHashAndPassword([]byte(hash), []byte(request.Password)) != nil {
		return Account{}, Session{}, apperror.New(401, "INVALID_CREDENTIALS", "Email or password is incorrect.", nil)
	}
	if err != nil {
		return Account{}, Session{}, apperror.Wrap(err)
	}
	session, err := s.newSession(ctx, account.ID)
	return account, session, err
}

func (s *Service) Authenticate(ctx context.Context, token string) (Account, error) {
	if token == "" {
		return Account{}, pgx.ErrNoRows
	}
	account, err := s.repo.AccountBySession(ctx, hashToken(token))
	if err != nil {
		return Account{}, err
	}
	return account, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.repo.RevokeSession(ctx, hashToken(token))
}

func (s *Service) ListTravellers(ctx context.Context, accountID uuid.UUID) ([]Traveller, error) {
	travellers, err := s.repo.ListTravellers(ctx, accountID)
	if err != nil {
		return nil, apperror.Wrap(err)
	}
	return travellers, nil
}

func (s *Service) CreateTraveller(ctx context.Context, accountID uuid.UUID, request TravellerRequest) (Traveller, error) {
	traveller, err := validateTraveller(request)
	if err != nil {
		return Traveller{}, err
	}
	count, err := s.repo.CountTravellers(ctx, accountID)
	if err != nil {
		return Traveller{}, apperror.Wrap(err)
	}
	if count >= 20 {
		return Traveller{}, apperror.New(409, "TRAVELLER_LIMIT_REACHED", "A passenger account can save up to 20 travellers.", nil)
	}
	traveller.ID = uuid.New()
	traveller, err = s.repo.InsertTraveller(ctx, accountID, traveller)
	if err != nil {
		return Traveller{}, apperror.Wrap(err)
	}
	return traveller, nil
}

func (s *Service) UpdateTraveller(ctx context.Context, accountID, travellerID uuid.UUID, request TravellerRequest) (Traveller, error) {
	traveller, err := validateTraveller(request)
	if err != nil {
		return Traveller{}, err
	}
	traveller.ID = travellerID
	traveller, err = s.repo.UpdateTraveller(ctx, accountID, traveller)
	if errors.Is(err, pgx.ErrNoRows) {
		return Traveller{}, apperror.New(404, "TRAVELLER_NOT_FOUND", "Saved traveller was not found.", nil)
	}
	if err != nil {
		return Traveller{}, apperror.Wrap(err)
	}
	return traveller, nil
}

func (s *Service) DeleteTraveller(ctx context.Context, accountID, travellerID uuid.UUID) error {
	err := s.repo.DeleteTraveller(ctx, accountID, travellerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.New(404, "TRAVELLER_NOT_FOUND", "Saved traveller was not found.", nil)
	}
	if err != nil {
		return apperror.Wrap(err)
	}
	return nil
}

func (s *Service) GetPreferences(ctx context.Context, accountID uuid.UUID) (Preferences, error) {
	preferences, err := s.repo.GetPreferences(ctx, accountID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Preferences{PreferredCoachClass: "ANY", PreferredSeatType: "ANY", Language: "en"}, nil
	}
	if err != nil {
		return Preferences{}, apperror.Wrap(err)
	}
	return preferences, nil
}

func (s *Service) UpdatePreferences(ctx context.Context, accountID uuid.UUID, request PreferencesRequest) (Preferences, error) {
	preferences := Preferences{
		PreferredCoachClass: strings.ToUpper(strings.TrimSpace(request.PreferredCoachClass)),
		PreferredSeatType:   strings.ToUpper(strings.TrimSpace(request.PreferredSeatType)),
		Language:            strings.ToLower(strings.TrimSpace(request.Language)),
	}
	if !oneOf(preferences.PreferredCoachClass, "ANY", "FIRST", "SECOND") {
		return Preferences{}, apperror.Validation("preferredCoachClass", "Choose ANY, FIRST, or SECOND.")
	}
	if !oneOf(preferences.PreferredSeatType, "ANY", "WINDOW", "AISLE") {
		return Preferences{}, apperror.Validation("preferredSeatType", "Choose ANY, WINDOW, or AISLE.")
	}
	if !oneOf(preferences.Language, "en", "si", "ta") {
		return Preferences{}, apperror.Validation("language", "Choose en, si, or ta.")
	}
	preferences, err := s.repo.UpsertPreferences(ctx, accountID, preferences)
	if err != nil {
		return Preferences{}, apperror.Wrap(err)
	}
	return preferences, nil
}

func validateTraveller(request TravellerRequest) (Traveller, error) {
	request.FullName = strings.TrimSpace(request.FullName)
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Phone = strings.TrimSpace(request.Phone)
	if len(request.FullName) < 2 || len(request.FullName) > 120 {
		return Traveller{}, apperror.Validation("fullName", "Full name must contain 2 to 120 characters.")
	}
	if request.Email != "" {
		address, err := mail.ParseAddress(request.Email)
		if err != nil || address.Address != request.Email || len(request.Email) > 254 {
			return Traveller{}, apperror.Validation("email", "Enter a valid email address.")
		}
	}
	if request.Phone != "" && !phonePattern.MatchString(request.Phone) {
		return Traveller{}, apperror.Validation("phone", "Use international phone format, for example +94770000000.")
	}
	if request.Email == "" && request.Phone == "" {
		return Traveller{}, apperror.Validation("email", "Enter an email address or phone number.")
	}
	return Traveller{FullName: request.FullName, Email: request.Email, Phone: request.Phone}, nil
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func (s *Service) newSession(ctx context.Context, accountID uuid.UUID) (Session, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return Session{}, apperror.Wrap(err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	expiresAt := time.Now().Add(s.sessionTTL)
	if err := s.repo.InsertSession(ctx, uuid.New(), accountID, hashToken(token), expiresAt); err != nil {
		return Session{}, apperror.Wrap(err)
	}
	return Session{Token: token, ExpiresAt: expiresAt}, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func validatePassword(password string) error {
	if len(password) < 10 || len(password) > 72 {
		return apperror.Validation("password", "Password must contain 10 to 72 characters.")
	}
	var upper, lower, digit bool
	for _, char := range password {
		upper = upper || char >= 'A' && char <= 'Z'
		lower = lower || char >= 'a' && char <= 'z'
		digit = digit || char >= '0' && char <= '9'
	}
	if !upper || !lower || !digit {
		return apperror.Validation("password", "Password must include upper-case, lower-case, and numeric characters.")
	}
	return nil
}
