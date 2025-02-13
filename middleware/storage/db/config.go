package db

import (
	"github.com/lazygophers/utils/app"
	"gorm.io/gorm/logger"
	"os"
)

const (
	Sqlite = "sqlite"
	MySQL  = "mysql"
)

type Config struct {
	// Database type, support sqlite, mysql, postgres, sqlserver, default sqlite
	// sqlite: sqlite|sqlite3
	// mysql: mysql
	// postgres: postgres|pg|postgresql|pgsql
	// sqlserver: sqlserver|mssql
	Type string `yaml:"type,omitempty" json:"type,omitempty"`

	// Database debug, default false
	Debug bool `yaml:"debug,omitempty" json:"debug,omitempty"`

	// Database address
	// sqlite: full filepath, default exec path
	// mysql: database address, default 127.0.0.1
	// postgres: database address, default 127.0.0.1
	// sqlserver: database address, default 127.0.0.1
	Address string `yaml:"address,omitempty" json:"address,omitempty"`

	// Database port
	// sqlite: empty
	// mysql: database port, default 3306
	// postgres: database port, default 5432
	// sqlserver: database port, default 1433
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Database name
	// sqlite: database file name, default ice.db
	// mysql: database name, default ice
	// postgres: database name, default ice
	// sqlserver: database name, default ice
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
		c.Type = "sqlite"
	}

	switch c.Type {
	case "sqlite", "sqlite3":
		c.Type = "sqlite"

		if c.Address == "" {
			c.Address, _ = os.Executable()
		}

		if c.Name == "" {
			c.Name = app.Name + ".db"
		}

	case "mysql":
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
		c.Type = "postgres"

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
