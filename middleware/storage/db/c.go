//go:build cgo && !withoutc

package db

import (
	"fmt"
	"github.com/lazygophers/log"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"path/filepath"
)

func newSqlite(c *Config) gorm.Dialector {
	log.Infof("sqlite3://%s.db", filepath.ToSlash(filepath.Join(c.Address, c.Name)))

	dsn := fmt.Sprintf("%s.db?_vacuum=2&_journal=delete&_locking_mode=exclusive&mode=rwc&_sync=3&_timeout=9999999&_auth_user=%s&_auth_pass=%s", filepath.ToSlash(filepath.Join(c.Address, c.Name)), c.Username, c.Password)

	if len(c.Extras) > 0 {
		for key, value := range c.Extras {
			dsn = fmt.Sprintf("%s&%s=%s", dsn, key, value)
		}
	}

	return sqlite.Open(dsn)
}
