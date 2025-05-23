package queries

import "database/sql"

type QueryEngine struct {
	db *sql.DB
}

func New(db *sql.DB) *QueryEngine {
	return &QueryEngine{
		db: db,
	}
}
