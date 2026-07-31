package trainrun

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }
func (s *Service) List(ctx context.Context, filters Filters) ([]TrainRun, error) {
	items, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, apperror.Wrap(err)
	}
	return items, nil
}
func (s *Service) Get(ctx context.Context, id uuid.UUID) (TrainRun, error) {
	item, err := s.repo.Get(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return TrainRun{}, apperror.New(404, "TRAIN_RUN_NOT_FOUND", "Train run was not found.", nil)
	}
	if err != nil {
		return TrainRun{}, apperror.Wrap(err)
	}
	return item, nil
}
