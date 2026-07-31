package journey

import (
	"github.com/google/uuid"
	"time"
)

// Segment is the resolved, immutable half-open interval [OriginPosition, DestinationPosition).
type Segment struct {
	RouteID             uuid.UUID
	ServiceDate         time.Time
	OriginPosition      int32
	DestinationPosition int32
	DistanceM           int32
}
