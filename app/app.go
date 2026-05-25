package app

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/yugo412/schedule-core/config"
	"github.com/yugo412/schedule-core/integrations/umami"
)

type App struct {
	Config *config.Config
	DB     *sqlx.DB
	Logger *slog.Logger

	Umami *umami.Client
}
