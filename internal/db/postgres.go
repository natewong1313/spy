package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/natewong1313/spy/internal/errors"
)

func New(connString string) (*pgx.Conn, error) {
	ctx := context.Background()
	db, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, errors.Wrap(err, "open db")
	}
	if err := db.Ping(ctx); err != nil {
		return nil, errors.Wrap(err, "ping db")
	}
	return db, nil
}

func NewPool(connString string) (*pgxpool.Pool, error) {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, errors.Wrap(err, "open db")
	}
	if err := db.Ping(ctx); err != nil {
		return nil, errors.Wrap(err, "ping db")
	}
	return db, nil
}
