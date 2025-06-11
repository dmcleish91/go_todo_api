package main

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/dmcleish91/go_todo_api/internal/models"
	"github.com/labstack/echo/v4"
)

func (app *application) AddNewTodo(c echo.Context) error {
	var todo models.Todo
	if err := c.Bind(&todo); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	if todo.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	userID := GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "This server can't process this request"})
	}

	todo.UserID = userID

	todoID, err := app.todos.AddTodo(todo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message": "Todo added successfully",
		"data":    todoID,
	})
}

func (app *application) EditExistingTodo(c echo.Context) error {
	userId := GetUserID(c)
	var todo models.Todo
	if err := c.Bind(&todo); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	if todo.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}
	todo.UserID = userId

	result, err := app.todos.EditTodoByID(todo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Todo updated successfully",
		"data":    result,
	})
}

func (app *application) GetTodosByUserID(c echo.Context) error {
	userID := GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "This server can't process this request"})
	}

	todos, err := app.todos.GetTodosByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, todos)
}

func (app *application) ToggleTodoCompleted(c echo.Context) error {
	todoIDStr := c.QueryParam("todo_id")
	todoID, err := strconv.Atoi(todoIDStr)
	if err != nil || todoIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	userID := GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "This server can't process this request"})
	}

	rowsAffected, err := app.todos.ToggleTodoCompleted(todoID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "Todo status has been toggled",
		"rows_affected": rowsAffected,
	})
}

func (app *application) DeleteTodo(c echo.Context) error {
	todoIDStr := c.QueryParam("todo_id")
	todoID, err := strconv.Atoi(todoIDStr)
	if err != nil || todoIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	userID := GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "This server can't process this request"})
	}

	rowsAffected, err := app.todos.DeleteTodoByID(todoID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "Todo deleted successfully",
		"rows_affected": rowsAffected,
	})
}

func (app *application) AddNewTag(c echo.Context) error {
	userID := GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "This server can't process this request"})
	}

	tagName := c.QueryParam("tag_name")
	cleanedTagName := CleanStringRegex(tagName)
	if cleanedTagName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "tags must contain a nonempty value"})
	}

	tagId, err := app.todos.AddTag(cleanedTagName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Tag added successfully",
		"tag_id":  tagId,
	})
}

// Alternative implementation using regex
func CleanStringRegex(input string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	return reg.ReplaceAllString(strings.TrimSpace(input), "")
}
