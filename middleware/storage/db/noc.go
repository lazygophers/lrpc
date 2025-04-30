//go:build !cgo
// +build !cgo

package db

import (
	"github.com/glebarez/sqlite"
	"github.com/lazygophers/log"
	"gorm.io/gorm"
	"path/filepath"
)

func newSqlite(c *Config) gorm.Dialector {
	log.Infof("sqlite3://%s.db", filepath.ToSlash(filepath.Join(c.Address, c.Name)))

	return sqlite.Open(c.DSN())
}
