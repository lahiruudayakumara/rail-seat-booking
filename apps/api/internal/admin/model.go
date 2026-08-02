package admin

import "github.com/google/uuid"

type Dashboard struct {
	TrainRunID                uuid.UUID `json:"trainRunId"`
	SellableSeatSegments      int64     `json:"sellableSeatSegments"`
	OccupiedSeatSegments      int64     `json:"occupiedSeatSegments"`
	SegmentUtilizationPercent float64   `json:"segmentUtilizationPercent"`
	ConfirmedBookings         int64     `json:"confirmedBookings"`
	CancelledBookings         int64     `json:"cancelledBookings"`
	HeldBookings              int64     `json:"heldBookings"`
	GrossRevenueMinor         int64     `json:"grossRevenueMinor"`
	RefundedMinor             int64     `json:"refundedMinor"`
	NetRevenueMinor           int64     `json:"netRevenueMinor"`
	Currency                  string    `json:"currency"`
	PendingDeliveries         int64     `json:"pendingDeliveries"`
}
