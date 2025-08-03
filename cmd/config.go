package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateDatabaseConnection(url string) *pgxpool.Pool {
	conn, err := pgxpool.New(context.Background(), url)
	if err != nil {
		slog.Error("database_connection_failed",
			"error", err.Error(),
			"url", url,
		)
		os.Exit(1)
	}

	return conn
}
