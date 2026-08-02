package passengerauth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Account struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"fullName"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type Session struct {
	Token     string
	ExpiresAt time.Time
}

type accountContextKey struct{}

func WithAccount(ctx context.Context, account Account) context.Context {
	return context.WithValue(ctx, accountContextKey{}, account)
}

func AccountFromContext(ctx context.Context) (Account, bool) {
	account, ok := ctx.Value(accountContextKey{}).(Account)
	return account, ok
}

func AccountID(ctx context.Context) *uuid.UUID {
	account, ok := AccountFromContext(ctx)
	if !ok {
		return nil
	}
	return &account.ID
}
