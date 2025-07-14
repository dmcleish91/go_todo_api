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
	ProjectID    *uuid.UUID `json:"project_id"`
	UserID       uuid.UUID  `json:"user_id"`
	Content      string     `json:"content"`
	Description  string     `json:"description"`
	DueDate      *time.Time `json:"due_date"`     // nullable time.Time: use nil for null
	DueDatetime  *time.Time `json:"due_datetime"` // nullable time.Time: use nil for null
	Priority     int16      `json:"priority"`
	IsCompleted  bool       `json:"is_completed"`
	CompletedAt  *time.Time `json:"completed_at"` // nullable time.Time: use nil for null
	ParentTaskID *uuid.UUID `json:"parent_task_id"`
	Order        int        `json:"order"`
	Labels       []string   `json:"labels"`
	CreatedAt    time.Time  `json:"created_at"`
}

type TaskModel struct {
	DB *pgxpool.Pool
}

// NewTask is used for creating a new task from API input
// All fields are optional except content and task_id
// Fields correspond to nullable columns in the DB
// user_id is not included; it comes from JWT
// task_id is required; must be provided by frontend
type NewTask struct {
	TaskID       uuid.UUID  `json:"task_id"`              // REQUIRED: Frontend must provide task_id
	ProjectID    *uuid.UUID `json:"project_id,omitempty"`
	Content      string     `json:"content"`
	Description  *string    `json:"description,omitempty"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	DueDatetime  *time.Time `json:"due_datetime,omitempty"`
	Priority     *int16     `json:"priority,omitempty"`
	ParentTaskID *uuid.UUID `json:"parent_task_id,omitempty"`
	Labels       []string   `json:"labels,omitempty"`
	Order        *int       `json:"order,omitempty"`
}

// AddTask inserts a new task into the database using NewTask and userID
func (m *TaskModel) AddTask(input NewTask, userID uuid.UUID) (Task, error) {
	query := `
		INSERT INTO tasks (
			task_id, project_id, user_id, content, description, due_date, due_datetime, priority, parent_task_id, "order", labels
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING task_id, project_id, user_id, content, description, due_date, due_datetime, priority, is_completed, completed_at, parent_task_id, "order", labels
	`

	var createdTask Task
	orderValue := 0
	if input.Order != nil {
		orderValue = *input.Order
	}
	err := m.DB.QueryRow(
		context.Background(),
		query,
		input.TaskID,        // Use the provided task_id
		input.ProjectID,
		userID,
		input.Content,
		input.Description,
		input.DueDate,
		input.DueDatetime,
		input.Priority,
		input.ParentTaskID,
		orderValue,
		input.Labels,
	).Scan(
		&createdTask.TaskID,
		&createdTask.ProjectID,
		&createdTask.UserID,
		&createdTask.Content,
		&createdTask.Description,
		&createdTask.DueDate,
		&createdTask.DueDatetime,
		&createdTask.Priority,
		&createdTask.IsCompleted,
		&createdTask.CompletedAt,
		&createdTask.ParentTaskID,
		&createdTask.Order,
		&createdTask.Labels,
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
			due_datetime = $7,
			priority = $8,
			is_completed = $9,
			completed_at = $10,
			parent_task_id = $11,
			"order" = $12,
			labels = $13
		WHERE task_id = $1 AND user_id = $2
		RETURNING task_id, project_id, user_id, content, description, due_date, due_datetime, priority, is_completed, completed_at, parent_task_id, "order", labels
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
		task.DueDatetime,
		task.Priority,
		task.IsCompleted,
		task.CompletedAt,
		task.ParentTaskID,
		task.Order,
		task.Labels,
	).Scan(
		&updatedTask.TaskID,
		&updatedTask.ProjectID,
		&updatedTask.UserID,
		&updatedTask.Content,
		&updatedTask.Description,
		&updatedTask.DueDate,
		&updatedTask.DueDatetime,
		&updatedTask.Priority,
		&updatedTask.IsCompleted,
		&updatedTask.CompletedAt,
		&updatedTask.ParentTaskID,
		&updatedTask.Order,
		&updatedTask.Labels,
	)

	if err != nil {
		return Task{}, fmt.Errorf("unable to execute query: %v", err)
	}

	return updatedTask, nil
}

func (m *TaskModel) GetTasksByUserID(userID uuid.UUID) ([]Task, error) {
	query := `
		SELECT task_id, project_id, user_id, content, description, due_date, due_datetime, priority, is_completed, completed_at, parent_task_id, "order", labels, created_at
		FROM tasks
		WHERE user_id = $1
		ORDER BY created_at ASC`

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
			&task.DueDatetime,
			&task.Priority,
			&task.IsCompleted,
			&task.CompletedAt,
			&task.ParentTaskID,
			&task.Order,
			&task.Labels,
			&task.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %v", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (m *TaskModel) ToggleTaskCompleted(taskID uuid.UUID, userID uuid.UUID) (Task, error) {
	query := `
		UPDATE tasks
		SET completed_at = CASE WHEN is_completed THEN NULL ELSE CURRENT_TIMESTAMP END, is_completed = NOT is_completed
		WHERE task_id = $1 AND user_id = $2
		RETURNING task_id, project_id, user_id, content, description, due_date, due_datetime, priority, is_completed, completed_at, parent_task_id, labels`

	var updatedTask Task
	err := m.DB.QueryRow(
		context.Background(),
		query,
		taskID,
		userID,
	).Scan(
		&updatedTask.TaskID,
		&updatedTask.ProjectID,
		&updatedTask.UserID,
		&updatedTask.Content,
		&updatedTask.Description,
		&updatedTask.DueDate,
		&updatedTask.DueDatetime,
		&updatedTask.Priority,
		&updatedTask.IsCompleted,
		&updatedTask.CompletedAt,
		&updatedTask.ParentTaskID,
		&updatedTask.Labels,
	)

	if err != nil {
		return Task{}, fmt.Errorf("unable to execute query: %v", err)
	}

	return updatedTask, nil
}

func (m *TaskModel) DeleteTaskByID(taskID uuid.UUID, userID uuid.UUID) (int64, error) {
	query := `DELETE FROM tasks WHERE task_id = $1 AND user_id = $2`

	result, err := m.DB.Exec(context.Background(), query, taskID, userID)
	if err != nil {
		return 0, fmt.Errorf("unable to delete task: %v", err)
	}

	return result.RowsAffected(), nil
}

// ValidateTask validates a Task object
func ValidateTask(task *Task, v *Validator) {
	// Validate due date format if provided
	if task.DueDate != nil {
		dateStr := task.DueDate.Format("01/02/2006")
		v.Check(dateStr != "", "due_date", "Due date must be in MM/DD/YYYY format")
	}

	// Validate due_datetime format if provided
	if task.DueDatetime != nil {
		timeStr := task.DueDatetime.Format("03:04 PM")
		v.Check(timeStr != "", "due_datetime", "Due datetime must be in XX:XX AM/PM format")
	}
}

// BulkUpdateTaskOrder updates the order of sibling tasks for a user, project, and parent_task_id.
// All tasks must belong to the same user, project, and parent_task_id.
type TaskOrderUpdate struct {
	TaskID string `json:"task_id"`
	Order  int    `json:"order"`
}

func (m *TaskModel) BulkUpdateTaskOrder(userID uuid.UUID, projectID *uuid.UUID, parentTaskID *uuid.UUID, updates []TaskOrderUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := m.DB.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		}
	}()

	// Build the CASE statement and collect task IDs
	caseStmt := "CASE"
	taskIDs := make([]string, 0, len(updates))
	for _, upd := range updates {
		caseStmt += " WHEN task_id = '" + upd.TaskID + "' THEN " + fmt.Sprintf("%d", upd.Order)
		taskIDs = append(taskIDs, "'"+upd.TaskID+"'")
	}
	caseStmt += " END"

	// Build the WHERE clause for sibling tasks
	where := "user_id = $1"
	args := []interface{}{userID}
	argIdx := 2
	if projectID != nil {
		where += fmt.Sprintf(" AND project_id = $%d", argIdx)
		args = append(args, *projectID)
		argIdx++
	} else {
		where += fmt.Sprintf(" AND project_id IS NULL")
	}
	if parentTaskID != nil {
		where += fmt.Sprintf(" AND parent_task_id = $%d", argIdx)
		args = append(args, *parentTaskID)
		argIdx++
	} else {
		where += fmt.Sprintf(" AND parent_task_id IS NULL")
	}

	where += fmt.Sprintf(" AND task_id IN (%s)", joinStrings(taskIDs, ", "))

	query := fmt.Sprintf(`UPDATE tasks SET "order" = %s WHERE %s`, caseStmt, where)

	_, err = tx.Exec(context.Background(), query, args...)
	if err != nil {
		tx.Rollback(context.Background())
		return fmt.Errorf("failed to update task order: %w", err)
	}

	if err = tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// joinStrings joins a slice of strings with a separator.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// GetTaskByID fetches a single task by task_id and user_id
func (m *TaskModel) GetTaskByID(taskID uuid.UUID, userID uuid.UUID) (Task, error) {
	query := `
		SELECT task_id, project_id, user_id, content, description, due_date, due_datetime, priority, is_completed, completed_at, parent_task_id, "order", labels, created_at
		FROM tasks
		WHERE task_id = $1 AND user_id = $2
	`
	var task Task
	err := m.DB.QueryRow(context.Background(), query, taskID, userID).Scan(
		&task.TaskID,
		&task.ProjectID,
		&task.UserID,
		&task.Content,
		&task.Description,
		&task.DueDate,
		&task.DueDatetime,
		&task.Priority,
		&task.IsCompleted,
		&task.CompletedAt,
		&task.ParentTaskID,
		&task.Order,
		&task.Labels,
		&task.CreatedAt,
	)
	if err != nil {
		return Task{}, fmt.Errorf("unable to fetch task: %w", err)
	}
	return task, nil
}
