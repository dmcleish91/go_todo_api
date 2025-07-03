package main

import (
	"net/http"

	"github.com/dmcleish91/go_todo_api/internal/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func GetUserID(c echo.Context) string {
	if userID, ok := c.Get("user_id").(string); ok {
		return userID
	}
	return ""
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
	var input models.NewTask
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input", "message": err.Error()})
	}

	// Create a new validator
	v := models.NewValidator()

	// Validate the input (customize as needed for NewTask)
	if input.Content == "" {
		v.AddError("content", "Content is required")
	}
	// You can add more validation for other fields if needed

	if !v.Valid() {
		return c.JSON(http.StatusUnprocessableEntity, map[string]any{"errors": v.Errors})
	}

	userID := GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	created, err := app.tasks.AddTask(input, uid)
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

func (app *application) ToggleTaskCompletion(c echo.Context) error {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
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

	updatedTask, err := app.tasks.ToggleTaskCompleted(taskID, uid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]any{"message": "Task updated successfully", "data": updatedTask})
}

// Label Handlers
func (app *application) AddNewLabel(c echo.Context) error {
	var input struct {
		Name string `json:"name"`
	}
	if err := c.Bind(&input); err != nil {
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
	label := models.Label{UserID: uid, Name: input.Name}
	v := models.NewValidator()
	models.ValidateLabel(&label, v)
	if !v.Valid() {
		return c.JSON(http.StatusUnprocessableEntity, map[string]any{"errors": v.Errors})
	}
	created, err := app.labels.AddLabel(uid, input.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]any{"message": "Label added successfully", "data": created})
}

func (app *application) EditExistingLabel(c echo.Context) error {
	var input struct {
		LabelID string `json:"label_id"`
		Name    string `json:"name"`
	}
	if err := c.Bind(&input); err != nil {
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
	labelID, err := uuid.Parse(input.LabelID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid label ID"})
	}
	label := models.Label{LabelID: labelID, UserID: uid, Name: input.Name}
	v := models.NewValidator()
	models.ValidateLabel(&label, v)
	if !v.Valid() {
		return c.JSON(http.StatusUnprocessableEntity, map[string]any{"errors": v.Errors})
	}
	updated, err := app.labels.EditLabelByID(labelID, uid, input.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "Label updated successfully", "data": updated})
}

func (app *application) GetLabelsByUserID(c echo.Context) error {
	userID := GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}
	labels, err := app.labels.GetLabelsByUserID(uid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, labels)
}

func (app *application) DeleteLabel(c echo.Context) error {
	labelIDStr := c.QueryParam("label_id")
	labelID, err := uuid.Parse(labelIDStr)
	if err != nil || labelIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid label ID"})
	}
	userID := GetUserID(c)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}
	rowsAffected, err := app.labels.DeleteLabelByID(labelID, uid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "Label deleted successfully", "rows_affected": rowsAffected})
}
