package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
}

type APIError string

const (
	ErrInvalidInput       APIError = "invalid_input"
	ErrAuthentication     APIError = "authentication_required"
	ErrAuthorization      APIError = "access_denied"
	ErrNotFound           APIError = "resource_not_found"
	ErrInternal           APIError = "internal_error"
	ErrDatabase           APIError = "database_error"
	ErrValidation         APIError = "validation_failed"
	ErrBadRequest         APIError = "bad_request"
	ErrConflict           APIError = "conflict"
	ErrUnprocessable      APIError = "unprocessable_entity"
	ErrTooManyRequests    APIError = "too_many_requests"
	ErrServiceUnavailable APIError = "service_unavailable"
)

// Generic, non-revealing error messages for clients
var clientErrorMessages = map[APIError]string{
	ErrInvalidInput:       "The provided data is invalid",
	ErrAuthentication:     "Authentication is required",
	ErrAuthorization:      "Access to this resource is denied",
	ErrNotFound:           "The requested resource was not found",
	ErrInternal:           "An internal error occurred",
	ErrDatabase:           "A database error occurred",
	ErrValidation:         "The provided data failed validation",
	ErrBadRequest:         "The request could not be processed",
	ErrConflict:           "The request conflicts with the current state",
	ErrUnprocessable:      "The request could not be processed",
	ErrTooManyRequests:    "Too many requests, please try again later",
	ErrServiceUnavailable: "Service temporarily unavailable",
}

func getRequestID(c echo.Context) string {
	if requestID, ok := c.Get("X-Request-ID").(string); ok {
		return requestID
	}
	return "unknown"
}

func (app *application) sendError(c echo.Context, statusCode int, errorType APIError, message string) error {
	clientMsg := clientErrorMessages[errorType]
	if message != "" {
		clientMsg = message
	}

	response := ErrorResponse{
		Error:   string(errorType),
		Message: clientMsg,
	}

	app.logger.Error("api_error",
		"request_id", getRequestID(c),
		"error_type", string(errorType),
		"status_code", statusCode,
		"status_text", http.StatusText(statusCode),
		"user_id", GetUserID(c),
		"path", c.Request().URL.Path,
		"method", c.Request().Method,
		"uri", c.Request().RequestURI,
		"remote_ip", c.RealIP(),
		"user_agent", c.Request().UserAgent(),
		"client_message", clientMsg,
		"custom_message", message,
	)

	return c.JSON(statusCode, response)
}

func (app *application) sendValidationError(c echo.Context, errors map[string]string) error {
	response := ErrorResponse{
		Error:  string(ErrValidation),
		Errors: errors,
	}

	app.logger.Warn("validation_error",
		"request_id", getRequestID(c),
		"error_type", string(ErrValidation),
		"status_code", http.StatusUnprocessableEntity,
		"user_id", GetUserID(c),
		"path", c.Request().URL.Path,
		"method", c.Request().Method,
		"uri", c.Request().RequestURI,
		"remote_ip", c.RealIP(),
		"validation_errors", errors,
		"error_count", len(errors),
	)

	return c.JSON(http.StatusUnprocessableEntity, response)
}

func (app *application) handleDatabaseError(c echo.Context, err error, operation string) error {
	app.logger.Error("database_error",
		"request_id", getRequestID(c),
		"operation", operation,
		"error_type", string(ErrDatabase),
		"status_code", http.StatusInternalServerError,
		"user_id", GetUserID(c),
		"path", c.Request().URL.Path,
		"method", c.Request().Method,
		"uri", c.Request().RequestURI,
		"remote_ip", c.RealIP(),
		"error_message", err.Error(),
		"error_details", err,
	)

	return app.sendError(c, http.StatusInternalServerError, ErrDatabase, "")
}
