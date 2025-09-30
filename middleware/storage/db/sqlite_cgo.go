//go:build cgo
// +build cgo

package db

import (
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// hasCGOSupport indicates whether CGO is available
const hasCGOSupport = true

// newSqliteDialector creates a SQLite dialector with CGO support (SQLCipher encryption)
func newSqliteDialector(dsn string) gorm.Dialector {
	return sqlite.Open(dsn)
}