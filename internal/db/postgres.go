package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/natewong1313/spy/internal/errors"
)

func New(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, errors.Wrap(err, "open db")
	}
	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "ping db")
	}
	return db, nil
}
