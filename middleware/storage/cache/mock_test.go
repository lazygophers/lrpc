package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMockCache_SetData demonstrates the new SetData functionality
// This allows testing cache hit scenarios (P0 feature)
func TestMockCache_SetData(t *testing.T) {
	// Create mock cache
	mockCache := NewMockCache()

	// Define test data structure
	type Session struct {
		Uid       uint64
		Type      int32
		RoleId    int64
		ExpiresAt int64
	}

	// Set test data
	testSession := &Session{
		Uid:       456,
		Type:      1,
		RoleId:    1,
		ExpiresAt: 1704067200,
	}

	mockCache.SetData("session:123", testSession)

	// Test GetJson returns the set data
	var retrievedSession Session
	err := mockCache.GetJson("session:123", &retrievedSession)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, testSession.Uid, retrievedSession.Uid)
	assert.Equal(t, testSession.Type, retrievedSession.Type)
	assert.Equal(t, testSession.RoleId, retrievedSession.RoleId)
}

// TestMockCache_SetHashData demonstrates hash JSON data functionality
func TestMockCache_SetHashData(t *testing.T) {
	mockCache := NewMockCache()

	// Set hash field data
	permissions := []string{
		"GET:/api/admin/list",
		"POST:/api/admin/add",
		"DELETE:/api/admin/remove",
	}

	mockCache.SetHashData("rbac:1", "permissions", permissions)

	// Test HGetJson returns the set data
	var retrievedPerms []string
	err := mockCache.HGetJson("rbac:1", "permissions", &retrievedPerms)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, len(permissions), len(retrievedPerms))
	assert.Equal(t, permissions[0], retrievedPerms[0])
}

// TestMockCache_SetJsonString demonstrates JSON string functionality
func TestMockCache_SetJsonString(t *testing.T) {
	mockCache := NewMockCache()

	// Set using JSON string
	jsonStr := `{"uid":789,"type":2,"role_id":5}`
	mockCache.SetJsonString("user:789", jsonStr)

	// Test GetJson returns the parsed data
	type User struct {
		Uid    int `json:"uid"`
		Type   int `json:"type"`
		RoleId int `json:"role_id"`
	}

	var user User
	err := mockCache.GetJson("user:789", &user)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, 789, user.Uid)
	assert.Equal(t, 2, user.Type)
	assert.Equal(t, 5, user.RoleId)
}

// TestMockCache_CacheMiss demonstrates that errors are returned when data not set
func TestMockCache_CacheMiss(t *testing.T) {
	mockCache := NewMockCache()

	// Configure error for cache miss
	mockCache.GetJsonErr = ErrNotFound

	// Test GetJson returns error when no data is set
	var result interface{}
	err := mockCache.GetJson("nonexistent", &result)

	// Verify error is returned
	assert.Equal(t, ErrNotFound, err)
}

// TestMockCache_ChainableAPI demonstrates method chaining
func TestMockCache_ChainableAPI(t *testing.T) {
	mockCache := NewMockCache()

	// Chain multiple setup calls
	mockCache.
		SetData("key1", map[string]string{"field": "value1"}).
		SetData("key2", map[string]int{"count": 42}).
		SetHashData("hash1", "field1", "value1").
		SetJsonString("key3", `{"name":"test"}`)

	// Verify all data is set correctly
	var data1 map[string]string
	err := mockCache.GetJson("key1", &data1)
	assert.NoError(t, err)
	assert.Equal(t, "value1", data1["field"])

	var data2 map[string]int
	err = mockCache.GetJson("key2", &data2)
	assert.NoError(t, err)
	assert.Equal(t, 42, data2["count"])
}

// TestMockCache_CallTracking demonstrates call tracking still works
func TestMockCache_CallTracking(t *testing.T) {
	mockCache := NewMockCache()

	mockCache.SetData("key", "value")

	var result string
	mockCache.GetJson("key", &result)

	// Verify calls are recorded
	assert.True(t, mockCache.AssertCalled("GetJson"))
	assert.Equal(t, 1, mockCache.GetCallCount("GetJson"))

	calls := mockCache.GetCalls()
	assert.True(t, len(calls) > 0)
	assert.Equal(t, "GetJson", calls[0].Method)
}
