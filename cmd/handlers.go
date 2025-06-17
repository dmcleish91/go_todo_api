package main

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/dmcleish91/go_todo_api/internal/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (app *application) AddNewTodo(c echo.Context) error {
	var todo models.Todo
	if err := c.Bind(&todo); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	v := models.NewValidator()
	models.ValidateTodo(&todo, v)
	if !v.Valid() {
		return c.JSON(http.StatusUnprocessableEntity, map[string]any{"errors": v.Errors})
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

// Project Handlers
func (app *application) AddNewProject(c echo.Context) error {
	var project models.Project
	if err := c.Bind(&project); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	userID := GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}
	project.UserID = uid

	created, err := app.projects.AddProject(project)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]any{"message": "Project added successfully", "data": created})
}

func (app *application) EditExistingProject(c echo.Context) error {
	var project models.Project
	if err := c.Bind(&project); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}
	userID := GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}
	project.UserID = uid

	updated, err := app.projects.EditProjectByID(project)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "Project updated successfully", "data": updated})
}

func (app *application) GetProjectsByUserID(c echo.Context) error {
	userID := GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}
	projects, err := app.projects.GetProjectsByUserID(uid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, projects)
}

func (app *application) DeleteProject(c echo.Context) error {
	projectIDStr := c.QueryParam("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil || projectIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid project ID"})
	}
	userID := GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}
	rowsAffected, err := app.projects.DeleteProjectByID(projectID, uid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "Project deleted successfully", "rows_affected": rowsAffected})
}

// Task Handlers
func (app *application) AddNewTask(c echo.Context) error {
	var task models.Task
	if err := c.Bind(&task); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}
	userID := GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}
	task.UserID = uid
	if task.ProjectID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Project ID is required"})
	}
	created, err := app.tasks.AddTask(task)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]any{"message": "Task added successfully", "data": created})
}

func (app *application) EditExistingTask(c echo.Context) error {
	var task models.Task
	if err := c.Bind(&task); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}
	userID := GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}
	task.UserID = uid
	if task.ProjectID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Project ID is required"})
	}
	updated, err := app.tasks.EditTaskByID(task)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "Task updated successfully", "data": updated})
}

func (app *application) GetTasksByUserID(c echo.Context) error {
	userID := GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}
	tasks, err := app.tasks.GetTasksByUserID(uid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, tasks)
}

func (app *application) DeleteTask(c echo.Context) error {
	taskIDStr := c.QueryParam("task_id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil || taskIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid task ID"})
	}
	userID := GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}
	rowsAffected, err := app.tasks.DeleteTaskByID(taskID, uid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "Task deleted successfully", "rows_affected": rowsAffected})
}
