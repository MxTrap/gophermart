package logger

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
	"testing"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger()

	t.Run("logger creation", func(t *testing.T) {
		assert.NotNil(t, logger)
		assert.NotNil(t, logger.SugaredLogger)
	})

	t.Run("production level", func(t *testing.T) {
		// Проверяем, что логгер имеет уровень Info или выше (как в NewProduction)
		assert.True(t, logger.SugaredLogger.Desugar().Core().Enabled(zapcore.InfoLevel))
		assert.False(t, logger.SugaredLogger.Desugar().Core().Enabled(zapcore.DebugLevel))
	})

	t.Run("logging methods", func(t *testing.T) {
		// Проверяем, что основные методы логирования не паникуют
		logger.Info("test info")
		logger.Error("test error")
		logger.Warn("test warn")
		logger.With("key", "value").Info("test with fields")
		assert.NotPanics(t, func() {
			logger.SugaredLogger.Sync()
		})
	})
}
