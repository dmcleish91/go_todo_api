package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type SupabaseJWTClaims struct {
	Sub   string `json:"sub"`   // User ID from Supabase
	Email string `json:"email"` // User email
	Role  string `json:"role"`  // User role
	jwt.RegisteredClaims
}

func (app *application) Routes() *echo.Echo {
	e := echo.New()

	e.Use(ServerHeader)

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `[${time_rfc3339} ${status} ${method} ${host}${path} ${latency_human}]` + "\n",
	}))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "https://yata-one.vercel.app"},
		AllowMethods:     []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	secured := e.Group("/v1")

	secured.Use(app.SupabaseJWTMiddleware())

	secured.GET("/todos", app.GetTodosByUserID)

	secured.POST("/todos", app.AddNewTodo)

	secured.POST("/editTodo", app.EditExistingTodo)

	secured.POST("/tag", app.AddNewTag)

	secured.POST("/addTodoTag", app.AddTagToTodo)

	secured.POST("/toggleComplete", app.ToggleTodoCompleted)

	secured.DELETE("/todos", app.DeleteTodo)

	return e
}

func (app *application) SupabaseJWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			signingKey := os.Getenv("SUPABASE_JWT_SIGNINGKEY")
			if signingKey == "" {
				return echo.NewHTTPError(http.StatusInternalServerError, "SUPABASE_JWT_SIGNINGKEY not set")
			}

			// Parse and validate the token
			claims := &SupabaseJWTClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
				// Verify the signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(signingKey), nil
			})

			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token: "+err.Error())
			}

			if !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			// Add user info to context
			c.Set("user_id", claims.Sub)
			c.Set("user_email", claims.Email)
			c.Set("user_role", claims.Role)

			return next(c)
		}
	}
}

func GetUserID(c echo.Context) string {
	if userID, ok := c.Get("user_id").(string); ok {
		return userID
	}
	return ""
}

func ServerHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderServer, "TodoApi/0.1")

		return next(c)
	}
}
