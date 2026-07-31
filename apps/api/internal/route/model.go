package route

import "github.com/google/uuid"

type Route struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Direction string    `json:"direction"`
	Timezone  string    `json:"timezone"`
}
