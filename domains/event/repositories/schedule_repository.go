package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/yugo412/schedule-core/domains/event/models"
)

type ScheduleRepository struct {
	Db *sqlx.DB
}

func NewScheduleRepository(db *sqlx.DB) *ScheduleRepository {
	return &ScheduleRepository{
		Db: db,
	}
}

func (r *ScheduleRepository) FindBySlug(ctx context.Context, slug string) (*models.Schedule, error) {
	schedule := &models.Schedule{}

	query := `SELECT url, title, slug FROM schedules where slug = ? LIMIT 1`

	err := r.Db.GetContext(ctx, schedule, query, slug)
	if err != nil {
		return nil, err
	}

	return schedule, nil
}
