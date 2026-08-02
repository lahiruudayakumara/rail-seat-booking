package passengerauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
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
	if !strings.Contains(request.Email, "@") || len(request.Email) > 254 {
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
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && bcrypt.CompareHashAndPassword([]byte(hash), []byte(request.Password)) != nil) {
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
