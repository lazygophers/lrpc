//go:build sqlite_cgo
// +build sqlite_cgo

package db

import (
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newSqliteCGODialector creates a SQLite dialector with CGO support (SQLCipher encryption)
func newSqliteCGODialector(dsn string) gorm.Dialector {
	return sqlite.Open(dsn)
}