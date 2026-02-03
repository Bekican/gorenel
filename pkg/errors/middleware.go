package errors

import (
	"net/http"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// api response'unun parametrelerini tanımlıyoruz
type APIResponse struct {
	Sucess bool        `json:"sucess"`
	Data   interface{} `json:"data,omitempty"`
	Error  *AppError   `json:"error,omitempty"`
}
