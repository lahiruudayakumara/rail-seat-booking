package payment

import (
	"time"

	"github.com/google/uuid"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/booking"
)

type CheckoutRequest struct {
	BookingID    uuid.UUID `json:"bookingId"`
	BookingToken string    `json:"bookingToken"`
}

type Payment struct {
	ID                uuid.UUID `json:"id"`
	Status            string    `json:"status"`
	Provider          string    `json:"provider"`
	ProviderReference string    `json:"providerReference"`
	AmountMinor       int64     `json:"amountMinor"`
	Currency          string    `json:"currency"`
	PaidAt            time.Time `json:"paidAt"`
}

type Ticket struct {
	ID               uuid.UUID `json:"id"`
	VerificationCode string    `json:"verificationCode"`
	Status           string    `json:"status"`
}

type CheckoutResult struct {
	Payment Payment         `json:"payment"`
	Ticket  Ticket          `json:"ticket"`
	Booking booking.Booking `json:"booking"`
}

type checkoutRecord struct {
	BookingID uuid.UUID
	Payment   Payment
	TicketID  uuid.UUID
}
