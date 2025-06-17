package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Task struct {
	TaskID       uuid.UUID  `json:"task_id"`
	ProjectID    uuid.UUID  `json:"project_id"`
	UserID       uuid.UUID  `json:"user_id"`
	Content      string     `json:"content"`
	Description  *string    `json:"description"`
	DueDate      *time.Time `json:"due_date"`      // nullable time.Time: use nil for null
	DateDatetime *time.Time `json:"date_datetime"` // nullable time.Time: use nil for null
	Priority     *int16     `json:"priority"`
	IsCompleted  *bool      `json:"is_completed"`
	CompletedAt  *time.Time `json:"completed_at"` // nullable time.Time: use nil for null
	ParentTaskID *uuid.UUID `json:"parent_task_id"`
}

type TaskModel struct {
	DB *pgxpool.Pool
}

func (m *TaskModel) AddTask(task Task) (Task, error) {
	query := `
		INSERT INTO tasks (
			project_id, user_id, content, description, due_date, date_datetime, priority, is_completed, completed_at, parent_task_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING task_id, project_id, user_id, content, description, due_date, date_datetime, priority, is_completed, completed_at, parent_task_id
	`

	var createdTask Task
	err := m.DB.QueryRow(
		context.Background(),
		query,
		task.ProjectID,
		task.UserID,
		task.Content,
		task.Description,
		task.DueDate,
		task.DateDatetime,
		task.Priority,
		task.IsCompleted,
		task.CompletedAt,
		task.ParentTaskID,
	).Scan(
		&createdTask.TaskID,
		&createdTask.ProjectID,
		&createdTask.UserID,
		&createdTask.Content,
		&createdTask.Description,
		&createdTask.DueDate,
		&createdTask.DateDatetime,
		&createdTask.Priority,
		&createdTask.IsCompleted,
		&createdTask.CompletedAt,
		&createdTask.ParentTaskID,
	)

	if err != nil {
		return Task{}, fmt.Errorf("unable to execute query: %v", err)
	}

	return createdTask, nil
}

func (m *TaskModel) EditTaskByID(task Task) (Task, error) {
	query := `
		UPDATE tasks SET
			project_id = $3,
			content = $4,
			description = $5,
			due_date = $6,
			date_datetime = $7,
			priority = $8,
			is_completed = $9,
			completed_at = $10,
			parent_task_id = $11
		WHERE task_id = $1 AND user_id = $2
		RETURNING task_id, project_id, user_id, content, description, due_date, date_datetime, priority, is_completed, completed_at, parent_task_id
	`

	var updatedTask Task
	err := m.DB.QueryRow(
		context.Background(),
		query,
		task.TaskID,
		task.UserID,
		task.ProjectID,
		task.Content,
		task.Description,
		task.DueDate,
		task.DateDatetime,
		task.Priority,
		task.IsCompleted,
		task.CompletedAt,
		task.ParentTaskID,
	).Scan(
		&updatedTask.TaskID,
		&updatedTask.ProjectID,
		&updatedTask.UserID,
		&updatedTask.Content,
		&updatedTask.Description,
		&updatedTask.DueDate,
		&updatedTask.DateDatetime,
		&updatedTask.Priority,
		&updatedTask.IsCompleted,
		&updatedTask.CompletedAt,
		&updatedTask.ParentTaskID,
	)

	if err != nil {
		return Task{}, fmt.Errorf("unable to execute query: %v", err)
	}

	return updatedTask, nil
}

func (m *TaskModel) GetTasksByUserID(userID uuid.UUID) ([]Task, error) {
	query := `
		SELECT task_id, project_id, user_id, content, description, due_date, date_datetime, priority, is_completed, completed_at, parent_task_id
		FROM tasks
		WHERE user_id = $1
		ORDER BY due_date DESC, completed_at DESC, content ASC`

	rows, err := m.DB.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to query tasks: %v", err)
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.TaskID,
			&task.ProjectID,
			&task.UserID,
			&task.Content,
			&task.Description,
			&task.DueDate,
			&task.DateDatetime,
			&task.Priority,
			&task.IsCompleted,
			&task.CompletedAt,
			&task.ParentTaskID,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %v", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (m *TaskModel) ToggleTaskCompleted(taskID uuid.UUID, userID uuid.UUID) (int64, error) {
	query := `UPDATE tasks SET is_completed = NOT is_completed, completed_at = CASE WHEN is_completed THEN NULL ELSE CURRENT_TIMESTAMP END WHERE task_id = $1 AND user_id = $2`

	result, err := m.DB.Exec(context.Background(), query, taskID, userID)
	if err != nil {
		return 0, fmt.Errorf("unable to update task: %v", err)
	}

	return result.RowsAffected(), nil
}

func (m *TaskModel) DeleteTaskByID(taskID uuid.UUID, userID uuid.UUID) (int64, error) {
	query := `DELETE FROM tasks WHERE task_id = $1 AND user_id = $2`

	result, err := m.DB.Exec(context.Background(), query, taskID, userID)
	if err != nil {
		return 0, fmt.Errorf("unable to delete task: %v", err)
	}

	return result.RowsAffected(), nil
}
