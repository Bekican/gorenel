package errors

import (
	"fmt"
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
