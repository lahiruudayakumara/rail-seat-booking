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

func TestValidateTraveller(t *testing.T) {
	for _, test := range []struct {
		name    string
		request TravellerRequest
		valid   bool
	}{
		{name: "email contact", request: TravellerRequest{FullName: "Nimali Perera", Email: "NIMALI@example.com"}, valid: true},
		{name: "phone contact", request: TravellerRequest{FullName: "Nimali Perera", Phone: "+94770000000"}, valid: true},
		{name: "missing contact", request: TravellerRequest{FullName: "Nimali Perera"}, valid: false},
		{name: "invalid phone", request: TravellerRequest{FullName: "Nimali Perera", Phone: "0770000000"}, valid: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			traveller, err := validateTraveller(test.request)
			if (err == nil) != test.valid {
				t.Fatalf("valid=%v; want %v; error=%v", err == nil, test.valid, err)
			}
			if test.valid && test.request.Email != "" && traveller.Email != "nimali@example.com" {
				t.Fatalf("email=%q; want normalized email", traveller.Email)
			}
		})
	}
}
