package fare

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/database"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/journey"
	"time"
)

type Repository struct{ db database.DBTX }

func NewRepository(db database.DBTX) *Repository { return &Repository{db: db} }
func (r *Repository) ActiveRule(ctx context.Context, runID, seatID uuid.UUID) (Rule, error) {
	var x Rule
	err := r.db.QueryRow(ctx, `SELECT fr.id,fr.base_fee_minor,fr.rate_per_km_minor,fr.minimum_fare_minor,fr.class_multiplier_basis_points,fr.currency,fr.currency_scale FROM train_runs tr JOIN coaches c ON c.train_id=tr.train_id AND c.active AND c.reservation_type='RESERVED' JOIN seats s ON s.coach_id=c.id AND s.active JOIN fare_rules fr ON fr.route_id=tr.route_id AND fr.coach_class=c.coach_class AND fr.active AND fr.effective_from<=tr.service_date AND (fr.effective_to IS NULL OR fr.effective_to>tr.service_date) WHERE tr.id=$1 AND s.id=$2 ORDER BY fr.effective_from DESC LIMIT 1`, runID, seatID).Scan(&x.ID, &x.BaseFeeMinor, &x.RatePerKMMinor, &x.MinimumFareMinor, &x.MultiplierBasisPoints, &x.Currency, &x.CurrencyScale)
	return x, err
}
func (r *Repository) InsertQuote(ctx context.Context, id uuid.UUID, request QuoteRequest, segment journey.Segment, rule Rule, amount int64, breakdown map[string]int64, expiresAt time.Time) error {
	raw, err := json.Marshal(breakdown)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `INSERT INTO fare_quotes(id,train_run_id,seat_id,origin_station_id,destination_station_id,origin_position,destination_position,distance_m,fare_rule_id,amount_minor,currency,currency_scale,breakdown,expires_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, id, request.TrainRunID, request.SeatID, request.OriginStationID, request.DestinationStationID, segment.OriginPosition, segment.DestinationPosition, segment.DistanceM, rule.ID, amount, rule.Currency, rule.CurrencyScale, raw, expiresAt)
	return err
}
