//go:build !cgo
// +build !cgo

package db

import (
	"gorm.io/gorm"
)

// newSqliteCGODialector is a no-op for no-CGO builds
// This will cause a compile error if SqliteCGO is used without CGO
func newSqliteCGODialector(dsn string) gorm.Dialector {
	panic("sqlite-cgo requires CGO to be enabled. Please rebuild with CGO_ENABLED=1 or use 'sqlite' type instead")
}