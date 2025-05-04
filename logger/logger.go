package logger

import (
	"go.uber.org/zap"
)

type Logger struct {
	*zap.SugaredLogger
}

func NewLogger() *Logger {
	logger, _ := zap.NewProduction()
	logger = logger.WithOptions(zap.WithCaller(false), zap.AddStacktrace(zap.FatalLevel))
	sugar := logger.Sugar()
	return &Logger{
		sugar,
	}
}
