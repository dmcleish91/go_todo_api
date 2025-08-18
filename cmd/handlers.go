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

func (app *application) AddNewProject(c echo.Context) error {
	var project models.Project
	if err := c.Bind(&project); err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}

	userID := GetUserID(c)
	if userID == "" {
		return app.sendError(c, http.StatusUnauthorized, ErrAuthentication, "")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	project.UserID = uid

	created, err := app.projects.AddProject(project)
	if err != nil {
		return app.handleDatabaseError(c, err, "add_project")
	}
	return c.JSON(http.StatusCreated, map[string]any{"message": "Project created successfully", "data": created})
}

func (app *application) EditExistingProject(c echo.Context) error {
	var project models.Project
	if err := c.Bind(&project); err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	userID := GetUserID(c)
	if userID == "" {
		return app.sendError(c, http.StatusUnauthorized, ErrAuthentication, "")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	project.UserID = uid

	updated, err := app.projects.EditProjectByID(project)
	if err != nil {
		return app.handleDatabaseError(c, err, "edit_project")
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "Project updated successfully", "data": updated})
}

func (app *application) GetProjectsByUserID(c echo.Context) error {
	userID := GetUserID(c)
	if userID == "" {
		return app.sendError(c, http.StatusUnauthorized, ErrAuthentication, "")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	projects, err := app.projects.GetProjectsByUserID(uid)
	if err != nil {
		return app.handleDatabaseError(c, err, "get_projects")
	}
	return c.JSON(http.StatusOK, projects)
}

func (app *application) DeleteProject(c echo.Context) error {
	projectIDStr := c.QueryParam("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil || projectIDStr == "" {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	userID := GetUserID(c)
	if userID == "" {
		return app.sendError(c, http.StatusUnauthorized, ErrAuthentication, "")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	rowsAffected, err := app.projects.DeleteProjectByID(projectID, uid)
	if err != nil {
		return app.handleDatabaseError(c, err, "delete_project")
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "Project deleted successfully", "rows_affected": rowsAffected})
}

func (app *application) AddNewTask(c echo.Context) error {
	var input models.NewTask
	if err := c.Bind(&input); err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}

	v := models.NewValidator()

	if input.Content == "" {
		v.AddError("content", "Content is required")
	}

	if input.TaskID == uuid.Nil {
		v.AddError("task_id", "Task ID is required")
	}

	if input.Order != nil && *input.Order < 0 {
		v.AddError("order", "Order must be non-negative")
	}

	if !v.Valid() {
		return app.sendValidationError(c, v.Errors)
	}

	userID := GetUserID(c)
	if userID == "" {
		return app.sendError(c, http.StatusUnauthorized, ErrAuthentication, "")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}

	created, err := app.tasks.AddTask(input, uid)
	if err != nil {
		return app.handleDatabaseError(c, err, "add_task")
	}

	return c.JSON(http.StatusCreated, map[string]any{"message": "Task created successfully", "data": created})
}

func (app *application) EditExistingTask(c echo.Context) error {
	var task models.Task
	if err := c.Bind(&task); err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	userID := GetUserID(c)
	if userID == "" {
		return app.sendError(c, http.StatusUnauthorized, ErrAuthentication, "")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	task.UserID = uid

	updated, err := app.tasks.EditTaskByID(task)
	if err != nil {
		return app.handleDatabaseError(c, err, "edit_task")
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "Task updated successfully", "data": updated})
}

func (app *application) GetTasksByUserID(c echo.Context) error {
	userID := GetUserID(c)
	if userID == "" {
		return app.sendError(c, http.StatusUnauthorized, ErrAuthentication, "")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	tasks, err := app.tasks.GetTasksByUserID(uid)
	if err != nil {
		return app.handleDatabaseError(c, err, "get_tasks")
	}
	return c.JSON(http.StatusOK, tasks)
}

func (app *application) DeleteTask(c echo.Context) error {
	taskIDStr := c.QueryParam("task_id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil || taskIDStr == "" {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	userID := GetUserID(c)
	if userID == "" {
		return app.sendError(c, http.StatusUnauthorized, ErrAuthentication, "")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	rowsAffected, err := app.tasks.DeleteTaskByID(taskID, uid)
	if err != nil {
		return app.handleDatabaseError(c, err, "delete_task")
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "Task deleted successfully", "rows_affected": rowsAffected})
}

func (app *application) ToggleTaskCompletion(c echo.Context) error {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}

	userID := GetUserID(c)
	if userID == "" {
		return app.sendError(c, http.StatusUnauthorized, ErrAuthentication, "")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}

	updatedTask, err := app.tasks.ToggleTaskCompleted(taskID, uid)
	if err != nil {
		return app.handleDatabaseError(c, err, "toggle_task_completion")
	}

	return c.JSON(http.StatusOK, map[string]any{"message": "Task updated successfully", "data": updatedTask})
}

func (app *application) AddNewLabel(c echo.Context) error {
	var input struct {
		Name string `json:"name"`
	}
	if err := c.Bind(&input); err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	userID := GetUserID(c)
	if userID == "" {
		return app.sendError(c, http.StatusUnauthorized, ErrAuthentication, "")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	label := models.Label{UserID: uid, Name: input.Name}
	v := models.NewValidator()
	models.ValidateLabel(&label, v)
	if !v.Valid() {
		return app.sendValidationError(c, v.Errors)
	}
	created, err := app.labels.AddLabel(uid, input.Name)
	if err != nil {
		return app.handleDatabaseError(c, err, "add_label")
	}
	return c.JSON(http.StatusCreated, map[string]any{"message": "Label created successfully", "data": created})
}

func (app *application) EditExistingLabel(c echo.Context) error {
	var input struct {
		LabelID string `json:"label_id"`
		Name    string `json:"name"`
	}
	if err := c.Bind(&input); err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	userID := GetUserID(c)
	if userID == "" {
		return app.sendError(c, http.StatusUnauthorized, ErrAuthentication, "")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	labelID, err := uuid.Parse(input.LabelID)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	label := models.Label{LabelID: labelID, UserID: uid, Name: input.Name}
	v := models.NewValidator()
	models.ValidateLabel(&label, v)
	if !v.Valid() {
		return app.sendValidationError(c, v.Errors)
	}
	updated, err := app.labels.EditLabelByID(labelID, uid, input.Name)
	if err != nil {
		return app.handleDatabaseError(c, err, "edit_label")
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "Label updated successfully", "data": updated})
}

func (app *application) GetLabelsByUserID(c echo.Context) error {
	userID := GetUserID(c)
	if userID == "" {
		return app.sendError(c, http.StatusUnauthorized, ErrAuthentication, "")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	labels, err := app.labels.GetLabelsByUserID(uid)
	if err != nil {
		return app.handleDatabaseError(c, err, "get_labels")
	}
	return c.JSON(http.StatusOK, labels)
}

func (app *application) DeleteLabel(c echo.Context) error {
	labelIDStr := c.QueryParam("label_id")
	labelID, err := uuid.Parse(labelIDStr)
	if err != nil || labelIDStr == "" {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	userID := GetUserID(c)
	if userID == "" {
		return app.sendError(c, http.StatusUnauthorized, ErrAuthentication, "")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	rowsAffected, err := app.labels.DeleteLabelByID(labelID, uid)
	if err != nil {
		return app.handleDatabaseError(c, err, "delete_label")
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "Label deleted successfully", "rows_affected": rowsAffected})
}

func (app *application) HandleReorderTasks(c echo.Context) error {
	userID := GetUserID(c)
	if userID == "" {
		return app.sendError(c, http.StatusUnauthorized, ErrAuthentication, "")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}

	var updates []models.TaskOrderUpdate
	if err := c.Bind(&updates); err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	if len(updates) == 0 {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}

	firstTaskID, err := uuid.Parse(updates[0].TaskID)
	if err != nil {
		return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
	}
	task, err := app.tasks.GetTaskByID(firstTaskID, uid)
	if err != nil {
		return app.sendError(c, http.StatusNotFound, ErrNotFound, "")
	}
	projectID := task.ProjectID
	parentTaskID := task.ParentTaskID

	for _, upd := range updates {
		id, err := uuid.Parse(upd.TaskID)
		if err != nil {
			return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
		}
		t, err := app.tasks.GetTaskByID(id, uid)
		if err != nil {
			return app.sendError(c, http.StatusNotFound, ErrNotFound, "")
		}
		if (t.ProjectID == nil && projectID != nil) || (t.ProjectID != nil && projectID == nil) || (t.ProjectID != nil && projectID != nil && *t.ProjectID != *projectID) {
			return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
		}
		if (t.ParentTaskID == nil && parentTaskID != nil) || (t.ParentTaskID != nil && parentTaskID == nil) || (t.ParentTaskID != nil && parentTaskID != nil && *t.ParentTaskID != *parentTaskID) {
			return app.sendError(c, http.StatusBadRequest, ErrInvalidInput, "")
		}
	}

	if err := app.tasks.BulkUpdateTaskOrder(uid, projectID, parentTaskID, updates); err != nil {
		return app.handleDatabaseError(c, err, "reorder_tasks")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Task order updated successfully"})
}
