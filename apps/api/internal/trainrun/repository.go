package trainrun

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/database"
	"time"
)

type Repository struct{ db database.DBTX }

func NewRepository(db database.DBTX) *Repository { return &Repository{db: db} }
func (r *Repository) List(ctx context.Context, filters Filters) ([]TrainRun, error) {
	args := []any{filters.TravelDate}
	query := `SELECT id,train_id,route_id,service_date,departure_at,arrival_at,status FROM train_runs WHERE service_date=$1 AND status='SCHEDULED'`
	if filters.RouteID != nil {
		args = append(args, *filters.RouteID)
		query += fmt.Sprintf(" AND route_id=$%d", len(args))
	}
	query += " ORDER BY departure_at"
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TrainRun{}
	for rows.Next() {
		x, err := scan(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
func (r *Repository) Get(ctx context.Context, id uuid.UUID) (TrainRun, error) {
	return scan(func(dest ...any) error {
		return r.db.QueryRow(ctx, `SELECT id,train_id,route_id,service_date,departure_at,arrival_at,status FROM train_runs WHERE id=$1`, id).Scan(dest...)
	})
}
func scan(scanner func(...any) error) (TrainRun, error) {
	var x TrainRun
	var date time.Time
	err := scanner(&x.ID, &x.TrainID, &x.RouteID, &date, &x.DepartureAt, &x.ArrivalAt, &x.Status)
	x.ServiceDate = date.Format("2006-01-02")
	return x, err
}
