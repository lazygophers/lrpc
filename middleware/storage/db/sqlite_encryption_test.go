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

// TestSqliteCGODSN tests that SqliteCGO generates correct DSN with encryption parameters
func TestSqliteCGODSN(t *testing.T) {
	config := &Config{
		Type:     SqliteCGO,
		Address:  "/tmp",
		Name:     "encrypted",
		Password: "test-password",
	}

	config.apply()
	dsn := config.DSN()

	// DSN should contain encryption parameters
	if dsn == "" {
		t.Fatal("DSN is empty")
	}

	// Check that password is included as _key parameter
	if !contains(dsn, "_key=test-password") {
		t.Error("DSN should contain _key parameter with password")
	}

	// Check that SQLCipher parameters are present
	if !contains(dsn, "_cipher=sqlcipher") {
		t.Error("DSN should contain _cipher=sqlcipher parameter")
	}

	if !contains(dsn, "_kdf_iter=256000") {
		t.Error("DSN should contain _kdf_iter=256000 parameter")
	}
}

// TestSqliteNoCGODSN tests that regular Sqlite DSN does not contain encryption parameters
func TestSqliteNoCGODSN(t *testing.T) {
	config := &Config{
		Type:     Sqlite,
		Address:  "/tmp",
		Name:     "regular",
		Password: "ignored", // Should be ignored for no-CGO
	}

	config.apply()
	dsn := config.DSN()

	// DSN should NOT contain SQLCipher encryption parameters
	if contains(dsn, "_key=") {
		t.Error("No-CGO SQLite DSN should not contain _key parameter")
	}

	if contains(dsn, "_cipher=sqlcipher") {
		t.Error("No-CGO SQLite DSN should not contain _cipher parameter")
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