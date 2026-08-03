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

type GroupCheckoutRequest struct {
	GroupID    uuid.UUID `json:"groupId"`
	GroupToken string    `json:"groupToken"`
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

type GroupCheckoutResult struct {
	Payment Payment              `json:"payment"`
	Tickets []Ticket             `json:"tickets"`
	Group   booking.BookingGroup `json:"group"`
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

type PayHereCheckoutRequest struct {
	BookingID      uuid.UUID `json:"bookingId"`
	BookingToken   string    `json:"bookingToken"`
	BillingAddress string    `json:"billingAddress"`
	City           string    `json:"city"`
}

type PayHereGroupCheckoutRequest struct {
	GroupID        uuid.UUID `json:"groupId"`
	GroupToken     string    `json:"groupToken"`
	BillingAddress string    `json:"billingAddress"`
	City           string    `json:"city"`
}

type PayHereCheckoutResponse struct {
	PaymentID uuid.UUID         `json:"paymentId"`
	BookingID uuid.UUID         `json:"bookingId"`
	GroupID   *uuid.UUID        `json:"groupId,omitempty"`
	Status    string            `json:"status"`
	ExpiresAt time.Time         `json:"expiresAt"`
	ActionURL string            `json:"actionUrl"`
	Fields    map[string]string `json:"fields"`
}

type PayHerePaymentStatus struct {
	PaymentID         uuid.UUID             `json:"paymentId"`
	BookingID         uuid.UUID             `json:"bookingId"`
	Status            string                `json:"status"`
	ProviderReference string                `json:"providerReference"`
	AmountMinor       int64                 `json:"amountMinor"`
	Currency          string                `json:"currency"`
	PaidAt            *time.Time            `json:"paidAt,omitempty"`
	Booking           *booking.Booking      `json:"booking,omitempty"`
	Ticket            *Ticket               `json:"ticket,omitempty"`
	Group             *booking.BookingGroup `json:"group,omitempty"`
	Tickets           []Ticket              `json:"tickets,omitempty"`
}

type payHereCheckoutRecord struct {
	PaymentID     uuid.UUID
	BookingID     uuid.UUID
	GroupID       *uuid.UUID
	Status        string
	BookingStatus string
	AmountMinor   int64
	Currency      string
	ExpiresAt     time.Time
	FullName      string
	Email         string
	Phone         string
}

type payHerePaymentRecord struct {
	PaymentID         uuid.UUID
	BookingID         uuid.UUID
	GroupID           *uuid.UUID
	Status            string
	ProviderReference string
	AmountMinor       int64
	Currency          string
	PaidAt            *time.Time
	TicketID          *uuid.UUID
	TicketStatus      *string
}

type groupCheckoutRecord struct {
	GroupID uuid.UUID
	Payment Payment
}

type ticketIssue struct {
	ID        uuid.UUID
	BookingID uuid.UUID
	CodeHash  string
}
