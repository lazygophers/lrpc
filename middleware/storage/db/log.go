package db

import (
	"context"
	"io"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/lazygophers/log"

	"github.com/gookit/color"
	"gorm.io/gorm/logger"
)

type Logger struct {
	logger *log.Logger
}

var (
	syncOnce sync.Once
	_logger  logger.Interface
)

func GetDefaultLogger() logger.Interface {
	syncOnce.Do(func() {
		if _logger != nil {
			_logger = NewLogger()
		}
	})
	return _logger
}

func SetDefaultLogger(l logger.Interface) {
	_logger = l
}

func NewLogger() *Logger {
	l := &Logger{
		logger: log.Clone().SetCallerDepth(5),
	}
	l.LogMode(logger.Info)
	return l
}

func (l *Logger) SetOutput(writes ...io.Writer) *Logger {
	l.logger.SetOutput(writes...)
	return l
}

func (l *Logger) LogMode(logLevel logger.LogLevel) logger.Interface {
	switch logLevel {
	case logger.Silent:
		l.logger.SetLevel(log.TraceLevel)
	case logger.Error:
		l.logger.SetLevel(log.ErrorLevel)
	case logger.Warn:
		l.logger.SetLevel(log.WarnLevel)
	case logger.Info:
		l.logger.SetLevel(log.InfoLevel)
	default:
		l.logger.SetLevel(log.DebugLevel)
	}
	return l
}

func (l *Logger) Info(ctx context.Context, s string, i ...interface{}) {
	l.logger.Infof(s, i...)
}

func (l *Logger) Warn(ctx context.Context, s string, i ...interface{}) {
	l.logger.Warnf(s, i...)
}

func (l *Logger) Error(ctx context.Context, s string, i ...interface{}) {
	l.logger.Errorf(s, i...)
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	l.Log(4, begin, fc, err)
}

func (l *Logger) Log(skip int, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
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

	sql, rowsAffected := fc()
	b.WriteString(strings.ReplaceAll(sql, "\n", " "))
	b.WriteString(" ")

	b.WriteString(color.Blue.Sprintf("[%d rows]", rowsAffected))

	if err == nil {
		l.logger.Info(b.String())
	} else {
		l.logger.Error(b.String())
	}
}

type mysqlLogger struct {
}

func (*mysqlLogger) Print(v ...interface{}) {
	// MySQL driver logger implementation - intentionally minimal
	// This satisfies the mysql driver's Logger interface requirement
	_ = v // explicitly use the parameter to satisfy linters
}
