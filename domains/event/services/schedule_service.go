package services

import (
	"context"

	"github.com/yugo412/schedule-core/domains/event/models"
	"github.com/yugo412/schedule-core/domains/event/repositories"
)

type ScheduleService struct {
	repository *repositories.ScheduleRepository
}

func NewScheduleService(repository *repositories.ScheduleRepository) *ScheduleService {
	return &ScheduleService{
		repository: repository,
	}
}

func (s *ScheduleService) FindBySlug(ctx context.Context, slug string) (*models.Schedule, error) {
	return s.repository.FindBySlug(ctx, slug)
}
