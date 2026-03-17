package errors

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

type ErrorType string

const (
	TypeValidation   ErrorType = "VALIDATION_ERROR"
	TypeNotFound     ErrorType = "NOT_FOUND"
	TypeUnauthorized ErrorType = "UNAUTHORIZED"
	TypeInternal     ErrorType = "INTERNAL_ERROR"
	TypeConflict     ErrorType = "CONFLICT"
	TypeForbidden    ErrorType = "FORBIDDEN"
)

type AppError struct {
	Type        ErrorType         `json:"type"`
	Code        int               `json:"code"`
	Message     string            `json:"message"`
	FieldErrors map[string]string `json:"fields,omitempty"`
	Err         error             `json:"-"`
	stack       string
}

// standard error interface'in entegre ediyoruz.
func (e *AppError) Error() string {
	return e.Message
}

// error altındaki parametreleri unwrap et
func (e *AppError) Unwrap() error {
	return e.Err
}

// yığın izleme
func (e *AppError) StackTrace() string {
	return e.stack
}

// yığını çağırma debugging için
func captureStack() string {
	var sb strings.Builder

	for i := 3; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		if idx := strings.LastIndex(file, "/gorenel/"); idx != -1 {
			file = file[idx+1:]
		}
		sb.WriteString(fmt.Sprintf("%s:%d\n", file, line))
	}
	return sb.String()
}

// app error oluşturuyoruz yığına göre
func New(errType ErrorType, code int, msg string, cause error) *AppError {
	return &AppError{
		Type:    errType,
		Code:    code,
		Message: msg,
		Err:     cause,
		stack:   captureStack(),
	}
}

func NotFound(message string, cause error) *AppError {
	return New(TypeNotFound, http.StatusNotFound, message, cause)
}

func BadRequest(message string, cause error) *AppError {
	return New(TypeNotFound, http.StatusBadRequest, message, cause)
}

func Internal(cause error) *AppError {
	return New(TypeInternal, http.StatusInternalServerError, "Bir hata oluştu lütfen daha sonra tekrar deneyin.", cause)
}

func ValidationError(message string, fields map[string]string) *AppError {
	err := New(TypeValidation, http.StatusUnprocessableEntity, message, nil)
	err.FieldErrors = fields
	return err
}

func Unauthorized(message string) *AppError {
	return New(TypeUnauthorized, http.StatusUnauthorized, message, nil)
}

func Forbidden(message string) *AppError {
	return New(TypeForbidden, http.StatusForbidden, message, nil)
}
