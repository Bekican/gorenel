package main

import (
	"net/http"

	"github.com/Bekican/gorenel/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	//loggerı başlatıyoruz
	cfg := &logger.Config{
		Level:       "debug",
		OutputPath:  "stdout",
		Development: true,
		Encoding:    "console", //prodda json yap
	}

	if err := logger.Init(cfg); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("server_starting",
		zap.String("version", "1.0.0"),
		zap.Int("port", 8080),
	)

	reqLogger := logger.WithRequestID("abc-123")
	reqLogger.Debug("processing_request",
		zap.String("path", "/api/tunnels"),
	)

	logger.Error("connection_failed",
		zap.String("host", "example.com"),
		zap.Int("port", 5432),
		zap.Error(nil),
	)

	//HTTP server middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	handler := logger.ChainMiddleware(mux,
		logger.RecoveryMiddleware,
		logger.LoggingMiddleware,
	)

	logger.Info("server_started", zap.String("addr", ":8080"))
	http.ListenAndServe(":8080", handler)
}
