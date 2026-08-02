package availability

import "github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/seat"

type SeatStatus string

const (
	SeatAvailable SeatStatus = "AVAILABLE"
	SeatBooked    SeatStatus = "BOOKED"
)

type SeatMapItem struct {
	seat.Seat
	AvailabilityStatus SeatStatus `json:"availabilityStatus"`
}
