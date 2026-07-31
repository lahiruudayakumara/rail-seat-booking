package route

import (
	"context"
	"github.com/google/uuid"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/database"
)

type Repository struct{ db database.DBTX }

func NewRepository(db database.DBTX) *Repository { return &Repository{db: db} }
func (r *Repository) List(ctx context.Context) ([]Route, error) {
	rows, err := r.db.Query(ctx, `SELECT id,code,name,direction,timezone FROM routes WHERE active ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Route{}
	for rows.Next() {
		var x Route
		if err = rows.Scan(&x.ID, &x.Code, &x.Name, &x.Direction, &x.Timezone); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Route, error) {
	var x Route
	err := r.db.QueryRow(ctx, `SELECT id,code,name,direction,timezone FROM routes WHERE id=$1 AND active`, id).Scan(&x.ID, &x.Code, &x.Name, &x.Direction, &x.Timezone)
	return x, err
}
