package booking

import (
	"regexp"
	"testing"

	"github.com/google/uuid"
)

func TestGroupReferenceIsPublicSafeAndStable(t *testing.T) {
	id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	first := groupReference(id)
	second := groupReference(id)

	if first != second {
		t.Fatalf("group reference changed for the same id: %q != %q", first, second)
	}
	if !regexp.MustCompile(`^GR-[A-Z2-7]{12}$`).MatchString(first) {
		t.Fatalf("unexpected group reference format: %q", first)
	}
}
