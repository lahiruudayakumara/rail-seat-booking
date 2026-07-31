package fare

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/journey"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
	"time"
)

type Service struct {
	repo     *Repository
	journeys *journey.Service
	quoteTTL time.Duration
}

func NewService(repo *Repository, journeys *journey.Service, quoteTTL time.Duration) *Service {
	return &Service{repo: repo, journeys: journeys, quoteTTL: quoteTTL}
}
func (s *Service) CreateQuote(ctx context.Context, request QuoteRequest) (Quote, error) {
	segment, err := s.journeys.Resolve(ctx, request.TrainRunID, request.OriginStationID, request.DestinationStationID)
	if err != nil {
		return Quote{}, err
	}
	rule, err := s.repo.ActiveRule(ctx, request.TrainRunID, request.SeatID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Quote{}, apperror.New(422, "SEAT_NOT_RESERVABLE", "Seat or fare rule is not available.", nil)
	}
	if err != nil {
		return Quote{}, apperror.Wrap(err)
	}
	distanceFee, total := Calculate(segment.DistanceM, rule.BaseFeeMinor, rule.RatePerKMMinor, rule.MinimumFareMinor, rule.MultiplierBasisPoints)
	breakdown := map[string]int64{"baseFeeMinor": rule.BaseFeeMinor, "distanceFeeMinor": distanceFee, "classMultiplierBasisPoints": int64(rule.MultiplierBasisPoints)}
	id := uuid.New()
	expiresAt := time.Now().UTC().Add(s.quoteTTL)
	if err = s.repo.InsertQuote(ctx, id, request, segment, rule, total, breakdown, expiresAt); err != nil {
		return Quote{}, apperror.Wrap(err)
	}
	return Quote{ID: id, DistanceKM: fmt.Sprintf("%.3f", float64(segment.DistanceM)/1000), AmountMinor: total, Currency: rule.Currency, CurrencyScale: rule.CurrencyScale, Breakdown: breakdown, ExpiresAt: expiresAt}, nil
}
func Calculate(distanceM int32, base, ratePerKM, minimum int64, multiplierBasisPoints int32) (int64, int64) {
	distanceFee := (int64(distanceM)*ratePerKM + 999) / 1000
	amount := ((base+distanceFee)*int64(multiplierBasisPoints) + 9999) / 10000
	if amount < minimum {
		amount = minimum
	}
	return distanceFee, amount
}
