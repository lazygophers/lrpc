//go:build !cgo
// +build !cgo

package db

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// hasCGOSupport indicates whether CGO is available
const hasCGOSupport = false

// newSqliteDialector returns glebarez/sqlite when CGO is not available
func newSqliteDialector(dsn string) gorm.Dialector {
	return sqlite.Open(dsn)
}