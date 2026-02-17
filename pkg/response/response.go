package response

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    Meta        `json:"meta"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"requestId"`
}

func newMeta() Meta {
	return Meta{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: uuid.New().String(),
	}
}

func Success(c *fiber.Ctx, status int, data interface{}, message string) error {
	return c.Status(status).JSON(Response{
		Success: true,
		Data:    data,
		Message: message,
		Meta:    newMeta(),
	})
}

func Error(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
		Meta: newMeta(),
	})
}
