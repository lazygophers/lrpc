package mongo

import (
	"testing"
)

func TestLogSetDefaultLogger(t *testing.T) {
	// Create a new logger
	newLogger := NewLogger()

	// Set it as default
	SetDefaultLogger(newLogger)

	// Get default logger and verify it's the one we set
	defaultLogger := GetDefaultLogger()

	if defaultLogger != newLogger {
		t.Error("expected default logger to be the one we set")
	}
}

func TestLogGetDefaultLogger(t *testing.T) {
	// First get default logger (should create one if not already set)
	logger1 := GetDefaultLogger()

	if logger1 == nil {
		t.Error("expected default logger, got nil")
	}

	// Get it again - should be the same instance
	logger2 := GetDefaultLogger()

	if logger1 != logger2 {
		t.Error("expected same logger instance on repeated calls")
	}
}

func TestLogNewLogger(t *testing.T) {
	logger := NewLogger()

	if logger == nil {
		t.Error("expected logger, got nil")
	}

	// Verify it implements Logger interface
	var _ Logger = logger
}

func TestLoggerInfof(t *testing.T) {
	logger := NewLogger()

	// Just verify it doesn't panic
	logger.Infof("test info message: %s", "value")
}

func TestLoggerWarnf(t *testing.T) {
	logger := NewLogger()

	// Just verify it doesn't panic
	logger.Warnf("test warning message: %s", "value")
}

func TestLoggerErrorf(t *testing.T) {
	logger := NewLogger()

	// Just verify it doesn't panic
	logger.Errorf("test error message: %s", "value")
}

func TestLoggerDebugf(t *testing.T) {
	logger := NewLogger()

	// Just verify it doesn't panic
	logger.Debugf("test debug message: %s", "value")
}
