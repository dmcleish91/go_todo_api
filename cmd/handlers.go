package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
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
		return ctx.String(http.StatusInternalServerError, "Error retriving user from database")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return ctx.String(http.StatusUnauthorized, "email or password was incorrect")
	}

	if subtle.ConstantTimeCompare([]byte(email), []byte(user.Email)) != 1 {
		return ctx.String(http.StatusUnauthorized, "email or password was incorrect")
	}

	accessToken, err := createJwtToken(user.ID, user.Username)
	if err != nil {
		log.Println("Error Creating JWT token")
		return ctx.String(http.StatusInternalServerError, "Error create access token")
	}

	refreshToken, err := generateSecureToken()
	if err != nil {
		return ctx.String(http.StatusInternalServerError, "Error generating refresh token string")
	}
	refreshExpiry := time.Now().Add(time.Hour * 24 * 7)

	err = app.refreshTokens.CreateRefreshToken(user.ID, refreshToken, refreshExpiry)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, "Error creating refresh token")
	}

	cookie := new(http.Cookie)
	cookie.Name = "refresh_token"
	cookie.Value = refreshToken
	cookie.Expires = refreshExpiry
	cookie.HttpOnly = true
	cookie.Secure = false
	cookie.Path = "/v1/refresh-token"
	//cookie.SameSite = http.SameSiteDefaultMode
	ctx.SetCookie(cookie)

	return ctx.JSON(http.StatusOK, map[string]any{
		"ok": true,
		"data": map[string]any{
			"access_token": accessToken,
			"token_type":   "Bearer",
			"expires_in":   3600,
		},
	})
}

func (app *application) Logout(c echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Expires:  time.Now().Add(-1 * time.Hour),
		Path:     "/v1/refresh-token",
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Successfully logged out",
	})
}

func (app *application) RefreshToken(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Refresh token cookie not found",
		})
	}
	refreshToken := cookie.Value

	tokenData, err := app.refreshTokens.ValidateRefreshToken(refreshToken)
	if err != nil {
		c.SetCookie(&http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			HttpOnly: true,
			Expires:  time.Now().Add(-1 * time.Hour),
			Path:     "/v1/refresh-token",
		})
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid refresh token",
		})
	}

	user, err := app.users.GetUserByID(tokenData.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Error fetching user details",
		})
	}

	newAccessToken, err := createJwtToken(user.ID, user.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Error generating new access token",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"ok": true,
		"data": map[string]any{
			"access_token": newAccessToken,
			"token_type":   "Bearer",
			"expires_in":   3600,
		},
	})
}

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

	result, err := app.todos.EditTodoByID(todo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Todo updated successfully",
		"data":    result,
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
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
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

func generateSecureToken() (string, error) {
	// Generate 32 random bytes (256 bits)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}

	// Encode bytes to base64URL (URL-safe version of base64)
	// We use RawURLEncoding to avoid padding characters
	token := base64.RawURLEncoding.EncodeToString(bytes)

	return token, nil
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
