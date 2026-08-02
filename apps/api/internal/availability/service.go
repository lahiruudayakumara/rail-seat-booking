package availability

import (
	"context"
	"github.com/google/uuid"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/journey"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/seat"
)

type Service struct {
	repo     *Repository
	journeys *journey.Service
}

func NewService(repo *Repository, journeys *journey.Service) *Service {
	return &Service{repo: repo, journeys: journeys}
}
func (s *Service) List(ctx context.Context, runID, originID, destinationID uuid.UUID, coachClass string) ([]seat.Seat, error) {
	segment, err := s.journeys.Resolve(ctx, runID, originID, destinationID)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.List(ctx, runID, segment, coachClass)
	if err != nil {
		return nil, apperror.Wrap(err)
	}
	return items, nil
}

func (s *Service) SeatMap(ctx context.Context, runID, originID, destinationID uuid.UUID, coachClass string) ([]SeatMapItem, error) {
	segment, err := s.journeys.Resolve(ctx, runID, originID, destinationID)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.SeatMap(ctx, runID, segment, coachClass)
	if err != nil {
		return nil, apperror.Wrap(err)
	}
	return items, nil
}
