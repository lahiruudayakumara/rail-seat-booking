package station

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/database"
)

type Repository struct{ db database.DBTX }

func NewRepository(db database.DBTX) *Repository { return &Repository{db: db} }
func (r *Repository) List(ctx context.Context) ([]Station, error) {
	rows, err := r.db.Query(ctx, `SELECT id,code,name FROM stations WHERE active ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Station{}
	for rows.Next() {
		var x Station
		if err = rows.Scan(&x.ID, &x.Code, &x.Name); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
func (r *Repository) ListByRoute(ctx context.Context, routeID uuid.UUID) ([]Station, error) {
	rows, err := r.db.Query(ctx, `SELECT s.id,s.code,s.name,rs.position,rs.cumulative_distance_m FROM route_stations rs JOIN stations s ON s.id=rs.station_id WHERE rs.route_id=$1 AND s.active ORDER BY rs.position`, routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Station{}
	for rows.Next() {
		var x Station
		var position, distance int32
		if err = rows.Scan(&x.ID, &x.Code, &x.Name, &position, &distance); err != nil {
			return nil, err
		}
		km := fmt.Sprintf("%.3f", float64(distance)/1000)
		x.Position = &position
		x.CumulativeDistanceKM = &km
		items = append(items, x)
	}
	return items, rows.Err()
}
