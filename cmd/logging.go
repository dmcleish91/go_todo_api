package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func NewStructuredLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func StructuredLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			req := c.Request()
			res := c.Response()

			// Add a request ID to the context and response headers.
			// This is useful for tracking requests through the system.
			requestID := req.Header.Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = uuid.NewString()
			}
			c.Set(echo.HeaderXRequestID, requestID)
			res.Header().Set(echo.HeaderXRequestID, requestID)

			err := next(c)

			// If there is an error, log it.
			if err != nil {
				// To get the http status code, we can assert the error to an *echo.HTTPError
				httpError, ok := err.(*echo.HTTPError)
				if ok {
					res.Status = httpError.Code
				} else {
					// If it's not an echo.HTTPError, it's an internal server error.
					res.Status = http.StatusInternalServerError
				}
				c.Error(err)
			}

			// Log the request details
			logger.Info("request completed",
				"request_id", requestID,
				"method", req.Method,
				"uri", req.RequestURI,
				"status", res.Status,
				"latency", time.Since(start).String(),
				"remote_ip", c.RealIP(),
				"user_agent", req.UserAgent(),
			)

			return nil
		}
	}
}
