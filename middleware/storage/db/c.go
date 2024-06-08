//go:build cgo
// +build cgo

package db

import (
	"fmt"
	"github.com/lazygophers/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"path/filepath"
)

func newSqlite(c *Config) gorm.Dialector {
	log.Infof("sqlite3://%s.db", filepath.ToSlash(filepath.Join(c.Address, c.Name)))
	return sqlite.Open(fmt.Sprintf("%s.db?_vacuum=2&_journal=delete&_locking_mode=exclusive&mode=rwc&_sync=3&_timeout=9999999", filepath.ToSlash(filepath.Join(c.Address, c.Name))))
}
