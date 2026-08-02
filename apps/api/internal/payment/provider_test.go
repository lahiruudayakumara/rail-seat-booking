package payment

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestSandboxProviderCompletesChargeAndRefund(t *testing.T) {
	provider := SandboxProvider{}
	paymentID := uuid.New()
	charged, err := provider.Charge(context.Background(), paymentID, 69000, "LKR")
	if err != nil || charged.Status != "PAID" || charged.Reference == "" {
		t.Fatalf("unexpected charge result: %#v, %v", charged, err)
	}
	refunded, err := provider.Refund(context.Background(), uuid.New(), charged.Reference, 69000, "LKR")
	if err != nil || refunded.Status != "SUCCEEDED" || refunded.Reference == "" {
		t.Fatalf("unexpected refund result: %#v, %v", refunded, err)
	}
}
