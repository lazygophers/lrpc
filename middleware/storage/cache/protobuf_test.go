package cache

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"
)

func TestBaseCacheProtobuf(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Create a test protobuf message
	testMsg := &timestamppb.Timestamp{
		Seconds: 1234567890,
		Nanos:   123456789,
	}

	// Test SetPb
	err := cache.SetPb("proto_key", testMsg)
	assert.NilError(t, err)

	// Test GetPb
	var retrieved timestamppb.Timestamp
	err = cache.GetPb("proto_key", &retrieved)
	assert.NilError(t, err)
	assert.Equal(t, retrieved.Seconds, testMsg.Seconds)
	assert.Equal(t, retrieved.Nanos, testMsg.Nanos)

	// Test SetPbEx
	testMsg2 := &timestamppb.Timestamp{
		Seconds: 9876543210,
		Nanos:   987654321,
	}
	err = cache.SetPbEx("proto_key_ex", testMsg2, 1*time.Second)
	assert.NilError(t, err)

	// Test GetPb with SetPbEx
	var retrieved2 timestamppb.Timestamp
	err = cache.GetPb("proto_key_ex", &retrieved2)
	assert.NilError(t, err)
	assert.Equal(t, retrieved2.Seconds, testMsg2.Seconds)
	assert.Equal(t, retrieved2.Nanos, testMsg2.Nanos)

	// Test GetPb with non-existent key
	var notFound timestamppb.Timestamp
	err = cache.GetPb("nonexistent_proto", &notFound)
	assert.Equal(t, err, NotFound)
}

func TestBaseCacheProtobufErrors(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test GetPb with invalid protobuf data
	cache.Set("invalid_proto", "not valid protobuf data")
	var invalid timestamppb.Timestamp
	err := cache.GetPb("invalid_proto", &invalid)
	assert.Assert(t, err != nil, "Should return error for invalid protobuf data")
}
