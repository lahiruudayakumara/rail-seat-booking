package station

import "github.com/google/uuid"

type Station struct {
	ID                   uuid.UUID `json:"id"`
	Code                 string    `json:"code"`
	Name                 string    `json:"name"`
	Position             *int32    `json:"position,omitempty"`
	CumulativeDistanceKM *string   `json:"cumulativeDistanceKm,omitempty"`
}
