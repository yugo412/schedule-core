package models

type Schedule struct {
	ID   int    `db:"id"`
	Slug string `db:"slug"`
	Url  string `db:"url"`
}
