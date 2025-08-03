package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// NewStructuredLogger creates a new JSON logger with proper configuration
func NewStructuredLogger() *slog.Logger {
	// Configure JSON handler with proper options
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: false, // Set to true if you want source file/line info
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

// StructuredLogger middleware for comprehensive request logging
func StructuredLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			req := c.Request()
			res := c.Response()

			// Generate or use existing request ID
			requestID := req.Header.Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = uuid.NewString()
			}
			c.Set(echo.HeaderXRequestID, requestID)
			res.Header().Set(echo.HeaderXRequestID, requestID)

			// Log request start
			logger.Info("request started",
				"request_id", requestID,
				"method", req.Method,
				"uri", req.RequestURI,
				"path", req.URL.Path,
				"query", req.URL.RawQuery,
				"remote_ip", c.RealIP(),
				"user_agent", req.UserAgent(),
				"content_length", req.ContentLength,
				"content_type", req.Header.Get("Content-Type"),
				"user_id", GetUserID(c),
			)

			err := next(c)

			// Handle errors and set status
			if err != nil {
				httpError, ok := err.(*echo.HTTPError)
				if ok {
					res.Status = httpError.Code
				} else {
					res.Status = http.StatusInternalServerError
				}
				c.Error(err)
			}

			// Calculate latency
			latency := time.Since(start)

			// Log request completion with comprehensive details
			logger.Info("request completed",
				"request_id", requestID,
				"method", req.Method,
				"uri", req.RequestURI,
				"path", req.URL.Path,
				"status", res.Status,
				"status_text", http.StatusText(res.Status),
				"latency_ms", latency.Milliseconds(),
				"latency_ns", latency.Nanoseconds(),
				"remote_ip", c.RealIP(),
				"user_agent", req.UserAgent(),
				"content_length", req.ContentLength,
				"response_size", res.Size,
				"user_id", GetUserID(c),
				"error", err != nil,
			)

			// Log errors separately with more detail
			if err != nil {
				logger.Error("request error",
					"request_id", requestID,
					"method", req.Method,
					"uri", req.RequestURI,
					"status", res.Status,
					"error", err.Error(),
					"user_id", GetUserID(c),
					"latency_ms", latency.Milliseconds(),
				)
			}

			return nil
		}
	}
}
