package passengerauth

import "testing"

func TestValidatePassword(t *testing.T) {
	for _, tc := range []struct {
		name     string
		password string
		valid    bool
	}{
		{name: "strong", password: "ScenicRail2026", valid: true},
		{name: "too short", password: "Rail1", valid: false},
		{name: "missing uppercase", password: "scenicrail2026", valid: false},
		{name: "missing digit", password: "ScenicRailway", valid: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := validatePassword(tc.password) == nil; got != tc.valid {
				t.Fatalf("valid=%v; want %v", got, tc.valid)
			}
		})
	}
}

func TestHashTokenIsStableAndDoesNotExposeToken(t *testing.T) {
	token := "passenger-session-token"
	if hashToken(token) != hashToken(token) {
		t.Fatal("expected stable hash")
	}
	if hashToken(token) == token {
		t.Fatal("token must not be stored verbatim")
	}
}
