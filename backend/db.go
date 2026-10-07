package main

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func connectionDB() (*pgxpool.Pool, error) {
	databaseURL := os.Getenv("DATABASE_URL")

	ctx := context.Background()

	db, err := pgxpool.New(ctx, databaseURL)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
