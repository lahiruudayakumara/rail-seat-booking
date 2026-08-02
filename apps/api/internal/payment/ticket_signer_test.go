package payment

import (
	"testing"

	"github.com/google/uuid"
)

func TestTicketSignerReturnsStableTicketSpecificCodes(t *testing.T) {
	signer := NewTicketSigner("test-secret-with-sufficient-entropy")
	firstID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	secondID := uuid.MustParse("10000000-0000-4000-8000-000000000002")

	first := signer.Code(firstID)
	if repeated := signer.Code(firstID); repeated != first {
		t.Fatalf("expected stable code, got %q then %q", first, repeated)
	}
	if second := signer.Code(secondID); second == first {
		t.Fatal("expected different tickets to have different codes")
	}
}

func TestTicketSignerSecretsAreIsolated(t *testing.T) {
	ticketID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	first := NewTicketSigner("first-test-secret").Code(ticketID)
	second := NewTicketSigner("second-test-secret").Code(ticketID)
	if first == second {
		t.Fatal("expected different secrets to derive different codes")
	}
}
