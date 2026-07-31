package seat

import "github.com/google/uuid"

type Seat struct {
	ID         uuid.UUID `json:"id"`
	Label      string    `json:"label"`
	CoachID    uuid.UUID `json:"coachId"`
	CoachCode  string    `json:"coachCode"`
	CoachClass string    `json:"coachClass"`
	Attributes []string  `json:"attributes"`
}
