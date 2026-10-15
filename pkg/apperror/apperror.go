package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Status  int    `json:"-"`
	Message string `json:"message"`
	Details any    `json:"errors,omitempty"`
}

func (e *AppError) Error() string { return e.Message }

func New(status int, message string) *AppError {
	return &AppError{Status: status, Message: message}
}

func BadRequest(msg string) *AppError   { return New(http.StatusBadRequest, msg) }
func Unauthorized(msg string) *AppError { return New(http.StatusUnauthorized, msg) }
func Forbidden(msg string) *AppError    { return New(http.StatusForbidden, msg) }
func Conflict(msg string) *AppError     { return New(http.StatusConflict, msg) }

func NotFound(entity string) *AppError {
	return New(http.StatusNotFound, fmt.Sprintf("%s tidak ditemukan", entity))
}

func Validation(details map[string]string) *AppError {
	return &AppError{Status: http.StatusUnprocessableEntity, Message: "data tidak valid", Details: details}
}

func As(err error) (*AppError, bool) {
	var appErr *AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}
