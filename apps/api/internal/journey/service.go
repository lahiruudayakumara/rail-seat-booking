package journey

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }
func (s *Service) Resolve(ctx context.Context, runID, originID, destinationID uuid.UUID) (Segment, error) {
	segment, err := s.repo.Resolve(ctx, runID, originID, destinationID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && (segment.OriginPosition >= segment.DestinationPosition || segment.DistanceM <= 0)) {
		return Segment{}, apperror.New(422, "INVALID_JOURNEY_SEGMENT", "Origin and destination must be ordered stations on this route.", nil)
	}
	if err != nil {
		return Segment{}, apperror.Wrap(err)
	}
	return segment, nil
}
