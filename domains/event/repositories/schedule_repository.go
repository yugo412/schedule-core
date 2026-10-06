package repositories

import (
	"context"
	"database/sql"

	"github.com/vinovest/sqlx"
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
	var row struct {
		Url   sql.NullString `db:"url"`
		Title string         `db:"title"`
		Slug  string         `db:"slug"`
	}

	query := `SELECT url, title, slug FROM schedules where slug = ? LIMIT 1`

	err := r.Db.GetContext(ctx, &row, query, slug)
	if err != nil {
		return nil, err
	}

	return &models.Schedule{
		Slug:  row.Slug,
		Url:   row.Url.String,
		Title: row.Title,
	}, nil
}
