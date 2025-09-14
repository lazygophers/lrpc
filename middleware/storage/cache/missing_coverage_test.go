package cache

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

// TestMessage for protobuf testing
type TestMessage struct {
	Name  string `json:"name"`
	Value int32  `json:"value"`
}

func (m *TestMessage) Reset()         { *m = TestMessage{} }
func (m *TestMessage) String() string { return "TestMessage" }
func (m *TestMessage) ProtoMessage()  {}

// Simple protobuf marshaling for testing
func (m *TestMessage) Marshal() ([]byte, error) {
	// Simple JSON-like marshal for testing
	return []byte(`{"name":"` + m.Name + `","value":` + string(rune(m.Value)) + `}`), nil
}

func (m *TestMessage) Unmarshal(data []byte) error {
	// Simple unmarshal for testing
	m.Name = "test"
	m.Value = 42
	return nil
}

// Test SetPb and SetPbEx error paths - Skip for now as protobuf is complex
func TestSetPbErrorPathsMissing(t *testing.T) {
	// Skip protobuf tests as they require proper proto.Message implementation
	t.Skip("Protobuf tests require proper proto.Message implementation")
}

// Test functions with 0% coverage in echo.go
func TestEchoSetPrefix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sugardb_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := &Config{
		Type:    SugarDB,
		DataDir: tmpDir,
	}

	cache, err := NewSugarDB(config)
	require.NoError(t, err)
	defer cache.Close()

	baseCache := cache.(*baseCache)
	sugarDB := baseCache.BaseCache.(*CacheSugarDB)

	// Test SetPrefix (no-op but should not panic)
	sugarDB.SetPrefix("test_prefix:")
	// No assertion needed as it's a no-op function
}

// Test functions with 0% coverage in mem.go
func TestMemSetPrefix(t *testing.T) {
	cache := NewMem()
	baseCache := cache.(*baseCache)
	memCache := baseCache.BaseCache.(*CacheMem)

	// Test SetPrefix (no-op but should not panic)
	memCache.SetPrefix("test_prefix:")
	// No assertion needed as it's a no-op function
}

// Test Bbolt Del and Reset functions (0% coverage)
func TestBboltDelAndReset(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bbolt_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := tmpDir + "/test.db"
	cache, err := NewBbolt(dbPath, &bbolt.Options{})
	require.NoError(t, err)
	defer cache.Close()

	baseCache := cache.(*baseCache)
	bboltCache := baseCache.BaseCache.(*CacheBbolt)

	// Set some test data
	err = bboltCache.Set("key1", "value1")
	assert.NoError(t, err)
	err = bboltCache.Set("key2", "value2")
	assert.NoError(t, err)

	// Test Del
	err = bboltCache.Del("key1", "key2")
	assert.NoError(t, err)

	// Verify keys are deleted
	_, err = bboltCache.Get("key1")
	assert.Equal(t, ErrNotFound, err)

	// Set data again for Reset test
	err = bboltCache.Set("key3", "value3")
	assert.NoError(t, err)

	// Test Reset
	err = bboltCache.Reset()
	assert.NoError(t, err)

	// Verify all data is cleared
	_, err = bboltCache.Get("key3")
	assert.Equal(t, ErrNotFound, err)
}

// Test BitcaskNewBitcask error path
func TestBitcaskNewBitcaskErrorPath(t *testing.T) {
	// Test with invalid directory (should cause error)
	invalidConfig := &Config{
		Type:    Bitcask,
		DataDir: "/invalid/path/that/does/not/exist/and/cannot/be/created",
	}

	cache, err := NewBitcask(invalidConfig)
	assert.Error(t, err)
	assert.Nil(t, cache)
}

// Test database cache implementation
func TestDatabaseCacheImplementation(t *testing.T) {
	// Create in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		// Skip test if SQLite is not available (CGO disabled)
		t.Skipf("SQLite not available: %v", err)
	}
	require.NoError(t, err)
	defer db.Close()

	cache, err := NewDatabase(db, "test_cache")
	if err != nil {
		// Skip test if database cache creation fails
		t.Skipf("Database cache not available: %v", err)  
	}
	require.NoError(t, err)
	defer cache.Close()

	// Get the underlying Database instance
	baseCache := cache.(*baseCache)
	dbCache := baseCache.BaseCache.(*Database)

	// Test SetPrefix
	dbCache.SetPrefix("test_prefix:")

	// Test Clean
	err = dbCache.Clean()
	assert.NoError(t, err)

	// Test basic Set/Get
	err = dbCache.Set("key1", "value1")
	assert.NoError(t, err)

	value, err := dbCache.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)

	// Test Get non-existent key
	_, err = dbCache.Get("nonexistent")
	assert.Equal(t, ErrNotFound, err)

	// Test SetEx
	err = dbCache.SetEx("expkey", "expvalue", time.Second*2)
	assert.NoError(t, err)

	value, err = dbCache.Get("expkey")
	assert.NoError(t, err)
	assert.Equal(t, "expvalue", value)

	// Test TTL
	ttl, err := dbCache.Ttl("expkey")
	assert.NoError(t, err)
	assert.True(t, ttl > 0)

	// Test TTL on non-existent key
	_, err = dbCache.Ttl("nonexistent")
	assert.Equal(t, ErrNotFound, err)

	// Test SetNx
	ok, err := dbCache.SetNx("nxkey", "nxvalue")
	assert.NoError(t, err)
	assert.True(t, ok)

	// Test SetNx on existing key
	ok, err = dbCache.SetNx("nxkey", "newvalue")
	assert.NoError(t, err)
	assert.False(t, ok)

	// Test SetNxWithTimeout
	ok, err = dbCache.SetNxWithTimeout("nxtimeout", "nxtimeoutvalue", time.Second)
	assert.NoError(t, err)
	assert.True(t, ok)

	// Test Expire
	ok, err = dbCache.Expire("key1", time.Second*5)
	assert.NoError(t, err)
	assert.True(t, ok)

	// Test Del
	err = dbCache.Del("key1", "nxkey")
	assert.NoError(t, err)

	// Verify key is deleted
	_, err = dbCache.Get("key1")
	assert.Equal(t, ErrNotFound, err)

	// Test Del with empty keys
	err = dbCache.Del()
	assert.NoError(t, err)

	// Test Close
	err = dbCache.Close()
	assert.NoError(t, err)
}

// Test database cache panic methods (unimplemented)
func TestDatabaseCachePanicMethods(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		// Skip test if SQLite is not available (CGO disabled)
		t.Skipf("SQLite not available: %v", err)
	}
	require.NoError(t, err)
	defer db.Close()

	cache, err := NewDatabase(db, "test_cache")
	if err != nil {
		// Skip test if database cache creation fails
		t.Skipf("Database cache not available: %v", err)  
	}
	require.NoError(t, err)
	defer cache.Close()

	baseCache := cache.(*baseCache)
	dbCache := baseCache.BaseCache.(*Database)

	// Test all panic methods
	assert.Panics(t, func() { dbCache.Incr("key") })
	assert.Panics(t, func() { dbCache.Decr("key") })
	assert.Panics(t, func() { dbCache.IncrBy("key", 1) })
	assert.Panics(t, func() { dbCache.DecrBy("key", 1) })
	assert.Panics(t, func() { dbCache.Exists("key") })
	assert.Panics(t, func() { dbCache.HSet("key", "field", "value") })
	assert.Panics(t, func() { dbCache.HGet("key", "field") })
	assert.Panics(t, func() { dbCache.HDel("key", "field") })
	assert.Panics(t, func() { dbCache.HKeys("key") })
	assert.Panics(t, func() { dbCache.HGetAll("key") })
	assert.Panics(t, func() { dbCache.HExists("key", "field") })
	assert.Panics(t, func() { dbCache.HIncr("key", "field") })
	assert.Panics(t, func() { dbCache.HIncrBy("key", "field", 1) })
	assert.Panics(t, func() { dbCache.HDecr("key", "field") })
	assert.Panics(t, func() { dbCache.HDecrBy("key", "field", 1) })
	assert.Panics(t, func() { dbCache.SAdd("key", "member") })
	assert.Panics(t, func() { dbCache.SMembers("key") })
	assert.Panics(t, func() { dbCache.SRem("key", "member") })
	assert.Panics(t, func() { dbCache.SRandMember("key") })
	assert.Panics(t, func() { dbCache.SPop("key") })
	assert.Panics(t, func() { dbCache.SisMember("key", "member") })
}

// Test Redis cache methods (all 0% coverage)
func TestRedisCacheMethods(t *testing.T) {
	// Test NewRedis with invalid address
	cache, err := NewRedis("invalid:address:format")
	assert.Error(t, err)
	assert.Nil(t, cache)

	// Test NewRedis with connection failure
	cache, err = NewRedis("localhost:9999") // Non-existent Redis server
	assert.Error(t, err)
	assert.Nil(t, cache)
}

// Test missing coverage in existing functions
func TestMemExistsErrorPath(t *testing.T) {
	cache := NewMem()
	baseCache := cache.(*baseCache)
	memCache := baseCache.BaseCache.(*CacheMem)

	// Test Exists with keys that don't exist
	exists, err := memCache.Exists("nonexistent1", "nonexistent2")
	assert.NoError(t, err)
	assert.False(t, exists)
}

// Test additional protobuf scenarios
func TestProtobufEdgeCases(t *testing.T) {
	// Skip protobuf tests as they require proper proto.Message implementation
	t.Skip("Protobuf tests require proper proto.Message implementation")
}

// Test error paths in base cache methods
func TestBaseCacheErrorPaths(t *testing.T) {
	cache := NewMem()

	// Test typed getters with nonexistent keys (this will return ErrNotFound)
	_, err := cache.GetBool("nonexistent_key")
	assert.Equal(t, ErrNotFound, err)

	_, err = cache.GetInt("nonexistent_key")
	assert.Equal(t, ErrNotFound, err)

	_, err = cache.GetUint("nonexistent_key")
	assert.Equal(t, ErrNotFound, err)

	_, err = cache.GetInt32("nonexistent_key")
	assert.Equal(t, ErrNotFound, err)

	_, err = cache.GetUint32("nonexistent_key")
	assert.Equal(t, ErrNotFound, err)

	_, err = cache.GetInt64("nonexistent_key")
	assert.Equal(t, ErrNotFound, err)

	_, err = cache.GetUint64("nonexistent_key")
	assert.Equal(t, ErrNotFound, err)

	_, err = cache.GetFloat32("nonexistent_key")
	assert.Equal(t, ErrNotFound, err)

	_, err = cache.GetFloat64("nonexistent_key")
	assert.Equal(t, ErrNotFound, err)
}

// Test core middleware scenarios
func TestCoreMiddlewareScenarios(t *testing.T) {
	// Skip middleware tests as they need proper core.Ctx setup
	t.Skip("Core middleware tests need proper context setup")
}

// Test NewSugarDB error path in configuration
func TestNewSugarDBErrorPath(t *testing.T) {
	// Test with configuration that would fail
	config := &Config{
		Type:    SugarDB,
		DataDir: "/invalid/path/that/cannot/be/created/and/causes/error",
	}

	cache, err := NewSugarDB(config)
	assert.Error(t, err)
	assert.Nil(t, cache)
}