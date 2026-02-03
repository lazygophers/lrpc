package db

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gormLog "gorm.io/gorm/logger"
)

// TestGetDefaultLogger 测试获取默认日志器
func TestGetDefaultLogger(t *testing.T) {
	logger := GetDefaultLogger()
	assert.NotNil(t, logger)

	// 再次调用应该返回同一个实例
	logger2 := GetDefaultLogger()
	assert.Equal(t, logger, logger2)
}

// TestSetDefaultLogger 测试设置默认日志器
func TestSetDefaultLogger(t *testing.T) {
	customLogger := NewLogger()
	SetDefaultLogger(customLogger)

	logger := GetDefaultLogger()
	assert.Equal(t, customLogger, logger)
}

// TestNewLogger 测试创建新的日志器
func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	assert.NotNil(t, logger)
}

// TestLogger_SetOutput 测试设置日志输出
func TestLogger_SetOutput(t *testing.T) {
	logger := NewLogger()
	buf := &bytes.Buffer{}

	result := logger.SetOutput(buf)
	assert.NotNil(t, result)
	assert.Equal(t, logger, result)
}

// TestLogger_LogMode 测试日志级别设置
func TestLogger_LogMode(t *testing.T) {
	tests := []struct {
		name     string
		logLevel gormLog.LogLevel
	}{
		{
			name:     "silent mode",
			logLevel: gormLog.Silent,
		},
		{
			name:     "error mode",
			logLevel: gormLog.Error,
		},
		{
			name:     "warn mode",
			logLevel: gormLog.Warn,
		},
		{
			name:     "info mode",
			logLevel: gormLog.Info,
		},
		{
			name:     "default mode",
			logLevel: gormLog.LogLevel(999), // 未知级别，应该使用默认
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger()
			result := logger.LogMode(tt.logLevel)
			assert.NotNil(t, result)
		})
	}
}

// TestLogger_Info 测试 Info 日志
func TestLogger_Info(t *testing.T) {
	logger := NewLogger()
	buf := &bytes.Buffer{}
	logger.SetOutput(buf)

	ctx := context.Background()
	logger.Info(ctx, "test info message: %s", "value")

	// 验证日志已写入
	assert.NotEmpty(t, buf.String())
}

// TestLogger_Warn 测试 Warn 日志
func TestLogger_Warn(t *testing.T) {
	logger := NewLogger()
	buf := &bytes.Buffer{}
	logger.SetOutput(buf)

	ctx := context.Background()
	logger.Warn(ctx, "test warn message: %s", "value")

	// 验证日志已写入
	assert.NotEmpty(t, buf.String())
}

// TestLogger_Error 测试 Error 日志
func TestLogger_Error(t *testing.T) {
	logger := NewLogger()
	buf := &bytes.Buffer{}
	logger.SetOutput(buf)

	ctx := context.Background()
	logger.Error(ctx, "test error message: %s", "value")

	// 验证日志已写入
	assert.NotEmpty(t, buf.String())
}

// TestLogger_Trace 测试 Trace 日志
func TestLogger_Trace(t *testing.T) {
	logger := NewLogger()
	buf := &bytes.Buffer{}
	logger.SetOutput(buf)

	ctx := context.Background()
	begin := time.Now()

	fc := func() (sql string, rowsAffected int64) {
		return "SELECT * FROM users WHERE id = 1", 1
	}

	t.Run("trace without error", func(t *testing.T) {
		buf.Reset()
		logger.Trace(ctx, begin, fc, nil)
		assert.NotEmpty(t, buf.String())
		assert.Contains(t, buf.String(), "SELECT * FROM users WHERE id = 1")
		assert.Contains(t, buf.String(), "[1 rows]")
	})

	t.Run("trace with error", func(t *testing.T) {
		buf.Reset()
		err := errors.New("test error")
		logger.Trace(ctx, begin, fc, err)
		assert.NotEmpty(t, buf.String())
		assert.Contains(t, buf.String(), "SELECT * FROM users WHERE id = 1")
	})
}

// TestLogger_Log 测试 Log 方法
func TestLogger_Log(t *testing.T) {
	logger := NewLogger()
	buf := &bytes.Buffer{}
	logger.SetOutput(buf)

	begin := time.Now()

	fc := func() (sql string, rowsAffected int64) {
		return "INSERT INTO users (name) VALUES ('test')", 1
	}

	t.Run("log without error", func(t *testing.T) {
		buf.Reset()
		logger.Log(4, begin, fc, nil)
		output := buf.String()
		assert.NotEmpty(t, output)
		assert.Contains(t, output, "INSERT INTO users (name) VALUES ('test')")
		assert.Contains(t, output, "[1 rows]")
	})

	t.Run("log with error", func(t *testing.T) {
		buf.Reset()
		err := errors.New("database error")
		logger.Log(4, begin, fc, err)
		output := buf.String()
		assert.NotEmpty(t, output)
		assert.Contains(t, output, "INSERT INTO users (name) VALUES ('test')")
	})

	t.Run("log with multiline SQL", func(t *testing.T) {
		buf.Reset()
		fcMultiline := func() (sql string, rowsAffected int64) {
			return "SELECT *\nFROM users\nWHERE id = 1", 1
		}
		logger.Log(4, begin, fcMultiline, nil)
		output := buf.String()
		assert.NotEmpty(t, output)
		// SQL 中的换行符应该被替换为空格
		// 但日志输出本身可能包含换行符，所以我们检查 SQL 部分
		assert.Contains(t, output, "SELECT * FROM users WHERE id = 1")
	})
}

// TestMysqlLogger_Print 测试 MySQL 日志器
func TestMysqlLogger_Print(t *testing.T) {
	logger := &mysqlLogger{}

	// Print 方法应该不会 panic
	assert.NotPanics(t, func() {
		logger.Print("test message")
	})

	assert.NotPanics(t, func() {
		logger.Print("test", "multiple", "args")
	})

	assert.NotPanics(t, func() {
		logger.Print()
	})
}
