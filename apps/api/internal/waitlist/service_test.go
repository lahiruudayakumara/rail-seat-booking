package waitlist

import (
	"regexp"
	"testing"

	"github.com/google/uuid"
)

func TestWaitlistReference(t *testing.T) {
	reference := waitlistReference(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"))
	if !regexp.MustCompile(`^WL-[A-Z2-7]{12}$`).MatchString(reference) {
		t.Fatalf("unexpected reference %q", reference)
	}
}

func TestValidateCreateNormalizesContact(t *testing.T) {
	entry, err := validateCreate(CreateRequest{FullName: " Passenger Name ", Email: " Passenger@Example.COM ", Phone: "077 000 0123", PreferredCoachClass: "first"})
	if err != nil {
		t.Fatal(err)
	}
	if entry.Email != "passenger@example.com" || entry.Phone != "+94770000123" || entry.PreferredCoachClass != "FIRST" {
		t.Fatalf("unexpected normalized entry: %#v", entry)
	}
}
