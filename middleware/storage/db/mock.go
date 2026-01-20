package db

import (
	"database/sql"
	"errors"
	"sync"
	"time"

	"gorm.io/gorm"
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
	AutoMigratesErr error
	AutoMigrateErr  error
	PingErr         error
	CloseErr        error
	DatabaseRet     *gorm.DB
	SqlDBRet        *sql.DB
	SqlDBErr        error
	DriverTypeRet   string
	ScoopRet        *Scoop

	// Call tracking
	calls      []CallRecord
	callCounts map[string]int
}

// NewMockClient creates a new MockClient instance
func NewMockClient() *MockClient {
	return &MockClient{
		calls:      make([]CallRecord, 0),
		callCounts: make(map[string]int),
	}
}

// AutoMigrates simulates Client.AutoMigrates
func (m *MockClient) AutoMigrates(dst ...interface{}) error {
	m.recordCall("AutoMigrates", dst)
	return m.AutoMigratesErr
}

// AutoMigrate simulates Client.AutoMigrate
func (m *MockClient) AutoMigrate(table interface{}) error {
	m.recordCall("AutoMigrate", table)
	return m.AutoMigrateErr
}

// Database simulates Client.Database
func (m *MockClient) Database() *gorm.DB {
	m.recordCall("Database")
	return m.DatabaseRet
}

// SqlDB simulates Client.SqlDB
func (m *MockClient) SqlDB() (*sql.DB, error) {
	m.recordCall("SqlDB")
	return m.SqlDBRet, m.SqlDBErr
}

// DriverType simulates Client.DriverType
func (m *MockClient) DriverType() string {
	m.recordCall("DriverType")
	return m.DriverTypeRet
}

// NewScoop simulates Client.NewScoop
func (m *MockClient) NewScoop() *Scoop {
	m.recordCall("NewScoop")
	return m.ScoopRet
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

// Setup methods

// SetupAutoMigratesSuccess configures AutoMigrates to succeed
func (m *MockClient) SetupAutoMigratesSuccess() *MockClient {
	m.AutoMigratesErr = nil
	return m
}

// SetupAutoMigratesError configures AutoMigrates to return an error
func (m *MockClient) SetupAutoMigratesError(err error) *MockClient {
	m.AutoMigratesErr = err
	return m
}

// SetupAutoMigrateSuccess configures AutoMigrate to succeed
func (m *MockClient) SetupAutoMigrateSuccess() *MockClient {
	m.AutoMigrateErr = nil
	return m
}

// SetupAutoMigrateError configures AutoMigrate to return an error
func (m *MockClient) SetupAutoMigrateError(err error) *MockClient {
	m.AutoMigrateErr = err
	return m
}

// SetupDatabase sets the gorm.DB to return
func (m *MockClient) SetupDatabase(db *gorm.DB) *MockClient {
	m.DatabaseRet = db
	return m
}

// SetupSqlDB sets the sql.DB to return
func (m *MockClient) SetupSqlDB(db *sql.DB) *MockClient {
	m.SqlDBRet = db
	return m
}

// SetupSqlDBError configures SqlDB to return an error
func (m *MockClient) SetupSqlDBError(err error) *MockClient {
	m.SqlDBErr = err
	return m
}

// SetupDriverType sets the driver type to return
func (m *MockClient) SetupDriverType(driverType string) *MockClient {
	m.DriverTypeRet = driverType
	return m
}

// SetupScoop sets the scoop to return
func (m *MockClient) SetupScoop(scoop *Scoop) *MockClient {
	m.ScoopRet = scoop
	return m
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

// MockModel is a mock implementation of Model interface for testing
type MockModel[M any] struct {
	mu sync.RWMutex

	// Return values
	SaveErr      error
	UpdateErr    error
	DeleteErr    error
	FindErr      error
	FindByIDErr  error
	AllErr       error
	CountRet     int64
	CountErr     error
	ExistsRet    bool
	ExistsErr    error
	ScoopRet     *ModelScoop[M]

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

// Setup methods for MockModel

// SetupSaveSuccess configures Save to succeed
func (m *MockModel[M]) SetupSaveSuccess() *MockModel[M] {
	m.SaveErr = nil
	return m
}

// SetupSaveError configures Save to return an error
func (m *MockModel[M]) SetupSaveError(err error) *MockModel[M] {
	m.SaveErr = err
	return m
}

// SetupUpdateSuccess configures Update to succeed
func (m *MockModel[M]) SetupUpdateSuccess() *MockModel[M] {
	m.UpdateErr = nil
	return m
}

// SetupUpdateError configures Update to return an error
func (m *MockModel[M]) SetupUpdateError(err error) *MockModel[M] {
	m.UpdateErr = err
	return m
}

// SetupDeleteSuccess configures Delete to succeed
func (m *MockModel[M]) SetupDeleteSuccess() *MockModel[M] {
	m.DeleteErr = nil
	return m
}

// SetupDeleteError configures Delete to return an error
func (m *MockModel[M]) SetupDeleteError(err error) *MockModel[M] {
	m.DeleteErr = err
	return m
}

// SetupFindSuccess configures Find to succeed
func (m *MockModel[M]) SetupFindSuccess() *MockModel[M] {
	m.FindErr = nil
	return m
}

// SetupFindError configures Find to return an error
func (m *MockModel[M]) SetupFindError(err error) *MockModel[M] {
	m.FindErr = err
	return m
}

// SetupFindByIDSuccess configures FindByID to succeed
func (m *MockModel[M]) SetupFindByIDSuccess() *MockModel[M] {
	m.FindByIDErr = nil
	return m
}

// SetupFindByIDError configures FindByID to return an error
func (m *MockModel[M]) SetupFindByIDError(err error) *MockModel[M] {
	m.FindByIDErr = err
	return m
}

// SetupAllSuccess configures All to succeed
func (m *MockModel[M]) SetupAllSuccess() *MockModel[M] {
	m.AllErr = nil
	return m
}

// SetupAllError configures All to return an error
func (m *MockModel[M]) SetupAllError(err error) *MockModel[M] {
	m.AllErr = err
	return m
}

// SetupCount sets the count to return
func (m *MockModel[M]) SetupCount(count int64) *MockModel[M] {
	m.CountRet = count
	m.CountErr = nil
	return m
}

// SetupCountError configures Count to return an error
func (m *MockModel[M]) SetupCountError(err error) *MockModel[M] {
	m.CountErr = err
	return m
}

// SetupExists sets whether the model exists
func (m *MockModel[M]) SetupExists(exists bool) *MockModel[M] {
	m.ExistsRet = exists
	m.ExistsErr = nil
	return m
}

// SetupExistsError configures Exists to return an error
func (m *MockModel[M]) SetupExistsError(err error) *MockModel[M] {
	m.ExistsErr = err
	return m
}

// SetupScoop sets the scoop to return
func (m *MockModel[M]) SetupScoop(scoop *ModelScoop[M]) *MockModel[M] {
	m.ScoopRet = scoop
	return m
}
