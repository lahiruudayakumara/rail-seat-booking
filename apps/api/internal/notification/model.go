package notification

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Message struct {
	ID       uuid.UUID
	Topic    string
	Payload  json.RawMessage
	Attempts int
}
