package database

import (
	"github.com/jmoiron/sqlx"

	_ "modernc.org/sqlite"
)

func NewSqlite(path string) (*sqlx.DB, error) {
	return sqlx.Connect("sqlite", path)
}
