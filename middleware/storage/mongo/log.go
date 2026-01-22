package mongo

import (
	"io"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gookit/color"
	"github.com/lazygophers/log"
)

// Logger interface for MongoDB logging (for SQL output)
type Logger interface {
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Log(skip int, begin time.Time, fc func() (operation string, docsAffected int64), err error)
	SetOutput(writes ...io.Writer) *logger
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

// SetOutput sets the output writers for the logger
func (l *logger) SetOutput(writes ...io.Writer) *logger {
	l.logger.SetOutput(writes...)
	return l
}

// Log logs MongoDB operations with caller information and timing
// Similar to database/sql logging, but for MongoDB operations
func (l *logger) Log(skip int, begin time.Time, fc func() (operation string, docsAffected int64), err error) {
	var callerName string
	pc, file, callerLine, ok := runtime.Caller(skip)
	if ok {
		callerName = runtime.FuncForPC(pc).Name()
	}

	callerDir, callerFunc := log.SplitPackageName(callerName)
	b := log.GetBuffer()
	defer log.PutBuffer(b)
	b.WriteString(color.Yellow.Sprintf("%s:%d %s", path.Join(callerDir, path.Base(file)), callerLine, callerFunc))
	b.WriteString(" ")

	b.WriteString(color.Blue.Sprintf("[%s]", time.Since(begin).Truncate(time.Microsecond)))
	b.WriteString(" ")

	operation, docsAffected := fc()
	b.WriteString(strings.ReplaceAll(operation, "\n", " "))
	b.WriteString(" ")

	b.WriteString(color.Blue.Sprintf("[%d docs]", docsAffected))

	if err == nil {
		l.logger.Info(b.String())
	} else {
		l.logger.Error(b.String())
	}
}
