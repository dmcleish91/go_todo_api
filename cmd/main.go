package main

import (
	"fmt"
	"os"

	"github.com/dmcleish91/go_todo_api/internal/models"
	"github.com/joho/godotenv"
)

type application struct {
	todos    *models.TodoModel
	projects *models.ProjectModel
	tasks    *models.TaskModel
}

func main() {
	godotenv.Load()
	user := os.Getenv("user")
	password := os.Getenv("password")
	host := os.Getenv("host")
	port := os.Getenv("port")
	dbname := os.Getenv("dbname")

	DATABASE_URL := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, dbname)

	conn := CreateDatabaseConnection(DATABASE_URL)
	defer conn.Close()

	app := &application{
		todos:    &models.TodoModel{DB: conn},
		projects: &models.ProjectModel{DB: conn},
		tasks:    &models.TaskModel{DB: conn},
	}

	e := app.Routes()

	e.Logger.Fatal(e.Start(":1323"))
}
