package admin

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) Dashboard(ctx context.Context, trainRunID uuid.UUID) (Dashboard, error) {
	item, err := s.repo.Dashboard(ctx, trainRunID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Dashboard{}, apperror.New(404, "TRAIN_RUN_NOT_FOUND", "Train run was not found.", nil)
	}
	if err != nil {
		return Dashboard{}, apperror.Wrap(err)
	}
	return item, nil
}
