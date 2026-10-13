package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Status  int               `json:"-"`
	Message string            `json:"message"`
	Details map[string]string `json:"errors,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(status int, message string) *AppError {
	return &AppError{
		Status:  status,
		Message: message,
	}
}

func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, message)
}

func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, message)
}

func Conflict(message string) *AppError {
	return New(http.StatusConflict, message)
}

func NotFound(entity string) *AppError {
	return New(http.StatusNotFound, fmt.Sprintf("%s not found", entity))
}

func ValidationError(details map[string]string) *AppError {
	return &AppError{
		Status:  http.StatusUnprocessableEntity,
		Message: "Invalid input",
		Details: details,
	}
}

func As(err error) (*AppError, bool) {
	var appErr *AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}
