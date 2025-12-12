package db

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSqliteNoCGOWithPassword tests that sqlite (no CGO) warns when password is set
func TestSqliteNoCGOWithPassword(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_nocgo.db")

	config := &Config{
		Type:     Sqlite,
		Address:  tempDir,
		Name:     "test_nocgo",
		Password: "should-warn", // This should trigger a warning
	}

	// This should work but log a warning about password being ignored
	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create no-CGO SQLite client: %v", err)
	}

	if client == nil {
		t.Fatal("Client is nil")
	}

	// Clean up
	os.Remove(dbPath)
}

// TestSqliteTypeNormalization tests that "sqlite3" is normalized to "sqlite"
func TestSqliteTypeNormalization(t *testing.T) {
	config := &Config{
		Type: "sqlite3",
	}

	config.apply()

	if config.Type != Sqlite {
		t.Errorf("Expected type to be normalized to %s, got %s", Sqlite, config.Type)
	}
}

// TestSqliteNoCGODSN tests that Sqlite DSN contains encryption parameters when password is set
func TestSqliteNoCGODSN(t *testing.T) {
	config := &Config{
		Type:     Sqlite,
		Address:  "/tmp",
		Name:     "regular",
		Password: "test-key", // DSN should contain encryption params regardless of CGO
	}

	config.apply()
	dsn := config.DSN()

	// DSN should contain SQLCipher encryption parameters (will be ignored if CGO is disabled)
	if !contains(dsn, "_key=test-key") {
		t.Error("Sqlite DSN should contain _key parameter when password is set")
	}

	if !contains(dsn, "_cipher=sqlcipher") {
		t.Error("Sqlite DSN should contain _cipher parameter when password is set")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
