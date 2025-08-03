package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/dmcleish91/go_todo_api/internal/models"
	"github.com/joho/godotenv"
)

type application struct {
	projects *models.ProjectModel
	tasks    *models.TaskModel
	labels   *models.LabelModel
	logger   *slog.Logger
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		// Don't fail if .env file doesn't exist
		slog.Warn("no .env file found, using system environment variables")
	}

	// Get database configuration
	user := os.Getenv("user")
	password := os.Getenv("password")
	host := os.Getenv("host")
	port := os.Getenv("port")
	dbname := os.Getenv("dbname")

	// Initialize structured JSON logger
	logger := NewStructuredLogger()

	// Log application startup
	logger.Info("application_starting",
		"app_name", "go_todo_api",
		"version", "1.0.0",
		"environment", os.Getenv("ENV"),
		"database_host", host,
		"database_port", port,
		"database_name", dbname,
	)

	DATABASE_URL := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, dbname)

	// Create database connection
	conn := CreateDatabaseConnection(DATABASE_URL)
	defer conn.Close()

	logger.Info("database_connected",
		"database_host", host,
		"database_name", dbname,
	)

	app := &application{
		projects: &models.ProjectModel{DB: conn},
		tasks:    &models.TaskModel{DB: conn},
		labels:   &models.LabelModel{DB: conn},
		logger:   logger,
	}

	e := app.Routes()

	logger.Info("server_starting",
		"port", ":1323",
		"address", "0.0.0.0:1323",
	)

	e.Logger.Fatal(e.Start(":1323"))
}
