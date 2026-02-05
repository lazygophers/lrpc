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
	client   MongoClient
	database string
	db       MongoDatabase
}

// mockClientFactory 用于创建 Mock client 的工厂函数
// 通过 init() 或外部注册来设置，避免直接导入 mock 包造成循环依赖
var mockClientFactory func() MongoClient

// RegisterMockClientFactory 注册 Mock client 工厂函数
// 由 mock 包在 init() 中调用，实现自动注册
func RegisterMockClientFactory(factory func() MongoClient) {
	mockClientFactory = factory
	log.Debugf("MockClientFactory registered successfully")
}

// New creates a new MongoDB client with the given configuration
func New(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	// Apply defaults
	cfg.apply()

	c := &Client{
		cfg:      cfg,
		database: cfg.Database,
	}

	// Mock mode
	if cfg.Mock {
		if mockClientFactory == nil {
			return nil, fmt.Errorf("mock mode enabled but mockClientFactory not registered, please import _ \"github.com/lazygophers/lrpc/middleware/storage/mongo/mock\"")
		}
		mockClient := mockClientFactory()
		c.client = mockClient
		c.db = mockClient.Database(cfg.Database)
		log.Infof("MongoDB Mock mode enabled for database: %s", cfg.Database)
		return c, nil
	}

	// Real mode - original logic
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

	db := mongoClient.Database(cfg.Database)
	log.Infof("MongoDB database instance for '%s': %v", cfg.Database, db)

	// Wrap real MongoDB client with interface implementation
	realClient := NewRealClient(mongoClient)
	c.client = realClient
	c.db = realClient.Database(cfg.Database)

	log.Infof("successfully connected to MongoDB: %s:%d, database: %s", cfg.Address, cfg.Port, cfg.Database)

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

// AutoMigrates ensures that all provided models have their corresponding collections in MongoDB
// It iterates through each model and calls AutoMigrate for each one
func (c *Client) AutoMigrates(models ...interface{}) (err error) {
	for _, model := range models {
		err = c.AutoMigrate(model)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}

	return nil
}

// AutoMigrate ensures that the collection for a given model exists in MongoDB
// It retrieves the collection name from the model and creates the collection
// if it doesn't already exist
func (c *Client) AutoMigrate(model interface{}) (err error) {
	// Get collection name using reflection
	collectionName := getCollectionName(model)
	if collectionName == "" {
		return fmt.Errorf("unable to determine collection name for model type %T", model)
	}

	log.Infof("auto migrate collection %s", collectionName)

	// Check if collection exists
	collections, err := c.db.ListCollectionNames(context.Background(), nil)

	collectionExists := false
	if err == nil {
		for _, name := range collections {
			if name == collectionName {
				collectionExists = true
				break
			}
		}
	}

	// If collection doesn't exist or couldn't verify, create it
	if !collectionExists {
		err = c.db.CreateCollection(context.Background(), collectionName)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
		log.Infof("created collection %s", collectionName)
	}

	// Try to create indexes if the model implements an index interface
	// This is optional and can be extended in the future if needed
	if indexer, ok := model.(interface{ Indexes() []mongo.IndexModel }); ok {
		indexes := indexer.Indexes()
		if len(indexes) > 0 {
			coll := c.db.Collection(collectionName)
			_, err = coll.Indexes().CreateMany(context.Background(), indexes)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
			log.Infof("created indexes for collection %s", collectionName)
		}
	}

	return nil
}
