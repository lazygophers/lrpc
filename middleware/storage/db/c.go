//go:build cgo && !withoutc

package db

import (
	"github.com/lazygophers/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	//_ "github.com/mutecomm/go-sqlcipher/v4"
)

func newSqlite(c *Config) gorm.Dialector {
	log.Infof("sqlite3://%s.db", filepath.ToSlash(filepath.Join(c.Address, c.Name)))

	return sqlite.Open(c.DSN())
}
