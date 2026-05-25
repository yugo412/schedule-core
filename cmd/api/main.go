package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/yugo412/schedule-core/app"
	"github.com/yugo412/schedule-core/config"
	"github.com/yugo412/schedule-core/database"
	"github.com/yugo412/schedule-core/integrations/umami"
	"github.com/yugo412/schedule-core/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewSqlite(cfg.DbPath)
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	umamiClient := umami.NewClient(cfg.UmamiUrl, cfg.UmamiWebsiteId, logger)

	app := &app.App{
		Config: cfg,
		DB:     db,
		Logger: logger,
		Umami:  umamiClient,
	}

	r := router.New(app)

	log.Println("Server running on: " + cfg.AppPort)

	http.ListenAndServe(":"+cfg.AppPort, r)
}
