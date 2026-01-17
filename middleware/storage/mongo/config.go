package mongo

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config represents MongoDB configuration
type Config struct {
	// MongoDB server address, default 127.0.0.1
	Address string `yaml:"address,omitempty" json:"address,omitempty"`

	// MongoDB server port, default 27017
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Database name, default test
	Database string `yaml:"database,omitempty" json:"database,omitempty"`

	// MongoDB username for authentication
	Username string `yaml:"username,omitempty" json:"username,omitempty"`

	// MongoDB password for authentication
	Password string `yaml:"password,omitempty" json:"password,omitempty"`

	// Authentication database (default: admin)
	AuthSource string `yaml:"auth_source,omitempty" json:"auth_source,omitempty"`

	// Replica set name (optional)
	ReplicaSet string `yaml:"replica_set,omitempty" json:"replica_set,omitempty"`

	// Debug mode - enables verbose logging
	Debug bool `yaml:"debug,omitempty" json:"debug,omitempty"`

	// Connection timeout, default 10 seconds
	ConnectTimeout time.Duration `yaml:"connect_timeout,omitempty" json:"connect_timeout,omitempty"`

	// Context timeout for operations, default 30 seconds
	ContextTimeout time.Duration `yaml:"context_timeout,omitempty" json:"context_timeout,omitempty"`

	// Max connection pool size, default 100
	MaxPoolSize uint64 `yaml:"max_pool_size,omitempty" json:"max_pool_size,omitempty"`

	// Min connection pool size, default 10
	MinPoolSize uint64 `yaml:"min_pool_size,omitempty" json:"min_pool_size,omitempty"`

	// Max connection idle time, default 5 minutes
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time,omitempty" json:"max_conn_idle_time,omitempty"`

	// Logger interface for custom logging (for SQL output)
	Logger Logger `json:"-" yaml:"-"`
}

// apply applies default values to the configuration
func (c *Config) apply() {
	if c.Logger == nil {
		c.Logger = GetDefaultLogger()
	}

	if c.Address == "" {
		c.Address = "127.0.0.1"
	}

	if c.Port == 0 {
		c.Port = 27017
	}

	if c.ConnectTimeout <= 0 {
		c.ConnectTimeout = 10 * time.Second
	}

	if c.ContextTimeout <= 0 {
		c.ContextTimeout = 30 * time.Second
	}

	if c.MaxPoolSize == 0 {
		c.MaxPoolSize = 100
	}

	if c.MinPoolSize == 0 {
		c.MinPoolSize = 10
	}

	if c.MaxConnIdleTime == 0 {
		c.MaxConnIdleTime = 5 * time.Minute
	}
}

// buildURI builds MongoDB connection URI from config
func (c *Config) buildURI() string {
	uri := "mongodb://"

	// Add credentials if provided
	if c.Username != "" && c.Password != "" {
		uri += fmt.Sprintf("%s:%s@", c.Username, c.Password)
	}

	// Add host and port
	uri += fmt.Sprintf("%s:%d", c.Address, c.Port)

	// Build query parameters
	var params []string

	if c.ReplicaSet != "" {
		params = append(params, fmt.Sprintf("replicaSet=%s", c.ReplicaSet))
	}

	if c.AuthSource != "" {
		params = append(params, fmt.Sprintf("authSource=%s", c.AuthSource))
	}

	if len(params) > 0 {
		uri += "/?" + params[0]
		for _, param := range params[1:] {
			uri += "&" + param
		}
	} else {
		uri += "/"
	}

	return uri
}

// BuildClientOpts builds MongoDB client options from config
func (c *Config) BuildClientOpts() *options.ClientOptions {
	opts := options.Client().
		ApplyURI(c.buildURI()).
		SetMaxConnIdleTime(c.MaxConnIdleTime).
		SetMaxPoolSize(c.MaxPoolSize).
		SetMinPoolSize(c.MinPoolSize)

	return opts
}
