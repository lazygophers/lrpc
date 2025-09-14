//go:build cgo && withoutc

package db

import (
	"fmt"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"github.com/lazygophers/log"
	"gorm.io/gorm"
)

func newSqlite(c *Config) gorm.Dialector {
	log.Infof("sqlite3-pure://%s.db", filepath.ToSlash(filepath.Join(c.Address, c.Name)))

	dsn := fmt.Sprintf("%s.db?_vacuum=2&_journal=delete&_locking_mode=exclusive&mode=rwc&_sync=3&_timeout=9999999", filepath.ToSlash(filepath.Join(c.Address, c.Name)))

	if len(c.Extras) > 0 {
		for key, value := range c.Extras {
			dsn = fmt.Sprintf("%s&%s=%s", dsn, key, value)
		}
	}

	return sqlite.Open(dsn)
}
