package mongo

import (
	"context"
	"sync"
	"time"
)

// CallRecord records a method call on MockClient for assertion purposes
type CallRecord struct {
	Method string
	Args   []interface{}
	Time   time.Time
}

// MockClient is a mock implementation of Client interface for testing
type MockClient struct {
	mu sync.RWMutex

	// Return values
	PingErr     error
	CloseErr    error
	ConfigRet   *Config
	ContextRet  context.Context
	DatabaseRet string
	HealthErr   error
	ScoopRet    *Scoop
	WatchErr    error

	// Call tracking
	calls      []CallRecord
	callCounts map[string]int
}

// NewMockClient creates a new MockClient instance
func NewMockClient() *MockClient {
	return &MockClient{
		calls:      make([]CallRecord, 0),
		callCounts: make(map[string]int),
		ContextRet: context.Background(),
	}
}

// Ping simulates Client.Ping
func (m *MockClient) Ping() error {
	m.recordCall("Ping")
	return m.PingErr
}

// Close simulates Client.Close
func (m *MockClient) Close() error {
	m.recordCall("Close")
	return m.CloseErr
}

// GetConfig simulates Client.GetConfig
func (m *MockClient) GetConfig() *Config {
	m.recordCall("GetConfig")
	return m.ConfigRet
}

// Context simulates Client.Context
func (m *MockClient) Context() context.Context {
	m.recordCall("Context")
	if m.ContextRet == nil {
		return context.Background()
	}
	return m.ContextRet
}

// GetDatabase simulates Client.GetDatabase
func (m *MockClient) GetDatabase() string {
	m.recordCall("GetDatabase")
	return m.DatabaseRet
}

// Health simulates Client.Health
func (m *MockClient) Health() error {
	m.recordCall("Health")
	return m.HealthErr
}

// NewScoop simulates Client.NewScoop
func (m *MockClient) NewScoop(tx ...*Scoop) *Scoop {
	m.recordCall("NewScoop", tx)
	return m.ScoopRet
}

// WatchAllCollections simulates Client.WatchAllCollections
func (m *MockClient) WatchAllCollections() (*DatabaseChangeStream, error) {
	m.recordCall("WatchAllCollections")
	return nil, m.WatchErr
}

// Call recording and querying

// recordCall records a method call for testing assertions
func (m *MockClient) recordCall(method string, args ...interface{}) {
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
func (m *MockClient) GetCalls() []CallRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	calls := make([]CallRecord, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// GetCallCount returns the number of times a method was called
func (m *MockClient) GetCallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.callCounts[method]
}

// AssertCalled asserts that a method was called
func (m *MockClient) AssertCalled(method string) bool {
	return m.GetCallCount(method) > 0
}

// AssertNotCalled asserts that a method was not called
func (m *MockClient) AssertNotCalled(method string) bool {
	return m.GetCallCount(method) == 0
}

// ResetCalls clears all recorded calls
func (m *MockClient) ResetCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = make([]CallRecord, 0)
	m.callCounts = make(map[string]int)
}

// SetupPingSuccess configures Ping to succeed
func (m *MockClient) SetupPingSuccess() *MockClient {
	m.PingErr = nil
	return m
}

// SetupPingError configures Ping to return an error
func (m *MockClient) SetupPingError(err error) *MockClient {
	m.PingErr = err
	return m
}

// SetupCloseSuccess configures Close to succeed
func (m *MockClient) SetupCloseSuccess() *MockClient {
	m.CloseErr = nil
	return m
}

// SetupCloseError configures Close to return an error
func (m *MockClient) SetupCloseError(err error) *MockClient {
	m.CloseErr = err
	return m
}

// SetupConfig sets the configuration to return
func (m *MockClient) SetupConfig(cfg *Config) *MockClient {
	m.ConfigRet = cfg
	return m
}

// SetupContext sets the context to return
func (m *MockClient) SetupContext(ctx context.Context) *MockClient {
	m.ContextRet = ctx
	return m
}

// SetupDatabase sets the database name to return
func (m *MockClient) SetupDatabase(name string) *MockClient {
	m.DatabaseRet = name
	return m
}

// SetupHealthSuccess configures Health to succeed
func (m *MockClient) SetupHealthSuccess() *MockClient {
	m.HealthErr = nil
	return m
}

// SetupHealthError configures Health to return an error
func (m *MockClient) SetupHealthError(err error) *MockClient {
	m.HealthErr = err
	return m
}

// SetupScoop sets the scoop to return
func (m *MockClient) SetupScoop(scoop *Scoop) *MockClient {
	m.ScoopRet = scoop
	return m
}

// SetupWatchError configures WatchAllCollections to return an error
func (m *MockClient) SetupWatchError(err error) *MockClient {
	m.WatchErr = err
	return m
}

// MockModel is a mock implementation of Model interface for testing
type MockModel[M any] struct {
	mu sync.RWMutex

	// Return values
	CollectionNameRet string
	ModelRet          M
	NotFoundErrorRet  error
	ScoopRet          *ModelScoop[M]
	IsNotFoundRet     bool

	// Call tracking
	calls      []CallRecord
	callCounts map[string]int
}

// NewMockModel creates a new MockModel instance
func NewMockModel[M any]() *MockModel[M] {
	return &MockModel[M]{
		calls:      make([]CallRecord, 0),
		callCounts: make(map[string]int),
	}
}

// CollectionName simulates Model.CollectionName
func (m *MockModel[M]) CollectionName() string {
	m.recordCall("CollectionName")
	return m.CollectionNameRet
}

// IsNotFound simulates Model.IsNotFound
func (m *MockModel[M]) IsNotFound(err error) bool {
	m.recordCall("IsNotFound", err)
	return m.IsNotFoundRet
}

// NewScoop simulates Model.NewScoop
func (m *MockModel[M]) NewScoop(tx ...*Scoop) *ModelScoop[M] {
	m.recordCall("NewScoop", tx)
	return m.ScoopRet
}

// SetNotFound simulates Model.SetNotFound
func (m *MockModel[M]) SetNotFound(err error) *MockModel[M] {
	m.recordCall("SetNotFound", err)
	m.NotFoundErrorRet = err
	return m
}

// recordCall records a method call
func (m *MockModel[M]) recordCall(method string, args ...interface{}) {
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
func (m *MockModel[M]) GetCalls() []CallRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	calls := make([]CallRecord, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// GetCallCount returns the number of times a method was called
func (m *MockModel[M]) GetCallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.callCounts[method]
}

// AssertCalled asserts that a method was called
func (m *MockModel[M]) AssertCalled(method string) bool {
	return m.GetCallCount(method) > 0
}

// AssertNotCalled asserts that a method was not called
func (m *MockModel[M]) AssertNotCalled(method string) bool {
	return m.GetCallCount(method) == 0
}

// ResetCalls clears all recorded calls
func (m *MockModel[M]) ResetCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = make([]CallRecord, 0)
	m.callCounts = make(map[string]int)
}

// Setup methods for ModelMock

// SetupCollectionName sets the collection name to return
func (m *MockModel[M]) SetupCollectionName(name string) *MockModel[M] {
	m.CollectionNameRet = name
	return m
}

// SetupModel sets the model to return
func (m *MockModel[M]) SetupModel(model M) *MockModel[M] {
	m.ModelRet = model
	return m
}

// SetupScoop sets the scoop to return
func (m *MockModel[M]) SetupScoop(scoop *ModelScoop[M]) *MockModel[M] {
	m.ScoopRet = scoop
	return m
}

// SetupIsNotFound sets the result of IsNotFound check
func (m *MockModel[M]) SetupIsNotFound(result bool) *MockModel[M] {
	m.IsNotFoundRet = result
	return m
}

// SetupNotFoundError sets the not found error
func (m *MockModel[M]) SetupNotFoundError(err error) *MockModel[M] {
	m.NotFoundErrorRet = err
	return m
}
