package main

import (
	"log"
	"net/http"

	"github.com/yourusername/schedule-core/config"
	"github.com/yourusername/schedule-core/database"
	"github.com/yourusername/schedule-core/router"
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

	r := router.New(db)

	log.Println("Server running on: " + cfg.AppPort)

	http.ListenAndServe(":"+cfg.AppPort, r)
}
