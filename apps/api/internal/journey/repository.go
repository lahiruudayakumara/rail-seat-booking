package journey

import (
	"context"
	"github.com/google/uuid"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/database"
)

type Repository struct{ db database.DBTX }

func NewRepository(db database.DBTX) *Repository { return &Repository{db: db} }
func (r *Repository) Resolve(ctx context.Context, runID, originID, destinationID uuid.UUID) (Segment, error) {
	var x Segment
	err := r.db.QueryRow(ctx, `SELECT tr.route_id,tr.service_date,o.position,d.position,d.cumulative_distance_m-o.cumulative_distance_m FROM train_runs tr JOIN route_stations o ON o.route_id=tr.route_id AND o.station_id=$2 JOIN route_stations d ON d.route_id=tr.route_id AND d.station_id=$3 WHERE tr.id=$1 AND tr.status='SCHEDULED' AND tr.departure_at>now()`, runID, originID, destinationID).Scan(&x.RouteID, &x.ServiceDate, &x.OriginPosition, &x.DestinationPosition, &x.DistanceM)
	return x, err
}
