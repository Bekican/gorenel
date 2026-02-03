package errors

import (
	"errors"
	"net/http"

	"encoding/json"

	"github.com/Bekican/gorenel/pkg/logger"
	"go.uber.org/zap"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// api response'unun parametrelerini tanımlıyoruz
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *AppError   `json:"error,omitempty"`
}

// errorwrapper errorları yakalar,apperrorları handler,json formatında response döndürür
func ErrorWrapper(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//errorları anlık tespit edicez app crash olmasın diye
		defer func() {
			if rvr := recover(); rvr != nil {
				logger.Error("PANIC RECOVERED",
					zap.Any("panic", rvr),
					zap.String("path", r.URL.Path),
				)

				//500 status code user'a gönderiyoruz
				writeJSON(w, http.StatusInternalServerError, &APIResponse{
					Success: false,
					Error: &AppError{
						Type:    TypeInternal,
						Message: "Sunucu tarafında beklenmeyen bir hata oluştu.",
						Code:    http.StatusInternalServerError,
					},
				})
			}
		}()

		err := h(w, r)

		if err != nil {
			handleAppError(w, r, err)
		}
	}
}

// handleAppError processes different error types
func handleAppError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *AppError

	// Check if the error is our custom AppError
	if errors.As(err, &appErr) {
		// Log logic:
		// - 5xx errors -> Log as ERROR with stack trace
		// - 4xx errors -> Log as INFO or WARN (don't pollute error logs with bad user input)

		logFields := []zap.Field{
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.Int("code", appErr.Code),
			zap.String("error_type", string(appErr.Type)),
		}

		if appErr.Code >= 500 {
			// Add stack trace and original error for debugging
			logFields = append(logFields, zap.String("stack_trace", appErr.StackTrace()))
			if appErr.Err != nil {
				logFields = append(logFields, zap.String("original_error", appErr.Err.Error()))
			}
			logger.Error("Internal Server Error", logFields...)
		} else {
			// Just info for client errors
			logFields = append(logFields, zap.String("message", appErr.Message))
			logger.Warn("Client Error", logFields...)
		}

		// Send JSON response
		writeJSON(w, appErr.Code, &APIResponse{
			Success: false,
			Error:   appErr,
		})
		return
	}

	// If it's a standard Go error (not AppError), treat as 500 Internal
	logger.Error("Unhandled Error",
		zap.String("path", r.URL.Path),
		zap.Error(err),
	)

	writeJSON(w, http.StatusInternalServerError, &APIResponse{
		Success: false,
		Error:   Internal(err),
	})
}

// writeJSON helper to write secure JSON responses
func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
