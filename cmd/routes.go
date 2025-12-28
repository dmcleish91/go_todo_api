package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (app *application) Routes() *echo.Echo {
	e := echo.New()

	e.Use(ServerHeader)

	e.Use(StructuredLogger(app.logger))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "https://yata-delta.vercel.app"},
		AllowMethods:     []string{echo.GET, echo.PUT, echo.POST, echo.DELETE, echo.PATCH},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	secured := e.Group("/v1")

	// Project endpoints
	secured.POST("/projects", app.AddNewProject)
	secured.PUT("/projects", app.EditExistingProject)
	secured.GET("/projects", app.GetProjectsByUserID)
	secured.DELETE("/projects", app.DeleteProject)

	// Task endpoints
	secured.POST("/tasks", app.AddNewTask)
	secured.PUT("/tasks", app.EditExistingTask)
	secured.GET("/tasks", app.GetTasksByUserID)
	secured.DELETE("/tasks", app.DeleteTask)
	secured.PUT("/tasks/:id/toggle-completion", app.ToggleTaskCompletion)
	secured.PATCH("/tasks/reorder", app.HandleReorderTasks)

	// Label endpoints
	secured.POST("/labels", app.AddNewLabel)
	secured.PUT("/labels", app.EditExistingLabel)
	secured.GET("/labels", app.GetLabelsByUserID)
	secured.DELETE("/labels", app.DeleteLabel)

	// Reset endpoint
	secured.POST("/reset", app.ResetDemoData)

	return e
}


func ServerHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderServer, "TodoApi/0.1")

		return next(c)
	}
}
