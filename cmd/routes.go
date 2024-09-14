package main

import (
	"os"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type jwtCustomClaims struct {
	Sub   string `json:"sub"`  // Subject (user ID)
	Name  string `json:"name"` // Name of the user
	Admin bool   `json:"admin"`
	jwt.RegisteredClaims
}

func (app *application) Routes() *echo.Echo {
	signingKey := os.Getenv("SigningKey")

	e := echo.New()

	e.Use(ServerHeader)

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `[${time_rfc3339} ${status} ${method} ${host}${path} ${latency_human}]` + "\n",
	}))

	secured := e.Group("/secure")

	
	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwtCustomClaims)
		},
		SigningMethod: "HS256",
		SigningKey:    []byte(signingKey),
	}

	secured.Use(echojwt.WithConfig(config))

	e.POST("/register", app.RegisterUser)

	e.POST("/login", app.Login)

	return e
}

func ServerHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderServer, "PetAgency/0.1")

		return next(c)
	}
}