//go:build !sqlite_cgo
// +build !sqlite_cgo

package db

import (
	"gorm.io/gorm"
)

// newSqliteCGODialector is a no-op for builds without sqlite_cgo tag
// This will panic if SqliteCGO type is used without the proper build tag
func newSqliteCGODialector(dsn string) gorm.Dialector {
	panic("sqlite-cgo requires CGO to be enabled. Please rebuild with: go build -tags sqlite_cgo")
}