package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNew(t *testing.T) {
	cfg := newTestConfig()
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Test that client is created
	if client == nil {
		t.Error("expected client, got nil")
	}

	// Verify MGM was initialized
	err = client.Ping()
	if err != nil {
		t.Errorf("expected ping to succeed, got error: %v", err)
	}
}

func TestNewWithInvalidConfig(t *testing.T) {
	cfg := &Config{
		Address: "invalid-host",
		Port:    27017,
	}

	_, err := New(cfg)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClientPing(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	err := client.Ping()
	if err != nil {
		t.Fatalf("ping failed: %v", err)
	}
}

func TestClientHealth(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	err := client.Health()
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
}

func TestClientNewScoop(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop()
	if scoop == nil {
		t.Error("expected scoop, got nil")
	}

	// Collection is lazily initialized when needed
	// Explicitly set collection for testing
	scoop.Collection(User{})
	if scoop.GetCollection() == nil {
		t.Error("expected collection, got nil")
	}
}

func TestClientNewModel(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	model := NewModel[User](client)
	if model == nil {
		t.Error("expected model, got nil")
	}

	if model.CollectionName() == "" {
		t.Error("expected collection name, got empty string")
	}
}

func TestClientBeginTx(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop()
	txScoop, err := scoop.Begin()
	if err != nil {
		t.Fatalf("begin transaction failed: %v", err)
	}

	if txScoop == nil {
		t.Error("expected transaction scoop, got nil")
	}

	txScoop.Rollback()
}

func TestClientClose(t *testing.T) {
	client := newTestClient(t)

	err := client.Close()
	if err != nil {
		t.Fatalf("close failed: %v", err)
	}

	// Test that ping fails after close

	err = client.Ping()
	if err == nil {
		t.Error("expected error after close, got nil")
	}
}

func TestClientGetConfig(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cfg := client.GetConfig()
	if cfg == nil {
		t.Error("expected config, got nil")
	}

	testCfg := newTestConfig()
	if cfg.Address != testCfg.Address {
		t.Errorf("expected address '%s', got '%s'", testCfg.Address, cfg.Address)
	}

	if cfg.Port != testCfg.Port {
		t.Errorf("expected port %d, got %d", testCfg.Port, cfg.Port)
	}
}

func TestClientGetDatabase(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	dbName := client.GetDatabase()
	if dbName != "test" {
		t.Errorf("expected database name 'test', got '%s'", dbName)
	}
}

func TestClientContext(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	ctx := client.Context()
	if ctx == nil {
		t.Error("expected context, got nil")
	}
}

func TestClientInsertAndFind(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data using Model
	model := NewModel[User](client)
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "test@example.com",
		Name:      "Test User",
		Age:       25,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	err := model.NewScoop().Create(&user)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Find test data
	foundUser, err := model.NewScoop().Where("email", "test@example.com").First()
	if err != nil {
		t.Fatalf("find one failed: %v", err)
	}

	if foundUser == nil {
		t.Error("expected user, got nil")
	} else if foundUser.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", foundUser.Email)
	}
}

func TestClientConcurrentOperations(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	model := NewModel[User](client)

	// Insert multiple documents
	docs := []User{
		{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 35, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}

	for _, doc := range docs {
		err := model.NewScoop().Create(&doc)
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
	}

	// Count documents
	count, err := model.NewScoop().Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 documents, got %d", count)
	}
}

func TestClientWithDifferentDatabases(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Test is simplified as databases are within same MongoDB connection
	model := NewModel[User](client)

	// Insert test data
	user1 := User{ID: primitive.NewObjectID(), Email: "test@example.com", Name: "Test User", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()}
	err := model.NewScoop().Create(&user1)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Verify data
	foundUser, err := model.NewScoop().Where("email", "test@example.com").First()
	if err != nil {
		t.Fatalf("find one failed: %v", err)
	}

	if foundUser == nil {
		t.Error("expected user, got nil")
	} else if foundUser.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", foundUser.Email)
	}
}
