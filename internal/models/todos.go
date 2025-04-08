package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Todo struct {
	ID          int        `json:"todo_id"`
	UserID      int        `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	IsCompleted bool       `json:"is_completed"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Tags        []string   `json:"tags,omitempty"`
}

type TodoModel struct {
	DB *pgxpool.Pool
}

func (m *TodoModel) AddTodo(todo Todo) (Todo, error) {
	query := `
			INSERT INTO todos (user_id, title, description, due_date) 
			VALUES ($1, $2, $3, $4) 
			RETURNING todo_id, user_id, title, description, is_completed, due_date, created_at, updated_at`

	var createdTodo Todo
	err := m.DB.QueryRow(
		context.Background(),
		query,
		todo.UserID,
		todo.Title,
		todo.Description,
		todo.DueDate,
	).Scan(
		&createdTodo.ID,
		&createdTodo.UserID,
		&createdTodo.Title,
		&createdTodo.Description,
		&createdTodo.IsCompleted,
		&createdTodo.DueDate,
		&createdTodo.CreatedAt,
		&createdTodo.UpdatedAt,
	)

	if err != nil {
		return Todo{}, fmt.Errorf("unable to execute query: %v", err)
	}

	return createdTodo, nil
}

func (m *TodoModel) EditTodoByID(todo Todo) (Todo, error) {
	query := `
	UPDATE todos 
	SET title = $3,
    	description = $4,
    	due_date = $5,
    	updated_at = CURRENT_TIMESTAMP
	WHERE todo_id = $1 
	AND user_id = $2
	RETURNING todo_id, title, description, due_date, is_completed, updated_at`

	var updatedTodo Todo
	err := m.DB.QueryRow(
		context.Background(),
		query,
		todo.ID,
		todo.UserID,
		todo.Title,
		todo.Description,
		todo.DueDate,
	).Scan(
		&updatedTodo.ID,
		&updatedTodo.Title,
		&updatedTodo.Description,
		&updatedTodo.DueDate,
		&updatedTodo.IsCompleted,
		&updatedTodo.UpdatedAt,
	)

	if err != nil {
		return Todo{}, fmt.Errorf("unable to execute query: %v", err)
	}

	return updatedTodo, nil
}

func (m *TodoModel) GetTodosByUserID(userID int) ([]Todo, error) {
	query := `
    SELECT 
        t.todo_id, t.user_id, t.title, t.description, t.is_completed, t.due_date, t.created_at, t.updated_at, 
        STRING_AGG(tag.name, ', ') AS tags
    FROM todos t
    LEFT JOIN todo_tags tt ON t.todo_id = tt.todo_id
    LEFT JOIN tags tag ON tt.tag_id = tag.tag_id
    WHERE t.user_id = $1
    GROUP BY t.todo_id
    ORDER BY t.created_at DESC`

	rows, err := m.DB.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to query todos: %v", err)
	}
	defer rows.Close()

	var todos []Todo

	for rows.Next() {
		var todo Todo
		var tags *string

		err := rows.Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.IsCompleted, &todo.DueDate, &todo.CreatedAt, &todo.UpdatedAt, &tags)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %v", err)
		}

		if tags != nil {
			todo.Tags = splitTags(*tags) // Helper function to split tags string into slice
		}

		todos = append(todos, todo)
	}

	return todos, nil
}

func (m *TodoModel) ToggleTodoCompleted(todoID int, userID int) (int64, error) {
	query := "UPDATE todos SET is_completed = NOT is_completed, updated_at = CURRENT_TIMESTAMP WHERE todo_id = $1 AND user_id = $2"

	result, err := m.DB.Exec(context.Background(), query, todoID, userID)
	if err != nil {
		return 0, fmt.Errorf("unable to update todo: %v", err)
	}

	return result.RowsAffected(), nil
}

func (m *TodoModel) DeleteTodoByID(todoID int, userID int) (int64, error) {
	query := "DELETE FROM todos WHERE todo_id = $1 AND user_id = $2"

	result, err := m.DB.Exec(context.Background(), query, todoID, userID)
	if err != nil {
		return 0, fmt.Errorf("unable to delete todo: %v", err)
	}

	return result.RowsAffected(), nil
}

func (m *TodoModel) AddTagToTodo(todoID int, tagID int) (int64, error) {
	query := "INSERT INTO todo_tags (todo_id, tag_id) VALUES ($1, $2)"

	result, err := m.DB.Exec(context.Background(), query, todoID, tagID)
	if err != nil {
		return 0, fmt.Errorf("unable to add tag: %v", err)
	}

	return result.RowsAffected(), nil
}

func (m *TodoModel) AddTag(name string) (int, error) {
	query := "INSERT INTO tags (name) VALUES ($1) RETURNING tag_id"

	var tagID int
	err := m.DB.QueryRow(context.Background(), query, name).Scan(&tagID)
	if err != nil {
		return 0, fmt.Errorf("unable to add tag: %v", err)
	}

	return tagID, nil
}

// Helper function to split the tags string into a slice
func splitTags(tagString string) []string {
	if tagString == "" {
		return []string{}
	}
	return strings.Split(tagString, ", ")
}
