package logger

import (
	"github.com/sigif/sigif-go/internal/shared/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"sync"
)

var (
	globalLogger *zap.Logger
	loggerMu     sync.RWMutex
)

var Module = fx.Provide(NewLogger)

func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	if cfg.App.Env == "production" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	level, err := zapcore.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	var output io.Writer = os.Stdout
	if cfg.Logging.Output != "stdout" {
		file, err := os.OpenFile(cfg.Logging.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		output = file
	}

	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.ErrorOutputPaths = []string{"stderr"}

	logger, err := zapConfig.Build(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, zapcore.NewCore(
			zapcore.NewJSONEncoder(zapConfig.EncoderConfig),
			zapcore.AddSync(output),
			zapConfig.Level,
		))
	}))
	if err != nil {
		return nil, err
	}

	loggerMu.Lock()
	globalLogger = logger
	loggerMu.Unlock()

	return logger, nil
}

func L() *zap.Logger {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	if globalLogger == nil {
		return zap.NewNop()
	}
	return globalLogger
}
