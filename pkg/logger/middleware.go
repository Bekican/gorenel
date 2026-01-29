package logger

import (
	"net/http"
)

// response writer http.ResponseWriterı kapsıyor -> statusCode'u öğrenmek için
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}
