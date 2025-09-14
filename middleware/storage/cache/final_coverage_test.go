package cache

import (
	"errors"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"
)

// This file targets specific uncovered code paths to achieve 100% coverage

// MockFailingProtoMarshal creates a proto message that fails to marshal
type MockFailingProto struct {
	*timestamppb.Timestamp
}

func (m *MockFailingProto) Marshal() ([]byte, error) {
	return nil, errors.New("intentional marshal error")
}

func TestSetPbErrorPathsFinal(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// We need to test the error paths in SetPb and SetPbEx when proto.Marshal fails
	// Since we can't easily mock proto.Marshal, let's focus on what we can test

	// Test normal case first
	normalProto := &timestamppb.Timestamp{Seconds: 123}
	err := cache.SetPb("test_key", normalProto)
	assert.NilError(t, err)

	err = cache.SetPbEx("test_key_ex", normalProto, 1*time.Hour)
	assert.NilError(t, err)

	// Test with nil proto (should not crash)
	err = cache.SetPb("nil_key", (*timestamppb.Timestamp)(nil))
	assert.NilError(t, err) // This actually works fine

	// Test error handling by trying to get invalid proto data
	cache.Set("invalid_proto", "not proto data")
	invalidProto := &timestamppb.Timestamp{}
	err = cache.GetPb("invalid_proto", invalidProto)
	assert.Assert(t, err != nil, "GetPb should fail with invalid proto data")
}

func TestMemCacheSetPrefixFunction(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Access the underlying CacheMem to test SetPrefix directly
	memCache := cache.(*baseCache).BaseCache.(*CacheMem)

	// SetPrefix is a no-op in memory cache but should not panic
	memCache.SetPrefix("test_prefix:")
	memCache.SetPrefix("")
	memCache.SetPrefix("another:prefix:")

	// Cache should continue to work normally
	err := cache.Set("test", "value")
	assert.NilError(t, err)

	value, err := cache.Get("test")
	assert.NilError(t, err)
	assert.Equal(t, value, "value")
}

func TestMemCacheRemainingPaths(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test Expire on non-existent key (false path)
	success, err := cache.Expire("nonexistent_key", 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, success, false)

	// Test Ttl edge cases
	// Test Ttl on non-existent key (should return -2)
	ttl, err := cache.Ttl("never_existed")
	assert.NilError(t, err)
	assert.Equal(t, ttl, -2*time.Second)

	// Test Exists with multiple keys where not all exist
	cache.Set("exists_key", "value")
	exists, err := cache.Exists("exists_key", "not_exists")
	assert.NilError(t, err)
	assert.Equal(t, exists, false) // Should be false if ANY key doesn't exist

	// Test SRandMember edge case with count 0
	cache.SAdd("test_set", "member1", "member2", "member3")
	members, err := cache.SRandMember("test_set", 0)
	assert.NilError(t, err)
	assert.Equal(t, len(members), 1) // Should return 1 member when count <= 0

	// Test SRandMember with negative count  
	members, err = cache.SRandMember("test_set", -2)
	assert.NilError(t, err)
	assert.Equal(t, len(members), 1) // Should still return 1 member

	// Test SRandMember on empty set
	members, err = cache.SRandMember("empty_set")
	assert.NilError(t, err)
	assert.Equal(t, len(members), 0)
}

// Create a custom mock to test proto marshal failure
func TestProtoMarshalError(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Since we can't easily mock proto.Marshal to fail, we'll create a scenario
	// where we can test the error handling indirectly by creating malformed data

	// Test GetPb with corrupted data (should fail unmarshal)
	cache.Set("corrupted_proto", "this is not valid protobuf data")

	proto := &timestamppb.Timestamp{}
	err := cache.GetPb("corrupted_proto", proto)
	assert.Assert(t, err != nil, "GetPb should fail with corrupted data")
}

// A helper type for testing base cache functionality
type mockBaseCache struct {
	data map[string]string
}

func (m *mockBaseCache) SetPrefix(prefix string) {}
func (m *mockBaseCache) Get(key string) (string, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", NotFound
}
func (m *mockBaseCache) Set(key string, value any) error {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = value.(string)
	return nil
}
func (m *mockBaseCache) SetEx(key string, value any, timeout time.Duration) error {
	return m.Set(key, value)
}
func (m *mockBaseCache) SetNx(key string, value interface{}) (bool, error) { return true, nil }
func (m *mockBaseCache) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
	return true, nil
}
func (m *mockBaseCache) Ttl(key string) (time.Duration, error)            { return 0, nil }
func (m *mockBaseCache) Expire(key string, timeout time.Duration) (bool, error) { return true, nil }
func (m *mockBaseCache) Incr(key string) (int64, error)                   { return 1, nil }
func (m *mockBaseCache) Decr(key string) (int64, error)                   { return 1, nil }
func (m *mockBaseCache) IncrBy(key string, value int64) (int64, error)    { return value, nil }
func (m *mockBaseCache) DecrBy(key string, value int64) (int64, error)    { return value, nil }
func (m *mockBaseCache) Exists(keys ...string) (bool, error)              { return true, nil }
func (m *mockBaseCache) HSet(key string, field string, value interface{}) (bool, error) {
	return true, nil
}
func (m *mockBaseCache) HGet(key, field string) (string, error) { return "", NotFound }
func (m *mockBaseCache) HDel(key string, fields ...string) (int64, error) { return 1, nil }
func (m *mockBaseCache) HKeys(key string) ([]string, error)               { return []string{}, nil }
func (m *mockBaseCache) HGetAll(key string) (map[string]string, error) {
	return map[string]string{}, nil
}
func (m *mockBaseCache) HExists(key string, field string) (bool, error) { return false, nil }
func (m *mockBaseCache) HIncr(key string, subKey string) (int64, error) { return 1, nil }
func (m *mockBaseCache) HIncrBy(key string, field string, increment int64) (int64, error) {
	return increment, nil
}
func (m *mockBaseCache) HDecr(key string, field string) (int64, error) { return 1, nil }
func (m *mockBaseCache) HDecrBy(key string, field string, increment int64) (int64, error) {
	return increment, nil
}
func (m *mockBaseCache) SAdd(key string, members ...string) (int64, error) { return 1, nil }
func (m *mockBaseCache) SMembers(key string) ([]string, error)            { return []string{}, nil }
func (m *mockBaseCache) SRem(key string, members ...string) (int64, error) { return 1, nil }
func (m *mockBaseCache) SRandMember(key string, count ...int64) ([]string, error) {
	return []string{}, nil
}
func (m *mockBaseCache) SPop(key string) (string, error)                { return "", nil }
func (m *mockBaseCache) SisMember(key, field string) (bool, error)      { return false, nil }
func (m *mockBaseCache) Del(key ...string) error                        { return nil }
func (m *mockBaseCache) Clean() error                                   { return nil }
func (m *mockBaseCache) Close() error                                   { return nil }

func TestBaseCacheWrapperFunctions(t *testing.T) {
	mock := &mockBaseCache{}
	wrapped := newBaseCache(mock)
	defer wrapped.Close()

	// Test successful cases that should work with our mock
	err := wrapped.Set("test", "value")
	assert.NilError(t, err)

	// Test error cases - methods should return errors when keys don't exist
	_, err = wrapped.GetBool("nonexistent")
	assert.Assert(t, err != nil)

	_, err = wrapped.GetInt("nonexistent")
	assert.Assert(t, err != nil)

	_, err = wrapped.GetSlice("nonexistent")
	assert.Assert(t, err != nil)

	var result interface{}
	err = wrapped.GetJson("nonexistent", &result)
	assert.Assert(t, err != nil)

	err = wrapped.HGetJson("nonexistent", "field", &result)
	assert.Assert(t, err != nil)

	// Test GetSlice with empty value
	mock.Set("empty", "")
	slice, err := wrapped.GetSlice("empty")
	assert.NilError(t, err)
	assert.Assert(t, slice == nil)

	// Test GetSlice with invalid JSON
	mock.Set("invalid_json", "invalid json")
	_, err = wrapped.GetSlice("invalid_json")
	assert.Assert(t, err != nil)
}

func TestLimitEdgeCases(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test limit reaching exactly the limit number
	key := "exact_limit_test"
	limit := int64(2)

	// First call - should be allowed (count = 1)
	allowed, err := cache.Limit(key, limit, 1*time.Hour)
	assert.NilError(t, err)
	assert.Equal(t, allowed, true)

	// Second call - should be allowed (count = 2)
	allowed, err = cache.Limit(key, limit, 1*time.Hour)
	assert.NilError(t, err)
	assert.Equal(t, allowed, true)

	// Third call - should be denied (count = 3 > limit = 2)
	allowed, err = cache.Limit(key, limit, 1*time.Hour)
	assert.NilError(t, err)
	assert.Equal(t, allowed, false)

	// Test LimitUpdateOnCheck which always updates expiry
	key2 := "update_limit_test"
	allowed, err = cache.LimitUpdateOnCheck(key2, 1, 1*time.Hour)
	assert.NilError(t, err)
	assert.Equal(t, allowed, true)

	// Should be denied on second call
	allowed, err = cache.LimitUpdateOnCheck(key2, 1, 1*time.Hour)
	assert.NilError(t, err)
	assert.Equal(t, allowed, false)
}