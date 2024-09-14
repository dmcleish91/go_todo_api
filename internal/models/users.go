package models

import (
    "context"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
    ID           int       `json:"user_id"`
    Username     string    `json:"username"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"password_hash"`
    CreatedAt    time.Time `json:"created_at"`
}

type UserModel struct {
    DB *pgxpool.Pool
}

// Register a new user
func (m *UserModel) RegisterNewUser(user User) (int64, error) {
    query := "INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)"

    result, err := m.DB.Exec(context.Background(), query, user.Username, user.Email, user.PasswordHash)
    if err != nil {
        return 0, fmt.Errorf("unable to register user: %v", err)
    }

    return result.RowsAffected(), nil
}

// Fetch a user by email
func (m *UserModel) GetUserByEmail(email string) (*User, error) {
    query := "SELECT user_id, username, email, password_hash, created_at FROM users WHERE email = $1"

    row := m.DB.QueryRow(context.Background(), query, email)
    user := &User{}

    err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
    if err != nil {
        return nil, fmt.Errorf("unable to find user by email: %v", err)
    }

    return user, nil
}

// Check if an email exists
func (m *UserModel) EmailExists(email string) (bool, error) {
    query := "SELECT COUNT(*) FROM users WHERE email = $1"

    var count int
    err := m.DB.QueryRow(context.Background(), query, email).Scan(&count)
    if err != nil {
        return false, fmt.Errorf("unable to check email: %v", err)
    }

    return count > 0, nil
}

// Update the email of an authenticated user
func (m *UserModel) UpdateUserEmail(oldEmail string, newEmail string) (int64, error) {
    query := "UPDATE users SET email = $1 WHERE email = $2"

    result, err := m.DB.Exec(context.Background(), query, newEmail, oldEmail)
    if err != nil {
        return 0, fmt.Errorf("unable to update email: %v", err)
    }

    return result.RowsAffected(), nil
}

// Update the username of an authenticated user
func (m *UserModel) UpdateUsername(email string, newUsername string) (int64, error) {
    query := "UPDATE users SET username = $1 WHERE email = $2"

    result, err := m.DB.Exec(context.Background(), query, newUsername, email)
    if err != nil {
        return 0, fmt.Errorf("unable to update username: %v", err)
    }

    return result.RowsAffected(), nil
}

// Get user details by user ID
func (m *UserModel) GetUserByID(userID int) (*User, error) {
    query := "SELECT user_id, username, email, password_hash, created_at FROM users WHERE user_id = $1"

    row := m.DB.QueryRow(context.Background(), query, userID)
    user := &User{}

    err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
    if err != nil {
        return nil, fmt.Errorf("unable to find user by ID: %v", err)
    }

    return user, nil
}

// Delete a user by user ID (optional, in case of user removal)
func (m *UserModel) DeleteUserByID(userID int) (int64, error) {
    query := "DELETE FROM users WHERE user_id = $1"

    result, err := m.DB.Exec(context.Background(), query, userID)
    if err != nil {
        return 0, fmt.Errorf("unable to delete user: %v", err)
    }

    return result.RowsAffected(), nil
}
