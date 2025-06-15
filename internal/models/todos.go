package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Todo struct {
	ID          int        `json:"todo_id"`
	UserID      string     `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	IsCompleted bool       `json:"is_completed"`
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Tags        []string   `json:"tags"`
}

type TodoModel struct {
	DB *pgxpool.Pool
}

func (m *TodoModel) AddTodo(todo Todo) (Todo, error) {
	query := `
			INSERT INTO todos (user_id, title, description, due_date, tags) 
			VALUES ($1, $2, $3, $4, $5) 
			RETURNING todo_id, user_id, title, description, is_completed, due_date, created_at, updated_at, tags`

	var createdTodo Todo
	err := m.DB.QueryRow(
		context.Background(),
		query,
		todo.UserID,
		todo.Title,
		todo.Description,
		todo.DueDate,
		todo.Tags,
	).Scan(
		&createdTodo.ID,
		&createdTodo.UserID,
		&createdTodo.Title,
		&createdTodo.Description,
		&createdTodo.IsCompleted,
		&createdTodo.DueDate,
		&createdTodo.CreatedAt,
		&createdTodo.UpdatedAt,
		&createdTodo.Tags,
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
		tags = $6,
    	updated_at = CURRENT_TIMESTAMP
	WHERE todo_id = $1 
	AND user_id = $2
	RETURNING todo_id, title, description, due_date, is_completed, updated_at, tags`

	var updatedTodo Todo
	err := m.DB.QueryRow(
		context.Background(),
		query,
		todo.ID,
		todo.UserID,
		todo.Title,
		todo.Description,
		todo.DueDate,
		todo.Tags,
	).Scan(
		&updatedTodo.ID,
		&updatedTodo.Title,
		&updatedTodo.Description,
		&updatedTodo.DueDate,
		&updatedTodo.IsCompleted,
		&updatedTodo.UpdatedAt,
		&updatedTodo.Tags,
	)

	if err != nil {
		return Todo{}, fmt.Errorf("unable to execute query: %v", err)
	}

	return updatedTodo, nil
}

func (m *TodoModel) GetTodosByUserID(userID string) ([]Todo, error) {
	query := `
    SELECT 
        todo_id, user_id, title, description, is_completed, due_date, created_at, updated_at, tags
    FROM todos
    WHERE user_id = $1
    ORDER BY created_at DESC`

	rows, err := m.DB.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to query todos: %v", err)
	}
	defer rows.Close()

	var todos []Todo

	for rows.Next() {
		var todo Todo
		err := rows.Scan(
			&todo.ID,
			&todo.UserID,
			&todo.Title,
			&todo.Description,
			&todo.IsCompleted,
			&todo.DueDate,
			&todo.CreatedAt,
			&todo.UpdatedAt,
			&todo.Tags,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %v", err)
		}

		todos = append(todos, todo)
	}

	return todos, nil
}

func (m *TodoModel) ToggleTodoCompleted(todoID int, userID string) (int64, error) {
	query := "UPDATE todos SET is_completed = NOT is_completed, updated_at = CURRENT_TIMESTAMP WHERE todo_id = $1 AND user_id = $2"

	result, err := m.DB.Exec(context.Background(), query, todoID, userID)
	if err != nil {
		return 0, fmt.Errorf("unable to update todo: %v", err)
	}

	return result.RowsAffected(), nil
}

func (m *TodoModel) DeleteTodoByID(todoID int, userID string) (int64, error) {
	query := "DELETE FROM todos WHERE todo_id = $1 AND user_id = $2"

	result, err := m.DB.Exec(context.Background(), query, todoID, userID)
	if err != nil {
		return 0, fmt.Errorf("unable to delete todo: %v", err)
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

func ValidateTodo(todo *Todo, v *Validator) {
	v.Check(todo.Title != "", "title", "Title must not be empty")
	if len(todo.Title) > 100 {
		v.AddError("title", "Title must not exceed 100 characters")
	}
}
