package mongo

import (
	"context"
	"fmt"

	"github.com/kamva/mgm/v3"
	"github.com/lazygophers/log"
	"go.mongodb.org/mongo-driver/mongo"
)

// Client represents a MongoDB client wrapper using MGM
type Client struct {
	cfg      *Config
	client   *mongo.Client
	database string
}

// New creates a new MongoDB client with the given configuration
func New(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	// Apply defaults
	cfg.apply()

	// Build MongoDB client options
	opts := cfg.BuildClientOpts()

	// Initialize MGM with configuration
	err := mgm.SetDefaultConfig(&mgm.Config{}, cfg.Database, opts)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Get the MongoDB client from MGM to verify connection
	_, mongoClient, _, err := mgm.DefaultConfigs()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Ping to verify connection
	err = mongoClient.Ping(context.Background(), nil)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	c := &Client{
		cfg:      cfg,
		client:   mongoClient,
		database: cfg.Database,
	}

	log.Infof("successfully connected to MongoDB: %s:%d", cfg.Address, cfg.Port)

	return c, nil
}

// Ping checks the connection to MongoDB
func (c *Client) Ping() error {
	_, client, _, err := mgm.DefaultConfigs()
	if err != nil {
		return err
	}
	return client.Ping(context.Background(), nil)
}

// Close closes the MongoDB client connection
func (c *Client) Close() error {
	_, client, _, err := mgm.DefaultConfigs()
	if err != nil {
		return err
	}
	if client != nil {
		err := client.Disconnect(context.Background())
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}
	return nil
}

// GetConfig returns the client configuration
func (c *Client) GetConfig() *Config {
	return c.cfg
}

// Context returns the operation context
func (c *Client) Context() context.Context {
	return context.Background()
}

// GetDatabase returns the database name
func (c *Client) GetDatabase() string {
	if c.cfg.Database == "" {
		return "test"
	}
	return c.cfg.Database
}

// Health checks the health of the connection
func (c *Client) Health() error {
	err := c.Ping()
	if err != nil {
		log.Errorf("err:%v", err)
		return fmt.Errorf("health check failed: %w", err)
	}
	return nil
}
