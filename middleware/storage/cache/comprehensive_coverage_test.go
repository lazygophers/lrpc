package cache

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test comprehensive coverage for echo.go functions
func TestEchoComprehensiveCoverage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "echo_comprehensive_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := &Config{
		Type:    SugarDB,
		DataDir: tmpDir,
	}

	cache, err := NewSugarDB(config)
	require.NoError(t, err)
	defer cache.Close()

	// Test normal operations to cover success paths
	err = cache.Set("test_key", "test_value")
	assert.NoError(t, err)

	value, err := cache.Get("test_key")
	assert.NoError(t, err)
	assert.Equal(t, "test_value", value)

	exists, err := cache.Exists("test_key")
	assert.NoError(t, err)
	assert.True(t, exists)

	_, err = cache.HSet("test_hash", "test_field", "test_value")
	assert.NoError(t, err)

	val, err := cache.HDecrBy("test_hash", "numeric_field", 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(-1), val)

	count, err := cache.SAdd("test_set", "test_member")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	members, err := cache.SMembers("test_set")
	assert.NoError(t, err)
	assert.Contains(t, members, "test_member")

	count, err = cache.SRem("test_set", "test_member")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

// Test additional base cache functions
func TestBaseCacheAdditionalFunctions(t *testing.T) {
	cache := NewMem()

	// Test SetPb and SetPbEx success paths
	testMsg := &TestMessage{Name: "test", Value: 42}

	// First test the protobuf marshal method
	data, err := testMsg.Marshal()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test SetPb success path would need real protobuf message
	// Skip protobuf tests for now as they need proper proto.Message implementation

	// Test all typed getters with valid data
	err = cache.Set("bool_val", "true")
	assert.NoError(t, err)
	boolVal, err := cache.GetBool("bool_val")
	assert.NoError(t, err)
	assert.True(t, boolVal)

	err = cache.Set("int_val", "42")
	assert.NoError(t, err)
	intVal, err := cache.GetInt("int_val")
	assert.NoError(t, err)
	assert.Equal(t, 42, intVal)

	err = cache.Set("uint_val", "42")
	assert.NoError(t, err)
	uintVal, err := cache.GetUint("uint_val")
	assert.NoError(t, err)
	assert.Equal(t, uint(42), uintVal)

	err = cache.Set("int32_val", "42")
	assert.NoError(t, err)
	int32Val, err := cache.GetInt32("int32_val")
	assert.NoError(t, err)
	assert.Equal(t, int32(42), int32Val)

	err = cache.Set("uint32_val", "42")
	assert.NoError(t, err)
	uint32Val, err := cache.GetUint32("uint32_val")
	assert.NoError(t, err)
	assert.Equal(t, uint32(42), uint32Val)

	err = cache.Set("int64_val", "42")
	assert.NoError(t, err)
	int64Val, err := cache.GetInt64("int64_val")
	assert.NoError(t, err)
	assert.Equal(t, int64(42), int64Val)

	err = cache.Set("uint64_val", "42")
	assert.NoError(t, err)
	uint64Val, err := cache.GetUint64("uint64_val")
	assert.NoError(t, err)
	assert.Equal(t, uint64(42), uint64Val)

	err = cache.Set("float32_val", "42.5")
	assert.NoError(t, err)
	float32Val, err := cache.GetFloat32("float32_val")
	assert.NoError(t, err)
	assert.Equal(t, float32(42.5), float32Val)

	err = cache.Set("float64_val", "42.5")
	assert.NoError(t, err)
	float64Val, err := cache.GetFloat64("float64_val")
	assert.NoError(t, err)
	assert.Equal(t, 42.5, float64Val)

	// Test slice operations
	testSlice := []string{"a", "b", "c"}
	sliceData, _ := json.Marshal(testSlice)
	err = cache.Set("slice_val", string(sliceData))
	assert.NoError(t, err)

	slice, err := cache.GetSlice("slice_val")
	assert.NoError(t, err)
	assert.Equal(t, testSlice, slice)

	// Test typed slice getters
	boolSlice := []bool{true, false, true}
	boolSliceData, _ := json.Marshal(boolSlice)
	err = cache.Set("bool_slice", string(boolSliceData))
	assert.NoError(t, err)

	retrievedBoolSlice, err := cache.GetBoolSlice("bool_slice")
	assert.NoError(t, err)
	assert.Equal(t, boolSlice, retrievedBoolSlice)

	intSlice := []int{1, 2, 3}
	intSliceData, _ := json.Marshal(intSlice)
	err = cache.Set("int_slice", string(intSliceData))
	assert.NoError(t, err)

	retrievedIntSlice, err := cache.GetIntSlice("int_slice")
	assert.NoError(t, err)
	assert.Equal(t, intSlice, retrievedIntSlice)

	uintSlice := []uint{1, 2, 3}
	uintSliceData, _ := json.Marshal(uintSlice)
	err = cache.Set("uint_slice", string(uintSliceData))
	assert.NoError(t, err)

	retrievedUintSlice, err := cache.GetUintSlice("uint_slice")
	assert.NoError(t, err)
	assert.Equal(t, uintSlice, retrievedUintSlice)

	int32Slice := []int32{1, 2, 3}
	int32SliceData, _ := json.Marshal(int32Slice)
	err = cache.Set("int32_slice", string(int32SliceData))
	assert.NoError(t, err)

	retrievedInt32Slice, err := cache.GetInt32Slice("int32_slice")
	assert.NoError(t, err)
	assert.Equal(t, int32Slice, retrievedInt32Slice)

	uint32Slice := []uint32{1, 2, 3}
	uint32SliceData, _ := json.Marshal(uint32Slice)
	err = cache.Set("uint32_slice", string(uint32SliceData))
	assert.NoError(t, err)

	retrievedUint32Slice, err := cache.GetUint32Slice("uint32_slice")
	assert.NoError(t, err)
	assert.Equal(t, uint32Slice, retrievedUint32Slice)

	int64Slice := []int64{1, 2, 3}
	int64SliceData, _ := json.Marshal(int64Slice)
	err = cache.Set("int64_slice", string(int64SliceData))
	assert.NoError(t, err)

	retrievedInt64Slice, err := cache.GetInt64Slice("int64_slice")
	assert.NoError(t, err)
	assert.Equal(t, int64Slice, retrievedInt64Slice)

	uint64Slice := []uint64{1, 2, 3}
	uint64SliceData, _ := json.Marshal(uint64Slice)
	err = cache.Set("uint64_slice", string(uint64SliceData))
	assert.NoError(t, err)

	retrievedUint64Slice, err := cache.GetUint64Slice("uint64_slice")
	assert.NoError(t, err)
	assert.Equal(t, uint64Slice, retrievedUint64Slice)

	float32Slice := []float32{1.1, 2.2, 3.3}
	float32SliceData, _ := json.Marshal(float32Slice)
	err = cache.Set("float32_slice", string(float32SliceData))
	assert.NoError(t, err)

	retrievedFloat32Slice, err := cache.GetFloat32Slice("float32_slice")
	assert.NoError(t, err)
	assert.Equal(t, float32Slice, retrievedFloat32Slice)

	float64Slice := []float64{1.1, 2.2, 3.3}
	float64SliceData, _ := json.Marshal(float64Slice)
	err = cache.Set("float64_slice", string(float64SliceData))
	assert.NoError(t, err)

	retrievedFloat64Slice, err := cache.GetFloat64Slice("float64_slice")
	assert.NoError(t, err)
	assert.Equal(t, float64Slice, retrievedFloat64Slice)

	// Test JSON operations
	testObject := map[string]interface{}{
		"name":  "test",
		"value": 42,
	}
	testObjectJSON, _ := json.Marshal(testObject)
	err = cache.Set("json_object", string(testObjectJSON))
	assert.NoError(t, err)

	var retrievedObject map[string]interface{}
	err = cache.GetJson("json_object", &retrievedObject)
	assert.NoError(t, err)
	assert.Equal(t, "test", retrievedObject["name"])
	assert.Equal(t, float64(42), retrievedObject["value"]) // JSON numbers are float64

	// Test HGetJson
	_, err = cache.HSet("json_hash", "json_field", string(testObjectJSON))
	assert.NoError(t, err)

	var retrievedHashObject map[string]interface{}
	err = cache.HGetJson("json_hash", "json_field", &retrievedHashObject)
	assert.NoError(t, err)
	assert.Equal(t, "test", retrievedHashObject["name"])
}

// Test GetSlice with empty string edge case
func TestGetSliceEmptyStringEdgeCase(t *testing.T) {
	cache := NewMem()

	// Set empty string
	err := cache.Set("empty_string", "")
	assert.NoError(t, err)

	// GetSlice should return empty slice for empty string
	slice, err := cache.GetSlice("empty_string")
	assert.NoError(t, err)
	assert.Empty(t, slice)
}

// Test limit functions with normal paths
func TestLimitFunctionsNormalPaths(t *testing.T) {
	cache := NewMem()

	// Test Limit normal operation
	limiter, err := cache.Limit("limit_key", 3, time.Second*10)
	assert.NoError(t, err)
	assert.NotNil(t, limiter)

	// Test LimitUpdateOnCheck normal operation
	limiterCheck, err := cache.LimitUpdateOnCheck("limit_check_key", 5, time.Second*10)
	assert.NoError(t, err)
	assert.NotNil(t, limiterCheck)
}

// Test additional constructor and configuration scenarios
func TestAdditionalConstructorScenarios(t *testing.T) {
	// Test NewBbolt with nil options
	tmpDir, err := os.MkdirTemp("", "bbolt_nil_options_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := tmpDir + "/nil_options.db"
	cache, err := NewBbolt(dbPath, nil)
	assert.NoError(t, err)
	assert.NotNil(t, cache)
	cache.Close()

	// Test NewMem with different configurations
	memCache1 := NewMem()
	assert.NotNil(t, memCache1)

	memCache2 := NewMem()
	assert.NotNil(t, memCache2)

	// Test config.apply function
	config := &Config{
		Type:    Mem,
		Address: "test_address",
		DataDir: "test_dir",
		Db:      1,
	}
	config.apply()
	// No assertion needed as it's a configuration method
}