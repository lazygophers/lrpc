package db

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/lazygophers/utils/app"
	"gorm.io/gorm/logger"
)

const (
	Sqlite     = "sqlite"
	MySQL      = "mysql"
	Postgres   = "postgres"
	ClickHouse = "clickhouse"
	TiDB       = "tidb"
	GaussDB    = "gaussdb"
)

type Config struct {
	// Database type, support sqlite, mysql, postgres, clickhouse, tidb, gaussdb, default sqlite
	// sqlite: sqlite|sqlite3
	// mysql: mysql
	// postgres: postgres|pg|postgresql|pgsql
	// clickhouse: clickhouse|ch
	// tidb: tidb (MySQL-compatible)
	// gaussdb: gaussdb (PostgreSQL-compatible)
	Type string `yaml:"type,omitempty" json:"type,omitempty"`

	// Database debug, default false
	Debug bool `yaml:"debug,omitempty" json:"debug,omitempty"`

	// Database address
	// sqlite: full filepath, default exec path
	// mysql: database address, default 127.0.0.1
	// postgres: database address, default 127.0.0.1
	// clickhouse: database address, default 127.0.0.1
	// tidb: database address, default 127.0.0.1
	// gaussdb: database address, default 127.0.0.1
	Address string `yaml:"address,omitempty" json:"address,omitempty"`

	// Database port
	// sqlite: empty
	// mysql: database port, default 3306
	// postgres: database port, default 5432
	// clickhouse: database port, default 9000 (native) or 8123 (http)
	// tidb: database port, default 4000
	// gaussdb: database port, default 5432
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Database name
	// sqlite: database file name, default ice.db
	// mysql: database name, default ice
	// postgres: database name, default ice
	// clickhouse: database name, default default
	// tidb: database name, default ice
	// gaussdb: database name, default ice
	Name string `yaml:"name,omitempty" json:"name,omitempty"`

	// Database username
	// sqlite: empty
	// mysql: database username
	// postgres: database username
	// sqlserver: database username
	Username string `yaml:"username,omitempty" json:"username,omitempty"`

	// Database password
	// sqlite: empty
	// mysql: database password
	// postgres: database password
	// sqlserver: database password
	Password string `yaml:"password,omitempty" json:"password,omitempty"`

	Extras map[string]string `yaml:"extras,omitempty" json:"extras,omitempty"`

	Logger logger.Interface `json:"-" yaml:"-"`
}

func (c *Config) apply() {
	if c.Type == "" {
		c.Type = Sqlite
	}

	switch c.Type {
	case Sqlite, "sqlite3":
		c.Type = Sqlite

		if c.Address == "" {
			c.Address, _ = os.Executable()
		}

		if !strings.HasPrefix(c.Address, "file:") {
			c.Address = "file:" + c.Address
		}

		if c.Name == "" {
			c.Name = app.Name + ".db"
		}

	case MySQL:
		if c.Address == "" {
			c.Address = "127.0.0.1"
		}

		if c.Port == 0 {
			c.Port = 3306
		}

		if c.Name == "" {
			c.Name = app.Name
		}

	case "postgres", "pg", "postgresql", "pgsql":
		c.Type = Postgres

		if c.Address == "" {
			c.Address = "127.0.0.1"
		}

		if c.Port == 0 {
			c.Port = 5432
		}

		if c.Name == "" {
			c.Name = app.Name
		}

	case "clickhouse", "ch":
		c.Type = ClickHouse

		if c.Address == "" {
			c.Address = "127.0.0.1"
		}

		if c.Port == 0 {
			c.Port = 9000 // Native protocol port
		}

		if c.Name == "" {
			c.Name = "default"
		}

	case "tidb":
		c.Type = TiDB

		if c.Address == "" {
			c.Address = "127.0.0.1"
		}

		if c.Port == 0 {
			c.Port = 4000
		}

		if c.Name == "" {
			c.Name = app.Name
		}

	case "gaussdb":
		c.Type = GaussDB

		if c.Address == "" {
			c.Address = "127.0.0.1"
		}

		if c.Port == 0 {
			c.Port = 5432
		}

		if c.Name == "" {
			c.Name = app.Name
		}

	case "sqlserver", "mssql":
		c.Type = "sqlserver"

		if c.Address == "" {
			c.Address = "127.0.0.1"
		}

		if c.Port == 0 {
			c.Port = 1433
		}

		if c.Name == "" {
			c.Name = app.Name
		}
	}
}

func (c *Config) DSN() string {
	switch c.Type {
	case Sqlite:
		query := &url.Values{}

		dsn := fmt.Sprintf("%s.db", filepath.ToSlash(filepath.Join(c.Address, c.Name)))

		query.Set("_vacuum", "2")
		query.Set("_journal", "delete")
		query.Set("_locking_mode", "exclusive")
		query.Set("mode", "rwc")
		query.Set("_sync", "3")
		query.Set("_timeout", "9999999")

		if c.Username != "" && c.Password != "" {
			query.Set("_auth", "1")

			if c.Username != "" {
				query.Set("_auth_user", c.Username)
			}

			if c.Password != "" {
				query.Set("_auth_pass", c.Password)
			}

			query.Set("_auth_crypt", "sha512")
			query.Set("_auth_salt", app.Name)
		}

		for key, value := range c.Extras {
			query.Set(key, value)
		}

		return dsn + "?" + query.Encode()

	case MySQL:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.Username, c.Password, c.Address, c.Port, c.Name)

	case Postgres:
		query := &url.Values{}
		query.Set("sslmode", "disable")
		query.Set("TimeZone", "Asia/Shanghai")

		for key, value := range c.Extras {
			query.Set(key, value)
		}

		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s %s",
			c.Address, c.Port, c.Username, c.Password, c.Name, query.Encode())

	case ClickHouse:
		// ClickHouse native protocol DSN
		// clickhouse://username:password@host:port/database?dial_timeout=10s&max_execution_time=60
		query := &url.Values{}
		query.Set("dial_timeout", "10s")
		query.Set("max_execution_time", "60")
		query.Set("read_timeout", "30s")
		query.Set("write_timeout", "30s")

		for key, value := range c.Extras {
			query.Set(key, value)
		}

		if c.Username != "" && c.Password != "" {
			return fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s?%s",
				c.Username, c.Password, c.Address, c.Port, c.Name, query.Encode())
		}
		return fmt.Sprintf("clickhouse://%s:%d/%s?%s",
			c.Address, c.Port, c.Name, query.Encode())

	case TiDB:
		// TiDB is MySQL-compatible, use same DSN format as MySQL
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.Username, c.Password, c.Address, c.Port, c.Name)

	case GaussDB:
		// GaussDB is PostgreSQL-compatible, use same DSN format as PostgreSQL
		query := &url.Values{}
		query.Set("sslmode", "disable")
		query.Set("TimeZone", "Asia/Shanghai")

		for key, value := range c.Extras {
			query.Set(key, value)
		}

		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s %s",
			c.Address, c.Port, c.Username, c.Password, c.Name, query.Encode())

	default:
		return ""
	}
}
