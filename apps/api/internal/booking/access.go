package booking

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AccessSigner struct {
	secret []byte
	ttl    time.Duration
}

func NewAccessSigner(secret string, ttl time.Duration) *AccessSigner {
	return &AccessSigner{secret: []byte(secret), ttl: ttl}
}

func (s *AccessSigner) Sign(bookingID uuid.UUID) string {
	payload := bookingID.String() + "." + strconv.FormatInt(time.Now().Add(s.ttl).Unix(), 10)
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	return payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *AccessSigner) Verify(token string, bookingID uuid.UUID) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != bookingID.String() {
		return errors.New("invalid management token")
	}
	expiresAt, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || time.Now().Unix() >= expiresAt {
		return errors.New("expired management token")
	}
	payload := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	want := mac.Sum(nil)
	got, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(got, want) {
		return errors.New("invalid management token")
	}
	return nil
}
