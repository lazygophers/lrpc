package mongo

import (
	"testing"
)

// TestBuildURIWithCredentials tests URI building with username and password
func TestBuildURIWithCredentials(t *testing.T) {
	cfg := &Config{
		Address:  "localhost",
		Port:     27017,
		Username: "admin",
		Password: "password",
	}
	uri := cfg.buildURI()
	
	if uri != "mongodb://admin:password@localhost:27017/" {
		t.Errorf("expected 'mongodb://admin:password@localhost:27017/', got '%s'", uri)
	}
}

// TestBuildURIWithoutCredentials tests URI building without username and password
func TestBuildURIWithoutCredentials(t *testing.T) {
	cfg := &Config{
		Address: "localhost",
		Port:    27017,
	}
	uri := cfg.buildURI()
	
	if uri != "mongodb://localhost:27017/" {
		t.Errorf("expected 'mongodb://localhost:27017/', got '%s'", uri)
	}
}

// TestBuildURIWithPartialCredentials tests URI building with only password
func TestBuildURIWithPartialCredentials(t *testing.T) {
	cfg := &Config{
		Address:  "localhost",
		Port:     27017,
		Username: "admin",
		// No password
	}
	uri := cfg.buildURI()
	
	// Should not include credentials if password is missing
	if uri != "mongodb://localhost:27017/" {
		t.Errorf("expected 'mongodb://localhost:27017/', got '%s'", uri)
	}
}

// TestBuildURIWithReplicaSet tests URI building with replica set
func TestBuildURIWithReplicaSet(t *testing.T) {
	cfg := &Config{
		Address:   "localhost",
		Port:      27017,
		ReplicaSet: "rs0",
	}
	uri := cfg.buildURI()
	
	if uri != "mongodb://localhost:27017/?replicaSet=rs0" {
		t.Errorf("expected 'mongodb://localhost:27017/?replicaSet=rs0', got '%s'", uri)
	}
}

// TestBuildURIWithAuthSource tests URI building with auth source
func TestBuildURIWithAuthSource(t *testing.T) {
	cfg := &Config{
		Address:    "localhost",
		Port:       27017,
		AuthSource: "admin",
	}
	uri := cfg.buildURI()
	
	if uri != "mongodb://localhost:27017/?authSource=admin" {
		t.Errorf("expected 'mongodb://localhost:27017/?authSource=admin', got '%s'", uri)
	}
}

// TestBuildURIWithBothParams tests URI building with both replica set and auth source
func TestBuildURIWithBothParams(t *testing.T) {
	cfg := &Config{
		Address:    "localhost",
		Port:       27017,
		ReplicaSet: "rs0",
		AuthSource: "admin",
	}
	uri := cfg.buildURI()
	
	if uri != "mongodb://localhost:27017/?replicaSet=rs0&authSource=admin" {
		t.Errorf("expected 'mongodb://localhost:27017/?replicaSet=rs0&authSource=admin', got '%s'", uri)
	}
}

// TestBuildURIWithCredentialsAndParams tests URI building with credentials and parameters
func TestBuildURIWithCredentialsAndParams(t *testing.T) {
	cfg := &Config{
		Address:    "localhost",
		Port:       27017,
		Username:   "admin",
		Password:   "password",
		ReplicaSet: "rs0",
		AuthSource: "admin",
	}
	uri := cfg.buildURI()
	
	if uri != "mongodb://admin:password@localhost:27017/?replicaSet=rs0&authSource=admin" {
		t.Errorf("expected 'mongodb://admin:password@localhost:27017/?replicaSet=rs0&authSource=admin', got '%s'", uri)
	}
}

// TestBuildURIWithCustomAddress tests URI building with custom address
func TestBuildURIWithCustomAddress(t *testing.T) {
	cfg := &Config{
		Address: "mongo.example.com",
		Port:    27017,
	}
	uri := cfg.buildURI()
	
	if uri != "mongodb://mongo.example.com:27017/" {
		t.Errorf("expected 'mongodb://mongo.example.com:27017/', got '%s'", uri)
	}
}

// TestBuildURIWithCustomPort tests URI building with custom port
func TestBuildURIWithCustomPort(t *testing.T) {
	cfg := &Config{
		Address: "localhost",
		Port:    27018,
	}
	uri := cfg.buildURI()
	
	if uri != "mongodb://localhost:27018/" {
		t.Errorf("expected 'mongodb://localhost:27018/', got '%s'", uri)
	}
}

// TestApplyDefaults tests that apply() sets default values
func TestApplyDefaults(t *testing.T) {
	cfg := &Config{}
	cfg.apply()
	
	if cfg.Address == "" {
		t.Error("expected Address to be set, got empty string")
	}
	if cfg.Port == 0 {
		t.Error("expected Port to be set, got 0")
	}
}

// TestApplyWithExistingValues tests that apply() doesn't override existing values
func TestApplyWithExistingValues(t *testing.T) {
	cfg := &Config{
		Address: "custom.com",
		Port:    27018,
	}
	originalAddress := cfg.Address
	originalPort := cfg.Port
	
	cfg.apply()
	
	if cfg.Address != originalAddress {
		t.Errorf("expected Address to remain '%s', got '%s'", originalAddress, cfg.Address)
	}
	if cfg.Port != originalPort {
		t.Errorf("expected Port to remain %d, got %d", originalPort, cfg.Port)
	}
}

// TestGetDatabaseWithDefaultValue tests GetDatabase returns "test" when not set
func TestGetDatabaseWithDefaultValue(t *testing.T) {
	client := &Client{
		cfg: &Config{
			Database: "", // Empty database
		},
	}
	
	dbName := client.GetDatabase()
	if dbName != "test" {
		t.Errorf("expected 'test', got '%s'", dbName)
	}
}

// TestGetDatabaseWithCustomValue tests GetDatabase returns custom database
func TestGetDatabaseWithCustomValue(t *testing.T) {
	client := &Client{
		cfg: &Config{
			Database: "mydb",
		},
	}
	
	dbName := client.GetDatabase()
	if dbName != "mydb" {
		t.Errorf("expected 'mydb', got '%s'", dbName)
	}
}
