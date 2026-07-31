package trainrun

import (
	"github.com/google/uuid"
	"time"
)

type TrainRun struct {
	ID          uuid.UUID `json:"id"`
	RouteID     uuid.UUID `json:"routeId"`
	TrainID     uuid.UUID `json:"trainId"`
	ServiceDate string    `json:"serviceDate"`
	DepartureAt time.Time `json:"departureAt"`
	ArrivalAt   time.Time `json:"arrivalAt"`
	Status      string    `json:"status"`
}
type Filters struct {
	TravelDate time.Time
	RouteID    *uuid.UUID
}
