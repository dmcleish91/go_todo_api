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

			requestID := req.Header.Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = uuid.NewString()
			}
			c.Set(echo.HeaderXRequestID, requestID)
			res.Header().Set(echo.HeaderXRequestID, requestID)

			err := next(c)

			if err != nil {
				httpError, ok := err.(*echo.HTTPError)
				if ok {
					res.Status = httpError.Code
				} else {
					res.Status = http.StatusInternalServerError
				}
				c.Error(err)
			}

			// Skip logging for OPTIONS requests
			if req.Method != http.MethodOptions {
				logger.Info("request completed",
					"request_id", requestID,
					"method", req.Method,
					"uri", req.RequestURI,
					"status", res.Status,
					"latency", time.Since(start).String(),
					"remote_ip", c.RealIP(),
					"user_agent", req.UserAgent(),
				)
			}

			return nil
		}
	}
}
