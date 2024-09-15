package main

import (
	"os"

	"github.com/dmcleish91/go_todo_api/internal/models"
	"github.com/joho/godotenv"
)

type application struct {
	users *models.UserModel
	todos  *models.TodoModel
}

func main() {
	godotenv.Load()
	DATABASE_URL := os.Getenv("DATABASE_URL")

	conn := CreateDatabaseConnection(DATABASE_URL)
	defer conn.Close()

	app := &application{
		users: &models.UserModel{DB: conn},
		todos:  &models.TodoModel{DB: conn},
	}

	e := app.Routes()

	e.Logger.Fatal(e.Start(":1323"))
}