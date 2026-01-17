package mongo

import (
	"sync"

	"github.com/lazygophers/log"
)

// Logger interface for MongoDB logging (for SQL output)
type Logger interface {
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

type logger struct {
	logger *log.Logger
}

var (
	syncOnce sync.Once
	_logger  Logger
)

// GetDefaultLogger returns a default logger if not set
func GetDefaultLogger() Logger {
	syncOnce.Do(func() {
		if _logger == nil {
			_logger = NewLogger()
		}
	})
	return _logger
}

// SetDefaultLogger sets the default logger
func SetDefaultLogger(l Logger) {
	_logger = l
}

// NewLogger creates a new logger
func NewLogger() Logger {
	return &logger{
		logger: log.Clone().SetCallerDepth(5),
	}
}

// Infof logs an info message
func (l *logger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

// Warnf logs a warning message
func (l *logger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

// Errorf logs an error message
func (l *logger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

// Debugf logs a debug message
func (l *logger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}
