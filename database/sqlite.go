package database

import (
	"strings"

	"github.com/vinovest/sqlx"

	_ "modernc.org/sqlite"
)

func NewSqlite(path string) (*sqlx.DB, error) {
	return sqlx.Connect("sqlite", dsn(path))
}

func dsn(path string) string {
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}

	return path +
		separator +
		"_pragma=busy_timeout(5000)" +
		"&_pragma=journal_mode(WAL)"
}
