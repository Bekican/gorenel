package logger

import (
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// global logger'ı tanımlıyoruz
var (
	Log  *zap.Logger
	once sync.Once
)

// config logger configurations
type Config struct {
	Level       string
	OutputPath  string
	Development bool
	Encoding    string
}

// sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Level:       "info",
		OutputPath:  "stdout",
		Development: false,
		Encoding:    "json",
	}
}

// global logger'ı initialize etme
func Init(cfg *Config) error {
	var err error
	once.Do(func() {
		Log, err = newLogger(cfg)
	})
	return err
}

// yeni zap logger oluşturuyoruz
func newLogger(cfg *Config) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	//encoder configuration
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	//encoder görünümü console or json
	var encoder zapcore.Encoder
	if cfg.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	//output config
	var output zapcore.WriteSyncer
	switch cfg.OutputPath {
	case "stdout":
		output = zapcore.AddSync(os.Stdout)
	case "stderr":
		output = zapcore.AddSync(os.Stderr)
	default:
		file, err := os.OpenFile(cfg.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		output = zapcore.AddSync(file)
	}

	//temel oluşturma
	core := zapcore.NewCore(encoder, output, level)

	//seçenklerle logger tanımlıyoruz
	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}

	if cfg.Development {
		opts = append(opts, zap.Development())
	}
	return zap.New(core, opts...), nil
}

// request id'ye göre child logger oluşturma - hangi request?
func WithRequestID(requestID string) *zap.Logger {
	return Log.With(zap.String("request_id", requestID))
}

// request id'nin yanına user etiketi atıyoruz
func WithFields(fields ...zap.Field) *zap.Logger {
	return Log.With(fields...)
}

// olaya göre log oluştur
func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}

// bellek kapanırken içerde log kaldıysa diske yazdıryoruz ki ramde log kalmasın
func Sync() error {
	if Log != nil {
		return Log.Sync()
	}
	return nil
}

// http trafiğini izliyoruz
func LogRequest(method, path string, statusCode int, duration time.Duration, requestID string) {
	Log.Info("http_request",
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
		zap.String("request_id", requestID),
	)
}
