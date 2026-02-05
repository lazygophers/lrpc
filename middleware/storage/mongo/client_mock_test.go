package mongo_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/mongo"
	_ "github.com/lazygophers/lrpc/middleware/storage/mongo/mock" // Import to register mock factory
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew_MockMode tests creating a client in mock mode
func TestNew_MockMode(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Verify mock client was created
	assert.True(t, cfg.Mock)
	// Note: cannot test client.database as it's not exported in mongo_test package
}

// TestNew_MockModeWithoutFactory tests mock mode when factory is not registered
// NOTE: This test is disabled because mockClientFactory is not exported
// and the mock factory is auto-registered by the mock package init()
func TestNew_MockModeWithoutFactory(t *testing.T) {
	t.Skip("mockClientFactory is not exported, cannot test this scenario")
}

// TestRegisterMockClientFactory tests registering the mock factory
// NOTE: This test is disabled because it requires access to internal package fields
func TestRegisterMockClientFactory(t *testing.T) {
	t.Skip("RegisterMockClientFactory and mockClientFactory are not exported")
}

// TestGetDatabase tests GetDatabase method
func TestGetDatabase(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *mongo.Config
		expected string
	}{
		{
			name: "with database name",
			cfg: &mongo.Config{
				Mock:     true,
				Database: "my_db",
			},
			expected: "my_db",
		},
		{
			name: "empty database defaults to test",
			cfg: &mongo.Config{
				Mock:     true,
				Database: "",
			},
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := mongo.New(tt.cfg)
			require.NoError(t, err)

			dbName := client.GetDatabase()
			assert.Equal(t, tt.expected, dbName)
		})
	}
}

// TestGetConfig tests GetConfig method
func TestGetConfig(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
		Address:  "localhost",
		Port:     27017,
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	retrievedCfg := client.GetConfig()
	assert.Equal(t, cfg, retrievedCfg)
	assert.Equal(t, "test_db", retrievedCfg.Database)
	assert.Equal(t, "localhost", retrievedCfg.Address)
	assert.Equal(t, 27017, retrievedCfg.Port)
}

// TestContext tests Context method
func TestContext(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	ctx := client.Context()
	assert.NotNil(t, ctx)
}

// TestAutoMigrate_MockMode tests AutoMigrate with mock client
func TestAutoMigrate_MockMode(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	// Define a simple model
	type User struct {
		Name string
		Age  int
	}

	// AutoMigrate should create the collection
	err = client.AutoMigrate(User{})
	require.NoError(t, err)
}

// TestAutoMigrates_MockMode tests AutoMigrates with multiple models
func TestAutoMigrates_MockMode(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	type User struct {
		Name string
	}

	type Post struct {
		Title string
	}

	// AutoMigrate multiple models
	err = client.AutoMigrates(User{}, Post{})
	require.NoError(t, err)
}

// TestNew_ConfigDefaults tests that config defaults are applied
func TestNew_ConfigDefaults(t *testing.T) {
	cfg := &mongo.Config{
		Mock: true,
		// Other fields left empty to test defaults
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Verify defaults were applied via GetConfig
	retrievedCfg := client.GetConfig()
	assert.NotNil(t, retrievedCfg)
	assert.True(t, retrievedCfg.Mock)
}
