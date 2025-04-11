package logger

import (
	"go.uber.org/zap"
)

type Logger struct {
	*zap.SugaredLogger
}

func NewLogger() *Logger {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	return &Logger{
		sugar,
	}
}
