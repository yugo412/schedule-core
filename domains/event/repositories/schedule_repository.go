package repositories

import (
	"context"
	"database/sql"
	"time"

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
		Url       sql.NullString `db:"url"`
		Title     string         `db:"title"`
		Slug      string         `db:"slug"`
		StartedAt sql.NullString `db:"started_at"`
	}

	query := `SELECT url, title, slug, started_at FROM schedules where slug = ? LIMIT 1`

	err := r.Db.GetContext(ctx, &row, query, slug)
	if err != nil {
		return nil, err
	}

	return &models.Schedule{
		Slug:      row.Slug,
		Url:       row.Url.String,
		Title:     row.Title,
		StartedAt: parseTime(row.StartedAt),
	}, nil
}

// parseTime reads the UTC datetime that Laravel stores in the schedules table.
func parseTime(value sql.NullString) *time.Time {
	if !value.Valid || value.String == "" {
		return nil
	}

	for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339} {
		parsed, err := time.Parse(layout, value.String)
		if err == nil {
			return &parsed
		}
	}

	return nil
}
