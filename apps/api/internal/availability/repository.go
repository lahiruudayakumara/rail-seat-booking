package availability

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/database"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/journey"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/seat"
)

type Repository struct{ db database.DBTX }

func NewRepository(db database.DBTX) *Repository { return &Repository{db: db} }
func (r *Repository) List(ctx context.Context, runID uuid.UUID, segment journey.Segment, coachClass string) ([]seat.Seat, error) {
	args := []any{runID, segment.OriginPosition, segment.DestinationPosition}
	filter := ""
	if coachClass != "" {
		args = append(args, coachClass)
		filter = fmt.Sprintf(" AND c.coach_class=$%d", len(args))
	}
	query := `SELECT s.id,s.label,c.id,c.code,c.coach_class,s.attributes FROM train_runs tr JOIN coaches c ON c.train_id=tr.train_id AND c.active AND c.reservation_type='RESERVED' JOIN seats s ON s.coach_id=c.id AND s.active WHERE tr.id=$1` + filter + ` AND NOT EXISTS (SELECT 1 FROM bookings b WHERE b.train_run_id=tr.id AND b.seat_id=s.id AND b.status IN ('HELD','CONFIRMED') AND int4range(b.origin_position,b.destination_position,'[)') && int4range($2,$3,'[)')) ORDER BY c.sequence,s.row_number,s.column_code`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []seat.Seat{}
	for rows.Next() {
		var x seat.Seat
		var attributes []byte
		if err = rows.Scan(&x.ID, &x.Label, &x.CoachID, &x.CoachCode, &x.CoachClass, &attributes); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(attributes, &x.Attributes); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
