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

	Mock bool `yaml:"mock,omitempty" json:"mock,omitempty"`
}

// apply applies default values to the configuration
func (p *Config) apply() {
	if p.Logger == nil {
		p.Logger = GetDefaultLogger()
	}

	if p.Address == "" {
		p.Address = "127.0.0.1"
	}

	if p.Port == 0 {
		p.Port = 27017
	}

	if p.ConnectTimeout <= 0 {
		p.ConnectTimeout = 10 * time.Second
	}

	if p.ContextTimeout <= 0 {
		p.ContextTimeout = 30 * time.Second
	}

	if p.MaxPoolSize == 0 {
		p.MaxPoolSize = 100
	}

	if p.MinPoolSize == 0 {
		p.MinPoolSize = 10
	}

	if p.MaxConnIdleTime == 0 {
		p.MaxConnIdleTime = 5 * time.Minute
	}
}

// buildURI builds MongoDB connection URI from config
func (p *Config) buildURI() string {
	uri := "mongodb://"

	// Add credentials if provided
	if p.Username != "" && p.Password != "" {
		uri += fmt.Sprintf("%s:%s@", p.Username, p.Password)
	}

	// Add host and port
	uri += fmt.Sprintf("%s:%d", p.Address, p.Port)

	// Build query parameters
	var params []string

	if p.ReplicaSet != "" {
		params = append(params, fmt.Sprintf("replicaSet=%s", p.ReplicaSet))
	}

	if p.AuthSource != "" {
		params = append(params, fmt.Sprintf("authSource=%s", p.AuthSource))
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
func (p *Config) BuildClientOpts() *options.ClientOptions {
	opts := options.Client().
		ApplyURI(p.buildURI()).
		SetMaxConnIdleTime(p.MaxConnIdleTime).
		SetMaxPoolSize(p.MaxPoolSize).
		SetMinPoolSize(p.MinPoolSize)

	return opts
}
