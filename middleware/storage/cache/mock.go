package cache

import (
	"encoding/json"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"

	"github.com/lazygophers/log"
)

// CallRecord records a method call on Mock for assertion purposes
type CallRecord struct {
	Method string
	Args   []interface{}
	Time   time.Time
}

// MockCache is a mock implementation of Cache interface for testing
type MockCache struct {
	mu sync.RWMutex

	// Return values for Set operations
	SetErr           error
	SetExErr         error
	SetNxRet         bool
	SetNxErr         error
	SetNxWithTimeErr error

	// Return values for Get operations
	GetRet string
	GetErr error

	// Return values for KV operations
	TtlRet       time.Duration
	TtlErr       error
	ExpireRet    bool
	ExpireErr    error
	ExistsRet    bool
	ExistsErr    error
	DelErr       error

	// Return values for Incr/Decr
	IncrRet    int64
	IncrErr    error
	DecrRet    int64
	DecrErr    error
	IncrByRet  int64
	IncrByErr  error
	DecrByRet  int64
	DecrByErr  error

	// Return values for Hash operations
	HSetRet       bool
	HSetErr       error
	HGetRet       string
	HGetErr       error
	HDelRet       int64
	HDelErr       error
	HKeysRet      []string
	HKeysErr      error
	HGetAllRet    map[string]string
	HGetAllErr    error
	HExistsRet    bool
	HExistsErr    error
	HIncrRet      int64
	HIncrErr      error
	HIncrByRet    int64
	HIncrByErr    error
	HDecrRet      int64
	HDecrErr      error
	HDecrByRet    int64
	HDecrByErr    error
	HGetJsonErr   error

	// Return values for Set operations
	SAddRet         int64
	SAddErr         error
	SMembersRet     []string
	SMembersErr     error
	SRemRet         int64
	SRemErr         error
	SRandMemberRet  []string
	SRandMemberErr  error
	SPopRet         string
	SPopErr         error
	SisMemberRet    bool
	SisMemberErr    error

	// Return values for Pub/Sub
	PublishRet     int64
	PublishErr     error
	SubscribeErr   error

	// Return values for Stream operations
	XAddRet        string
	XAddErr        error
	XLenRet        int64
	XLenErr        error
	XRangeRet      []map[string]interface{}
	XRangeErr      error
	XRevRangeRet   []map[string]interface{}
	XRevRangeErr   error
	XDelRet        int64
	XDelErr        error
	XTrimRet       int64
	XTrimErr       error
	XGroupErr      error
	XReadGroupErr  error
	XAckRet        int64
	XAckErr        error
	XPendingRet    int64
	XPendingErr    error

	// Return values for other operations
	CleanErr  error
	CloseErr  error
	PingErr   error

	// Return values for type conversions
	GetBoolRet     bool
	GetBoolErr     error
	GetIntRet      int
	GetIntErr      error
	GetUintRet     uint
	GetUintErr     error
	GetInt32Ret    int32
	GetInt32Err    error
	GetUint32Ret   uint32
	GetUint32Err   error
	GetInt64Ret    int64
	GetInt64Err    error
	GetUint64Ret   uint64
	GetUint64Err   error
	GetFloat32Ret  float32
	GetFloat32Err  error
	GetFloat64Ret  float64
	GetFloat64Err  error

	// Return values for slice conversions
	GetSliceRet       []string
	GetSliceErr       error
	GetBoolSliceRet   []bool
	GetBoolSliceErr   error
	GetIntSliceRet    []int
	GetIntSliceErr    error
	GetUintSliceRet   []uint
	GetUintSliceErr   error
	GetInt32SliceRet  []int32
	GetInt32SliceErr  error
	GetUint32SliceRet []uint32
	GetUint32SliceErr error
	GetInt64SliceRet  []int64
	GetInt64SliceErr  error
	GetUint64SliceRet []uint64
	GetUint64SliceErr error
	GetFloat32SliceRet []float32
	GetFloat32SliceErr error
	GetFloat64SliceRet []float64
	GetFloat64SliceErr error

	// Return values for serialization
	GetJsonErr    error
	SetPbErr      error
	SetPbExErr    error
	GetPbErr      error

	// Return values for limiting
	LimitRet           bool
	LimitErr           error
	LimitUpdateErr     error

	// Call tracking
	calls      []CallRecord
	callCounts map[string]int

	// Storage for mock data
	jsonData map[string]interface{}  // For JSON/Proto serialization
	pbData   map[string][]byte        // For Protobuf raw bytes
	hJsonData map[string]map[string]interface{} // For hash JSON data
}

// NewMockCache creates a new MockCache instance
func NewMockCache() *MockCache {
	return &MockCache{
		calls:        make([]CallRecord, 0),
		callCounts:   make(map[string]int),
		jsonData:     make(map[string]interface{}),
		pbData:       make(map[string][]byte),
		hJsonData:    make(map[string]map[string]interface{}),
	}
}

// recordCall records a method call for testing assertions
func (m *MockCache) recordCall(method string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record := CallRecord{
		Method: method,
		Args:   args,
		Time:   time.Now(),
	}
	m.calls = append(m.calls, record)
	m.callCounts[method]++
}

// GetCalls returns all recorded calls
func (m *MockCache) GetCalls() []CallRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	calls := make([]CallRecord, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// GetCallCount returns the number of times a method was called
func (m *MockCache) GetCallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.callCounts[method]
}

// AssertCalled asserts that a method was called
func (m *MockCache) AssertCalled(method string) bool {
	return m.GetCallCount(method) > 0
}

// AssertNotCalled asserts that a method was not called
func (m *MockCache) AssertNotCalled(method string) bool {
	return m.GetCallCount(method) == 0
}

// ResetCalls clears all recorded calls
func (m *MockCache) ResetCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = make([]CallRecord, 0)
	m.callCounts = make(map[string]int)
}

// BaseCache interface implementation

// SetPrefix implements BaseCache.SetPrefix
func (m *MockCache) SetPrefix(prefix string) {
	m.recordCall("SetPrefix", prefix)
}

// Get implements BaseCache.Get
func (m *MockCache) Get(key string) (string, error) {
	m.recordCall("Get", key)
	return m.GetRet, m.GetErr
}

// Set implements BaseCache.Set
func (m *MockCache) Set(key string, value any) error {
	m.recordCall("Set", key, value)
	return m.SetErr
}

// SetEx implements BaseCache.SetEx
func (m *MockCache) SetEx(key string, value any, timeout time.Duration) error {
	m.recordCall("SetEx", key, value, timeout)
	return m.SetExErr
}

// SetNx implements BaseCache.SetNx
func (m *MockCache) SetNx(key string, value interface{}) (bool, error) {
	m.recordCall("SetNx", key, value)
	return m.SetNxRet, m.SetNxErr
}

// SetNxWithTimeout implements BaseCache.SetNxWithTimeout
func (m *MockCache) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
	m.recordCall("SetNxWithTimeout", key, value, timeout)
	return m.SetNxRet, m.SetNxWithTimeErr
}

// Ttl implements BaseCache.Ttl
func (m *MockCache) Ttl(key string) (time.Duration, error) {
	m.recordCall("Ttl", key)
	return m.TtlRet, m.TtlErr
}

// Expire implements BaseCache.Expire
func (m *MockCache) Expire(key string, timeout time.Duration) (bool, error) {
	m.recordCall("Expire", key, timeout)
	return m.ExpireRet, m.ExpireErr
}

// Incr implements BaseCache.Incr
func (m *MockCache) Incr(key string) (int64, error) {
	m.recordCall("Incr", key)
	return m.IncrRet, m.IncrErr
}

// Decr implements BaseCache.Decr
func (m *MockCache) Decr(key string) (int64, error) {
	m.recordCall("Decr", key)
	return m.DecrRet, m.DecrErr
}

// IncrBy implements BaseCache.IncrBy
func (m *MockCache) IncrBy(key string, value int64) (int64, error) {
	m.recordCall("IncrBy", key, value)
	return m.IncrByRet, m.IncrByErr
}

// DecrBy implements BaseCache.DecrBy
func (m *MockCache) DecrBy(key string, value int64) (int64, error) {
	m.recordCall("DecrBy", key, value)
	return m.DecrByRet, m.DecrByErr
}

// Exists implements BaseCache.Exists
func (m *MockCache) Exists(keys ...string) (bool, error) {
	m.recordCall("Exists", keys)
	return m.ExistsRet, m.ExistsErr
}

// HSet implements BaseCache.HSet
func (m *MockCache) HSet(key, field string, value interface{}) (bool, error) {
	m.recordCall("HSet", key, field, value)
	return m.HSetRet, m.HSetErr
}

// HGet implements BaseCache.HGet
func (m *MockCache) HGet(key, field string) (string, error) {
	m.recordCall("HGet", key, field)
	return m.HGetRet, m.HGetErr
}

// HDel implements BaseCache.HDel
func (m *MockCache) HDel(key string, fields ...string) (int64, error) {
	m.recordCall("HDel", key, fields)
	return m.HDelRet, m.HDelErr
}

// HKeys implements BaseCache.HKeys
func (m *MockCache) HKeys(key string) ([]string, error) {
	m.recordCall("HKeys", key)
	return m.HKeysRet, m.HKeysErr
}

// HGetAll implements BaseCache.HGetAll
func (m *MockCache) HGetAll(key string) (map[string]string, error) {
	m.recordCall("HGetAll", key)
	return m.HGetAllRet, m.HGetAllErr
}

// HExists implements BaseCache.HExists
func (m *MockCache) HExists(key, field string) (bool, error) {
	m.recordCall("HExists", key, field)
	return m.HExistsRet, m.HExistsErr
}

// HIncr implements BaseCache.HIncr
func (m *MockCache) HIncr(key, subKey string) (int64, error) {
	m.recordCall("HIncr", key, subKey)
	return m.HIncrRet, m.HIncrErr
}

// HIncrBy implements BaseCache.HIncrBy
func (m *MockCache) HIncrBy(key, field string, increment int64) (int64, error) {
	m.recordCall("HIncrBy", key, field, increment)
	return m.HIncrByRet, m.HIncrByErr
}

// HDecr implements BaseCache.HDecr
func (m *MockCache) HDecr(key, field string) (int64, error) {
	m.recordCall("HDecr", key, field)
	return m.HDecrRet, m.HDecrErr
}

// HDecrBy implements BaseCache.HDecrBy
func (m *MockCache) HDecrBy(key, field string, increment int64) (int64, error) {
	m.recordCall("HDecrBy", key, field, increment)
	return m.HDecrByRet, m.HDecrByErr
}

// SAdd implements BaseCache.SAdd
func (m *MockCache) SAdd(key string, members ...string) (int64, error) {
	m.recordCall("SAdd", key, members)
	return m.SAddRet, m.SAddErr
}

// SMembers implements BaseCache.SMembers
func (m *MockCache) SMembers(key string) ([]string, error) {
	m.recordCall("SMembers", key)
	return m.SMembersRet, m.SMembersErr
}

// SRem implements BaseCache.SRem
func (m *MockCache) SRem(key string, members ...string) (int64, error) {
	m.recordCall("SRem", key, members)
	return m.SRemRet, m.SRemErr
}

// SRandMember implements BaseCache.SRandMember
func (m *MockCache) SRandMember(key string, count ...int64) ([]string, error) {
	m.recordCall("SRandMember", key, count)
	return m.SRandMemberRet, m.SRandMemberErr
}

// SPop implements BaseCache.SPop
func (m *MockCache) SPop(key string) (string, error) {
	m.recordCall("SPop", key)
	return m.SPopRet, m.SPopErr
}

// SisMember implements BaseCache.SisMember
func (m *MockCache) SisMember(key, field string) (bool, error) {
	m.recordCall("SisMember", key, field)
	return m.SisMemberRet, m.SisMemberErr
}

// Publish implements BaseCache.Publish
func (m *MockCache) Publish(channel string, message interface{}) (int64, error) {
	m.recordCall("Publish", channel, message)
	return m.PublishRet, m.PublishErr
}

// Subscribe implements BaseCache.Subscribe
func (m *MockCache) Subscribe(handler func(channel string, message []byte) error, channels ...string) error {
	m.recordCall("Subscribe", channels)
	return m.SubscribeErr
}

// XAdd implements BaseCache.XAdd
func (m *MockCache) XAdd(stream string, values map[string]interface{}) (string, error) {
	m.recordCall("XAdd", stream, values)
	return m.XAddRet, m.XAddErr
}

// XLen implements BaseCache.XLen
func (m *MockCache) XLen(stream string) (int64, error) {
	m.recordCall("XLen", stream)
	return m.XLenRet, m.XLenErr
}

// XRange implements BaseCache.XRange
func (m *MockCache) XRange(stream, start, stop string, count ...int64) ([]map[string]interface{}, error) {
	m.recordCall("XRange", stream, start, stop, count)
	return m.XRangeRet, m.XRangeErr
}

// XRevRange implements BaseCache.XRevRange
func (m *MockCache) XRevRange(stream, start, stop string, count ...int64) ([]map[string]interface{}, error) {
	m.recordCall("XRevRange", stream, start, stop, count)
	return m.XRevRangeRet, m.XRevRangeErr
}

// XDel implements BaseCache.XDel
func (m *MockCache) XDel(stream string, ids ...string) (int64, error) {
	m.recordCall("XDel", stream, ids)
	return m.XDelRet, m.XDelErr
}

// XTrim implements BaseCache.XTrim
func (m *MockCache) XTrim(stream string, maxLen int64) (int64, error) {
	m.recordCall("XTrim", stream, maxLen)
	return m.XTrimRet, m.XTrimErr
}

// XGroupCreate implements BaseCache.XGroupCreate
func (m *MockCache) XGroupCreate(stream, group, start string) error {
	m.recordCall("XGroupCreate", stream, group, start)
	return m.XGroupErr
}

// XGroupDestroy implements BaseCache.XGroupDestroy
func (m *MockCache) XGroupDestroy(stream, group string) error {
	m.recordCall("XGroupDestroy", stream, group)
	return m.XGroupErr
}

// XGroupSetID implements BaseCache.XGroupSetID
func (m *MockCache) XGroupSetID(stream, group, id string) error {
	m.recordCall("XGroupSetID", stream, group, id)
	return m.XGroupErr
}

// XReadGroup implements BaseCache.XReadGroup
func (m *MockCache) XReadGroup(handler func(stream string, id string, body []byte) error, group, consumer, stream string) error {
	m.recordCall("XReadGroup", group, consumer, stream)
	return m.XReadGroupErr
}

// XAck implements BaseCache.XAck
func (m *MockCache) XAck(stream, group string, ids ...string) (int64, error) {
	m.recordCall("XAck", stream, group, ids)
	return m.XAckRet, m.XAckErr
}

// XPending implements BaseCache.XPending
func (m *MockCache) XPending(stream, group string) (int64, error) {
	m.recordCall("XPending", stream, group)
	return m.XPendingRet, m.XPendingErr
}

// Del implements BaseCache.Del
func (m *MockCache) Del(key ...string) error {
	m.recordCall("Del", key)
	return m.DelErr
}

// Clean implements BaseCache.Clean
func (m *MockCache) Clean() error {
	m.recordCall("Clean")
	return m.CleanErr
}

// Close implements BaseCache.Close
func (m *MockCache) Close() error {
	m.recordCall("Close")
	return m.CloseErr
}

// Ping implements BaseCache.Ping
func (m *MockCache) Ping() error {
	m.recordCall("Ping")
	return m.PingErr
}

// Cache interface additional methods

// Base implements Cache.Base
func (m *MockCache) Base() BaseCache {
	m.recordCall("Base")
	return m
}

// GetBool implements Cache.GetBool
func (m *MockCache) GetBool(key string) (bool, error) {
	m.recordCall("GetBool", key)
	return m.GetBoolRet, m.GetBoolErr
}

// GetInt implements Cache.GetInt
func (m *MockCache) GetInt(key string) (int, error) {
	m.recordCall("GetInt", key)
	return m.GetIntRet, m.GetIntErr
}

// GetUint implements Cache.GetUint
func (m *MockCache) GetUint(key string) (uint, error) {
	m.recordCall("GetUint", key)
	return m.GetUintRet, m.GetUintErr
}

// GetInt32 implements Cache.GetInt32
func (m *MockCache) GetInt32(key string) (int32, error) {
	m.recordCall("GetInt32", key)
	return m.GetInt32Ret, m.GetInt32Err
}

// GetUint32 implements Cache.GetUint32
func (m *MockCache) GetUint32(key string) (uint32, error) {
	m.recordCall("GetUint32", key)
	return m.GetUint32Ret, m.GetUint32Err
}

// GetInt64 implements Cache.GetInt64
func (m *MockCache) GetInt64(key string) (int64, error) {
	m.recordCall("GetInt64", key)
	return m.GetInt64Ret, m.GetInt64Err
}

// GetUint64 implements Cache.GetUint64
func (m *MockCache) GetUint64(key string) (uint64, error) {
	m.recordCall("GetUint64", key)
	return m.GetUint64Ret, m.GetUint64Err
}

// GetFloat32 implements Cache.GetFloat32
func (m *MockCache) GetFloat32(key string) (float32, error) {
	m.recordCall("GetFloat32", key)
	return m.GetFloat32Ret, m.GetFloat32Err
}

// GetFloat64 implements Cache.GetFloat64
func (m *MockCache) GetFloat64(key string) (float64, error) {
	m.recordCall("GetFloat64", key)
	return m.GetFloat64Ret, m.GetFloat64Err
}

// GetSlice implements Cache.GetSlice
func (m *MockCache) GetSlice(key string) ([]string, error) {
	m.recordCall("GetSlice", key)
	return m.GetSliceRet, m.GetSliceErr
}

// GetBoolSlice implements Cache.GetBoolSlice
func (m *MockCache) GetBoolSlice(key string) ([]bool, error) {
	m.recordCall("GetBoolSlice", key)
	return m.GetBoolSliceRet, m.GetBoolSliceErr
}

// GetIntSlice implements Cache.GetIntSlice
func (m *MockCache) GetIntSlice(key string) ([]int, error) {
	m.recordCall("GetIntSlice", key)
	return m.GetIntSliceRet, m.GetIntSliceErr
}

// GetUintSlice implements Cache.GetUintSlice
func (m *MockCache) GetUintSlice(key string) ([]uint, error) {
	m.recordCall("GetUintSlice", key)
	return m.GetUintSliceRet, m.GetUintSliceErr
}

// GetInt32Slice implements Cache.GetInt32Slice
func (m *MockCache) GetInt32Slice(key string) ([]int32, error) {
	m.recordCall("GetInt32Slice", key)
	return m.GetInt32SliceRet, m.GetInt32SliceErr
}

// GetUint32Slice implements Cache.GetUint32Slice
func (m *MockCache) GetUint32Slice(key string) ([]uint32, error) {
	m.recordCall("GetUint32Slice", key)
	return m.GetUint32SliceRet, m.GetUint32SliceErr
}

// GetInt64Slice implements Cache.GetInt64Slice
func (m *MockCache) GetInt64Slice(key string) ([]int64, error) {
	m.recordCall("GetInt64Slice", key)
	return m.GetInt64SliceRet, m.GetInt64SliceErr
}

// GetUint64Slice implements Cache.GetUint64Slice
func (m *MockCache) GetUint64Slice(key string) ([]uint64, error) {
	m.recordCall("GetUint64Slice", key)
	return m.GetUint64SliceRet, m.GetUint64SliceErr
}

// GetFloat32Slice implements Cache.GetFloat32Slice
func (m *MockCache) GetFloat32Slice(key string) ([]float32, error) {
	m.recordCall("GetFloat32Slice", key)
	return m.GetFloat32SliceRet, m.GetFloat32SliceErr
}

// GetFloat64Slice implements Cache.GetFloat64Slice
func (m *MockCache) GetFloat64Slice(key string) ([]float64, error) {
	m.recordCall("GetFloat64Slice", key)
	return m.GetFloat64SliceRet, m.GetFloat64SliceErr
}

// GetJson implements Cache.GetJson
// Now supports returning cached data set via SetData()
func (m *MockCache) GetJson(key string, j interface{}) error {
	m.recordCall("GetJson", key, j)

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if data was set via SetData()
	if val, ok := m.jsonData[key]; ok {
		// Serialize the stored value to JSON then unmarshal into j
		b, err := json.Marshal(val)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
		err = json.Unmarshal(b, j)
		if err != nil {
			log.Errorf("err:%v", err)
		}
		return err
	}

	// Otherwise return the configured error
	return m.GetJsonErr
}

// SetPb implements Cache.SetPb
// Now supports storing protobuf data for later retrieval
func (m *MockCache) SetPb(key string, j proto.Message) error {
	m.recordCall("SetPb", key, j)

	if m.SetPbErr != nil {
		return m.SetPbErr
	}

	// Store the protobuf data for potential retrieval via GetPb
	b, err := proto.Marshal(j)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pbData == nil {
		m.pbData = make(map[string][]byte)
	}
	m.pbData[key] = append([]byte{}, b...)
	return nil
}

// SetPbEx implements Cache.SetPbEx
// Now supports storing protobuf data with timeout for later retrieval
func (m *MockCache) SetPbEx(key string, j proto.Message, timeout time.Duration) error {
	m.recordCall("SetPbEx", key, j, timeout)

	if m.SetPbExErr != nil {
		return m.SetPbExErr
	}

	// Store the protobuf data (timeout handling is left to specific implementation)
	b, err := proto.Marshal(j)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pbData == nil {
		m.pbData = make(map[string][]byte)
	}
	m.pbData[key] = append([]byte{}, b...)
	return nil
}

// GetPb implements Cache.GetPb
// Now supports returning cached protobuf data set via SetPb/SetPbEx/SetPbData
func (m *MockCache) GetPb(key string, j proto.Message) error {
	m.recordCall("GetPb", key, j)

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if protobuf data was set
	if data, ok := m.pbData[key]; ok {
		err := proto.Unmarshal(data, j)
		if err != nil {
			log.Errorf("err:%v", err)
		}
		return err
	}

	// Otherwise return the configured error
	return m.GetPbErr
}

// HGetJson implements Cache.HGetJson
// Now supports returning cached hash data set via SetHashData()
func (m *MockCache) HGetJson(key, field string, j interface{}) error {
	m.recordCall("HGetJson", key, field, j)

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if hash data was set via SetHashData()
	if hashData, ok := m.hJsonData[key]; ok {
		if val, fieldOk := hashData[field]; fieldOk {
			// Serialize the stored value to JSON then unmarshal into j
			b, err := json.Marshal(val)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
			err = json.Unmarshal(b, j)
			if err != nil {
				log.Errorf("err:%v", err)
			}
			return err
		}
	}

	// Otherwise return the configured error
	return m.HGetJsonErr
}

// Limit implements Cache.Limit
func (m *MockCache) Limit(key string, limit int64, timeout time.Duration) (bool, error) {
	m.recordCall("Limit", key, limit, timeout)
	return m.LimitRet, m.LimitErr
}

// LimitUpdateOnCheck implements Cache.LimitUpdateOnCheck
func (m *MockCache) LimitUpdateOnCheck(key string, limit int64, timeout time.Duration) (bool, error) {
	m.recordCall("LimitUpdateOnCheck", key, limit, timeout)
	return m.LimitRet, m.LimitUpdateErr
}

// Setup helper methods

// SetupError sets all error returns to the same value
func (m *MockCache) SetupError(err error) *MockCache {
	m.SetErr = err
	m.SetExErr = err
	m.SetNxErr = err
	m.SetNxWithTimeErr = err
	m.GetErr = err
	m.TtlErr = err
	m.ExpireErr = err
	m.ExistsErr = err
	m.DelErr = err
	return m
}

// SetupSuccess sets all error returns to nil
func (m *MockCache) SetupSuccess() *MockCache {
	return m.SetupError(nil)
}

// Data Setup Methods for Serialization Support

// SetData sets a value that will be returned by GetJson/HGetJson (P0 feature)
// This allows testing cache hit scenarios with actual data
func (m *MockCache) SetData(key string, value interface{}) *MockCache {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.jsonData == nil {
		m.jsonData = make(map[string]interface{})
	}
	m.jsonData[key] = value
	return m
}

// SetHashData sets hash field data that will be returned by HGetJson
func (m *MockCache) SetHashData(key, field string, value interface{}) *MockCache {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.hJsonData == nil {
		m.hJsonData = make(map[string]map[string]interface{})
	}
	if m.hJsonData[key] == nil {
		m.hJsonData[key] = make(map[string]interface{})
	}
	m.hJsonData[key][field] = value
	return m
}

// SetJsonString sets raw JSON string that will be returned by GetJson
func (m *MockCache) SetJsonString(key, jsonStr string) *MockCache {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.jsonData == nil {
		m.jsonData = make(map[string]interface{})
	}

	// Parse JSON string to validate and store
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		log.Errorf("err:%v", err)
		return m
	}
	m.jsonData[key] = data
	return m
}

// SetPbData sets protobuf data that will be returned by GetPb
func (m *MockCache) SetPbData(key string, data []byte) *MockCache {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pbData == nil {
		m.pbData = make(map[string][]byte)
	}
	// Make a copy to avoid external modifications
	m.pbData[key] = append([]byte{}, data...)
	return m
}
