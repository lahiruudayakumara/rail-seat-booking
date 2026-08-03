package booking

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAccessSigner(t *testing.T) {
	bookingID := uuid.New()
	signer := NewAccessSigner("test-secret-with-at-least-thirty-two-characters", time.Minute)
	token := signer.Sign(bookingID)
	if err := signer.Verify(token, bookingID); err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}
	if err := signer.Verify(token, uuid.New()); err == nil {
		t.Fatal("token accepted for a different booking")
	}
	tampered := token[:len(token)-1] + "x"
	if err := signer.Verify(tampered, bookingID); err == nil {
		t.Fatal("tampered token accepted")
	}
}

func TestAccessSignerRejectsExpiredToken(t *testing.T) {
	bookingID := uuid.New()
	signer := NewAccessSigner("test-secret-with-at-least-thirty-two-characters", -time.Second)
	if err := signer.Verify(signer.Sign(bookingID), bookingID); err == nil {
		t.Fatal("expired token accepted")
	}
}

func TestNormalizeLookupContact(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		" Passenger@Example.COM ": "passenger@example.com",
		"077 000 0123":            "+94770000123",
		"077-000-0123":            "+94770000123",
		"94770000123":             "+94770000123",
		"0094770000123":           "+94770000123",
		"+94 (77) 000 0123":       "+94770000123",
	}
	for input, want := range tests {
		input, want := input, want
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			if got := normalizeLookupContact(input); got != want {
				t.Fatalf("normalizeLookupContact(%q) = %q; want %q", input, got, want)
			}
		})
	}
}
