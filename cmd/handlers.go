package main

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dmcleish91/go_todo_api/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// User registration handler
func (app *application) RegisterUser(c echo.Context) error {
	username := c.FormValue("username")
	email := c.FormValue("email")
	password := c.FormValue("password")

	if email == "" || password == "" || username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	user := models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	rowsAffected, err := app.users.RegisterNewUser(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message":       "User registered successfully",
		"rows_affected": rowsAffected,
	})
}

func (app *application) Login(ctx echo.Context) error {
	email := ctx.FormValue("email")
	password := ctx.FormValue("password")

	user, err := app.users.GetUserByEmail(email)
	if err != nil {
		log.Printf("Error: %s", err)
		return ctx.String(http.StatusInternalServerError, "Something went wrong")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return ctx.String(http.StatusUnauthorized, "email or password was incorrect")
	}

	if subtle.ConstantTimeCompare([]byte(email), []byte(user.Email)) != 1 {
		return ctx.String(http.StatusUnauthorized, "email or password was incorrect")
	}

	token, err := createJwtToken(user.ID, user.Username)
	if err != nil {
		log.Println("Error Creating JWT token")
		return ctx.String(http.StatusInternalServerError, "Something went wrong")
	}

	return ctx.JSON(http.StatusOK, map[string]string{
		"message": "You were logged in!",
		"token":   token,
	})
}

// Update user email
func (app *application) UpdateUserEmail(c echo.Context) error {
	type UpdateEmailInput struct {
		NewEmail string `json:"new_email"`
	}
	email := c.Param("email")
	var input UpdateEmailInput

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	rowsAffected, err := app.users.UpdateUserEmail(email, input.NewEmail)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "Email updated successfully",
		"rows_affected": rowsAffected,
	})
}

// Add a new todo
func (app *application) AddTodoHandler(c echo.Context) error {
	var todo models.Todo
	if err := c.Bind(&todo); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	userID := GetUserIdFromToken(c)
	if userID == -1 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Server can't process request"})
	}

	todo.UserID = userID

	rowsAffected, err := app.todos.AddTodo(todo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message":       "Todo added successfully",
		"rows_affected": rowsAffected,
	})
}

// Get todos by user ID
func (app *application) GetTodosByUserID(c echo.Context) error {
	userID := GetUserIdFromToken(c)
	if userID == -1 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Server can't process request"})
	}

	todos, err := app.todos.GetTodosByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, todos)
}

// Mark a todo as completed
func (app *application) MarkTodoCompleteHandler(c echo.Context) error {
	todoID, err := strconv.Atoi(c.Param("todo_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid todo ID"})
	}

	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	rowsAffected, err := app.todos.MarkTodoComplete(todoID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "Todo marked as completed",
		"rows_affected": rowsAffected,
	})
}

// Delete a todo
func (app *application) DeleteTodoHandler(c echo.Context) error {
	todoID, err := strconv.Atoi(c.Param("todo_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid todo ID"})
	}

	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
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

// Add a tag to a todo
func (app *application) AddTagToTodoHandler(c echo.Context) error {
	todoID, err := strconv.Atoi(c.Param("todo_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid todo ID"})
	}

	tagID, err := strconv.Atoi(c.Param("tag_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid tag ID"})
	}

	rowsAffected, err := app.todos.AddTagToTodo(todoID, tagID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "Tag added successfully",
		"rows_affected": rowsAffected,
	})
}

// These are utility functions

func createJwtToken(userID int, username string) (string, error) {
	// Set custom claims
	claims := &jwtCustomClaims{
		Sub:   fmt.Sprintf("%d", userID),
		Name:  username,
		Admin: false,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signingKey := os.Getenv("SigningKey")

	tokenString, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetUserIdFromToken(c echo.Context) int {
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*jwtCustomClaims)
	userIdStr := claims.Sub

	log.Println("Decoded user id is", userIdStr)

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		log.Println("Error converting user ID to int64:", err)
		return -1
	}

	return int(userId)
}
