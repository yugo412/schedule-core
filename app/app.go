package app

import "github.com/jmoiron/sqlx"

type App struct {
	DB *sqlx.DB
}
