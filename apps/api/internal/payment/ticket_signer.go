package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"github.com/google/uuid"
)

// TicketSigner derives a stable, unguessable verification code from a ticket ID.
// Stable codes let an idempotent checkout retry return the original credential
// without storing the credential itself in the database.
type TicketSigner struct {
	secret []byte
}

func NewTicketSigner(secret string) *TicketSigner {
	return &TicketSigner{secret: []byte(secret)}
}

func (s *TicketSigner) Code(ticketID uuid.UUID) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte("ticket:" + ticketID.String()))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
