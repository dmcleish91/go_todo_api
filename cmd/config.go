package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DemoUserID string

func loadConfig() {
	DemoUserID = os.Getenv("DEMO_USER_ID")
	if DemoUserID == "" {
		log.Fatal("DEMO_USER_ID environment variable is required")
	}
}

func CreateDatabaseConnection(url string) *pgxpool.Pool {
	conn, err := pgxpool.New(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	return conn
}