package helpers

import "net/http"

type AppError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *AppError) Error() string {
	return e.Code + ": " + e.Message
}

func NewAppError(statusCode int, code, message string) *AppError {
	return &AppError{StatusCode: statusCode, Code: code, Message: message}
}

func BadRequest(code, message string) *AppError {
	return NewAppError(http.StatusBadRequest, code, message)
}

func Unauthorized(code, message string) *AppError {
	return NewAppError(http.StatusUnauthorized, code, message)
}

func Forbidden(code, message string) *AppError {
	return NewAppError(http.StatusForbidden, code, message)
}

func NotFound(code, message string) *AppError {
	return NewAppError(http.StatusNotFound, code, message)
}

func Conflict(code, message string) *AppError {
	return NewAppError(http.StatusConflict, code, message)
}

func UnprocessableEntity(code, message string) *AppError {
	return NewAppError(http.StatusUnprocessableEntity, code, message)
}
