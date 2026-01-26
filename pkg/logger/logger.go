package logger

import (
	"sync"

	"go.uber.org/zap"
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

//sensible defaults
