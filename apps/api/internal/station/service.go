package station

import (
	"context"
	"github.com/google/uuid"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }
func (s *Service) List(ctx context.Context) ([]Station, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Wrap(err)
	}
	return items, nil
}
func (s *Service) ListByRoute(ctx context.Context, id uuid.UUID) ([]Station, error) {
	items, err := s.repo.ListByRoute(ctx, id)
	if err != nil {
		return nil, apperror.Wrap(err)
	}
	if len(items) == 0 {
		return nil, apperror.New(404, "ROUTE_NOT_FOUND", "Route was not found.", nil)
	}
	return items, nil
}
