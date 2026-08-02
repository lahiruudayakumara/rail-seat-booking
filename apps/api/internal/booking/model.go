package booking

import (
	"github.com/google/uuid"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/fare"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/seat"
	"time"
)

type PassengerInput struct {
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}
type CreateRequest struct {
	FareQuoteID          uuid.UUID      `json:"fareQuoteId"`
	TrainRunID           uuid.UUID      `json:"trainRunId"`
	SeatID               uuid.UUID      `json:"seatId"`
	OriginStationID      uuid.UUID      `json:"originStationId"`
	DestinationStationID uuid.UUID      `json:"destinationStationId"`
	Passenger            PassengerInput `json:"passenger"`
}
type AccessRequest struct {
	Reference string `json:"reference"`
	Contact   string `json:"contact"`
}
type Booking struct {
	ID                   uuid.UUID  `json:"id"`
	Reference            string     `json:"reference"`
	Status               string     `json:"status"`
	TrainRunID           uuid.UUID  `json:"trainRunId"`
	Seat                 seat.Seat  `json:"seat"`
	OriginStationID      uuid.UUID  `json:"originStationId"`
	DestinationStationID uuid.UUID  `json:"destinationStationId"`
	Fare                 fare.Money `json:"fare"`
	CreatedAt            time.Time  `json:"createdAt"`
	ConfirmedAt          *time.Time `json:"confirmedAt,omitempty"`
	CancelledAt          *time.Time `json:"cancelledAt,omitempty"`
	ManagementToken      string     `json:"managementToken,omitempty"`
}
type QuoteSnapshot struct {
	TrainRunID           uuid.UUID
	SeatID               uuid.UUID
	OriginStationID      uuid.UUID
	DestinationStationID uuid.UUID
	FareRuleID           uuid.UUID
	OriginPosition       int32
	DestinationPosition  int32
	AmountMinor          int64
	Currency             string
	CurrencyScale        int16
	Breakdown            []byte
}
