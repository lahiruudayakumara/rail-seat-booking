package waitlist

import (
	"time"

	"github.com/google/uuid"
)

type CreateRequest struct {
	TrainRunID           uuid.UUID `json:"trainRunId"`
	OriginStationID      uuid.UUID `json:"originStationId"`
	DestinationStationID uuid.UUID `json:"destinationStationId"`
	FullName             string    `json:"fullName"`
	Email                string    `json:"email"`
	Phone                string    `json:"phone"`
	PreferredCoachClass  string    `json:"preferredCoachClass"`
}

type AccessRequest struct {
	Reference string `json:"reference"`
	Contact   string `json:"contact"`
}

type Entry struct {
	ID                   uuid.UUID  `json:"id"`
	Reference            string     `json:"reference"`
	TrainRunID           uuid.UUID  `json:"trainRunId"`
	OriginStationID      uuid.UUID  `json:"originStationId"`
	DestinationStationID uuid.UUID  `json:"destinationStationId"`
	OriginPosition       int32      `json:"-"`
	DestinationPosition  int32      `json:"-"`
	FullName             string     `json:"fullName"`
	Email                string     `json:"email,omitempty"`
	Phone                string     `json:"phone,omitempty"`
	PreferredCoachClass  string     `json:"preferredCoachClass"`
	Status               string     `json:"status"`
	CreatedAt            time.Time  `json:"createdAt"`
	NotifiedAt           *time.Time `json:"notifiedAt,omitempty"`
	ManagementToken      string     `json:"managementToken,omitempty"`
}
