package payment

import (
	"context"

	"github.com/google/uuid"
)

type ProviderResult struct {
	Reference string
	Status    string
}

type Provider interface {
	Charge(context.Context, uuid.UUID, int64, string) (ProviderResult, error)
}

type SandboxProvider struct{}

func (SandboxProvider) Charge(_ context.Context, paymentID uuid.UUID, _ int64, _ string) (ProviderResult, error) {
	return ProviderResult{Reference: "sandbox-" + paymentID.String(), Status: "PAID"}, nil
}
