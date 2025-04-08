package main

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dmcleish91/go_todo_api/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

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
func (app *application) AddNewTodo(c echo.Context) error {
	var todo models.Todo
	if err := c.Bind(&todo); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	if todo.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	userID := GetUserIdFromToken(c)
	if userID == -1 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "This server can't process this request"})
	}

	todo.UserID = userID

	todoID, err := app.todos.AddTodo(todo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Todo added successfully",
		"todo ID": todoID,
	})
}

// Edit a existing todo
func (app *application) EditExistingTodo(c echo.Context) error {
	userId := GetUserIdFromToken(c)
	var todo models.Todo
	if err := c.Bind(&todo); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	if todo.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}
	todo.UserID = userId

	rowsAffected, err := app.todos.EditTodoByID(todo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "Todo updated successfully",
		"rows_affected": rowsAffected,
	})
}

func (app *application) GetTodosByUserID(c echo.Context) error {
	userID := GetUserIdFromToken(c)
	if userID == -1 {
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

	userID := GetUserIdFromToken(c)
	if userID == -1 {
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

	userID := GetUserIdFromToken(c)
	if userID == -1 {
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

func (app *application) AddTagToTodo(c echo.Context) error {
	userID := GetUserIdFromToken(c)
	if userID == -1 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "This server can't process this request"})
	}

	todoIDStr := c.QueryParam("todo_id")
	todoID, err := strconv.Atoi(todoIDStr)
	if err != nil || todoIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	tagIDstr := c.QueryParam("tag_id")
	tagID, err := strconv.Atoi(tagIDstr)
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

func (app *application) AddNewTag(c echo.Context) error {
	userID := GetUserIdFromToken(c)
	if userID == -1 {
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

func createJwtToken(userID int, username string) (string, error) {
	claims := &jwtCustomClaims{
		Sub:   fmt.Sprintf("%d", userID),
		Name:  username,
		Admin: false,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

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

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		log.Println("Error converting user ID to int64:", err)
		return -1
	}

	return int(userId)
}

// Alternative implementation using regex
func CleanStringRegex(input string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	return reg.ReplaceAllString(strings.TrimSpace(input), "")
}
