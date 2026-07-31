package fare

import (
	"github.com/google/uuid"
	"time"
)

type Money struct {
	AmountMinor   int64  `json:"amountMinor"`
	Currency      string `json:"currency"`
	CurrencyScale int16  `json:"currencyScale"`
}
type QuoteRequest struct {
	TrainRunID           uuid.UUID `json:"trainRunId"`
	SeatID               uuid.UUID `json:"seatId"`
	OriginStationID      uuid.UUID `json:"originStationId"`
	DestinationStationID uuid.UUID `json:"destinationStationId"`
}
type Quote struct {
	ID            uuid.UUID        `json:"id"`
	DistanceKM    string           `json:"distanceKm"`
	AmountMinor   int64            `json:"amountMinor"`
	Currency      string           `json:"currency"`
	CurrencyScale int16            `json:"currencyScale"`
	Breakdown     map[string]int64 `json:"breakdown"`
	ExpiresAt     time.Time        `json:"expiresAt"`
}
type Rule struct {
	ID                    uuid.UUID
	BaseFeeMinor          int64
	RatePerKMMinor        int64
	MinimumFareMinor      int64
	MultiplierBasisPoints int32
	Currency              string
	CurrencyScale         int16
}
