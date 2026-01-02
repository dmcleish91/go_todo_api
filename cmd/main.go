package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/dmcleish91/go_todo_api/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type application struct {
	projects *models.ProjectModel
	tasks    *models.TaskModel
	labels   *models.LabelModel
	logger   *slog.Logger
}

var db *pgxpool.Pool

func main() {
	godotenv.Load()
	loadConfig()

	user := os.Getenv("user")
	password := os.Getenv("password")
	host := os.Getenv("host")
	port := os.Getenv("port")
	dbname := os.Getenv("dbname")

	logger := NewStructuredLogger()

	DATABASE_URL := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, dbname)

	conn := CreateDatabaseConnection(DATABASE_URL)
	db = conn
	defer conn.Close()

	app := &application{
		projects: &models.ProjectModel{DB: conn},
		tasks:    &models.TaskModel{DB: conn},
		labels:   &models.LabelModel{DB: conn},
		logger:   logger,
	}

	e := app.Routes()

	logger.Info("starting server on :1323")
	e.Logger.Fatal(e.Start(":1323"))
}
