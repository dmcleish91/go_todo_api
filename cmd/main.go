package main

import (
	"fmt"
	"os"

	"github.com/dmcleish91/go_todo_api/internal/models"
	"github.com/joho/godotenv"
)

type application struct {
	users         *models.UserModel
	todos         *models.TodoModel
	refreshTokens *models.RefreshTokenModel
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
		users:         &models.UserModel{DB: conn},
		todos:         &models.TodoModel{DB: conn},
		refreshTokens: &models.RefreshTokenModel{DB: conn},
	}

	e := app.Routes()

	e.Logger.Fatal(e.Start(":1323"))
}
