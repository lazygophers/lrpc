package db

import (
	"reflect"
)

// TestHelpers - Functions to expose internal functionality for testing purposes

// MysqlLogger exposes the internal mysqlLogger type for testing
type MysqlLogger = mysqlLogger

// CallPrint exposes the internal mysqlLogger Print function for testing
func CallPrint(v ...interface{}) {
	logger := &mysqlLogger{}
	logger.Print(v...)
}

// CallGetTableName exposes the internal getTableName function for testing
func CallGetTableName(elem reflect.Type) string {
	return getTableName(elem)
}

// CallDecode exposes the internal decode function for testing
func CallDecode(field reflect.Value, col []byte) error {
	return decode(field, col)
}

// CallApply exposes the internal apply function through config processing
// This is already exposed through New() but we can create a direct wrapper if needed
func CallApply(config *Config) error {
	// The apply function is called internally by New()
	// We can trigger it by creating a client
	_, err := New(config)
	return err
}
