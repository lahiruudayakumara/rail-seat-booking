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

type VerifyTicketRequest struct {
	VerificationCode string `json:"verificationCode"`
}

type TicketVerification struct {
	Valid                bool      `json:"valid"`
	TicketID             uuid.UUID `json:"ticketId"`
	TicketStatus         string    `json:"ticketStatus"`
	BookingReference     string    `json:"bookingReference"`
	BookingStatus        string    `json:"bookingStatus"`
	TrainRunID           uuid.UUID `json:"trainRunId"`
	CoachCode            string    `json:"coachCode"`
	SeatLabel            string    `json:"seatLabel"`
	OriginStationID      uuid.UUID `json:"originStationId"`
	DestinationStationID uuid.UUID `json:"destinationStationId"`
	VerifiedAt           time.Time `json:"verifiedAt"`
}

type checkoutRecord struct {
	BookingID uuid.UUID
	Payment   Payment
	TicketID  uuid.UUID
}
