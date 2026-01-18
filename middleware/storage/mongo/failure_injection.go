package mongo

import (
	"errors"
	"sync"
)

// NoopFailureInjector 默认实现 - 不注入任何故障
type NoopFailureInjector struct{}

func (n *NoopFailureInjector) ShouldFailPing() bool                { return false }
func (n *NoopFailureInjector) GetPingError() error                { return nil }
func (n *NoopFailureInjector) ShouldFailFind() bool                { return false }
func (n *NoopFailureInjector) GetFindError() error                { return nil }
func (n *NoopFailureInjector) ShouldFailCount() bool               { return false }
func (n *NoopFailureInjector) GetCountError() error               { return nil }
func (n *NoopFailureInjector) ShouldFailDelete() bool              { return false }
func (n *NoopFailureInjector) GetDeleteError() error              { return nil }
func (n *NoopFailureInjector) ShouldFailTransaction() bool         { return false }
func (n *NoopFailureInjector) GetTransactionError() error         { return nil }
func (n *NoopFailureInjector) ShouldFailWatch() bool               { return false }
func (n *NoopFailureInjector) GetWatchError() error               { return nil }
func (n *NoopFailureInjector) ShouldFailClose() bool               { return false }
func (n *NoopFailureInjector) GetCloseError() error               { return nil }

var defaultInjector FailureInjector = &NoopFailureInjector{}

// SimulatedFailure 模拟故障的类型
type SimulatedFailure struct {
	mu                      sync.RWMutex
	failPing                bool
	pingErr                 error
	failFind                bool
	findErr                 error
	failCount               bool
	countErr                error
	failDelete              bool
	deleteErr               error
	failTransaction         bool
	transactionErr          error
	failWatch               bool
	watchErr                error
	failClose               bool
	closeErr                error
	callCounts              map[string]int
}

// NewSimulatedFailure 创建新的故障注入器
func NewSimulatedFailure() *SimulatedFailure {
	return &SimulatedFailure{
		pingErr:        errors.New("simulated ping failure"),
		findErr:        errors.New("simulated find failure"),
		countErr:       errors.New("simulated count failure"),
		deleteErr:      errors.New("simulated delete failure"),
		transactionErr: errors.New("simulated transaction failure"),
		watchErr:       errors.New("simulated watch failure"),
		closeErr:       errors.New("simulated close failure"),
		callCounts:     make(map[string]int),
	}
}

// FailPing 启用 Ping 故障
func (sf *SimulatedFailure) FailPing(err error) *SimulatedFailure {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.failPing = true
	if err != nil {
		sf.pingErr = err
	}
	return sf
}

// FailFind 启用 Find 故障
func (sf *SimulatedFailure) FailFind(err error) *SimulatedFailure {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.failFind = true
	if err != nil {
		sf.findErr = err
	}
	return sf
}

// FailCount 启用 Count 故障
func (sf *SimulatedFailure) FailCount(err error) *SimulatedFailure {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.failCount = true
	if err != nil {
		sf.countErr = err
	}
	return sf
}

// FailDelete 启用 Delete 故障
func (sf *SimulatedFailure) FailDelete(err error) *SimulatedFailure {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.failDelete = true
	if err != nil {
		sf.deleteErr = err
	}
	return sf
}

// FailTransaction 启用事务故障
func (sf *SimulatedFailure) FailTransaction(err error) *SimulatedFailure {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.failTransaction = true
	if err != nil {
		sf.transactionErr = err
	}
	return sf
}


// FailWatch 启用 Watch 故障
func (sf *SimulatedFailure) FailWatch(err error) *SimulatedFailure {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.failWatch = true
	if err != nil {
		sf.watchErr = err
	}
	return sf
}

// FailClose 启用 Close 故障
func (sf *SimulatedFailure) FailClose(err error) *SimulatedFailure {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.failClose = true
	if err != nil {
		sf.closeErr = err
	}
	return sf
}

// ShouldFailPing 检查是否应该模拟 Ping 失败
func (sf *SimulatedFailure) ShouldFailPing() bool {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.failPing
}

// GetPingError 获取 Ping 错误
func (sf *SimulatedFailure) GetPingError() error {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.pingErr
}

// ShouldFailFind 检查是否应该模拟 Find 失败
func (sf *SimulatedFailure) ShouldFailFind() bool {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.failFind
}

// GetFindError 获取 Find 错误
func (sf *SimulatedFailure) GetFindError() error {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.findErr
}

// ShouldFailCount 检查是否应该模拟 Count 失败
func (sf *SimulatedFailure) ShouldFailCount() bool {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.failCount
}

// GetCountError 获取 Count 错误
func (sf *SimulatedFailure) GetCountError() error {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.countErr
}

// ShouldFailDelete 检查是否应该模拟 Delete 失败
func (sf *SimulatedFailure) ShouldFailDelete() bool {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.failDelete
}

// GetDeleteError 获取 Delete 错误
func (sf *SimulatedFailure) GetDeleteError() error {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.deleteErr
}

// ShouldFailTransaction 检查是否应该模拟事务失败
func (sf *SimulatedFailure) ShouldFailTransaction() bool {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.failTransaction
}

// GetTransactionError 获取事务错误
func (sf *SimulatedFailure) GetTransactionError() error {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.transactionErr
}


// ShouldFailWatch 检查是否应该模拟 Watch 失败
func (sf *SimulatedFailure) ShouldFailWatch() bool {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.failWatch
}

// GetWatchError 获取 Watch 错误
func (sf *SimulatedFailure) GetWatchError() error {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.watchErr
}

// ShouldFailClose 检查是否应该模拟 Close 失败
func (sf *SimulatedFailure) ShouldFailClose() bool {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.failClose
}

// GetCloseError 获取 Close 错误
func (sf *SimulatedFailure) GetCloseError() error {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.closeErr
}

// Reset 重置所有故障
func (sf *SimulatedFailure) Reset() *SimulatedFailure {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.failPing = false
	sf.failFind = false
	sf.failCount = false
	sf.failDelete = false
	sf.failTransaction = false
	sf.failWatch = false
	sf.failClose = false
	sf.callCounts = make(map[string]int)
	return sf
}

// RecordCall 记录函数调用（用于验证）
func (sf *SimulatedFailure) RecordCall(funcName string) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.callCounts[funcName]++
}

// GetCallCount 获取函数调用次数
func (sf *SimulatedFailure) GetCallCount(funcName string) int {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.callCounts[funcName]
}

// SetGlobalInjector 设置全局故障注入器（用于测试）
func SetGlobalInjector(injector FailureInjector) {
	if injector != nil {
		defaultInjector = injector
	} else {
		defaultInjector = &NoopFailureInjector{}
	}
}

// GetGlobalInjector 获取全局故障注入器
func GetGlobalInjector() FailureInjector {
	return defaultInjector
}

// ResetGlobalInjector 重置全局故障注入器
func ResetGlobalInjector() {
	defaultInjector = &NoopFailureInjector{}
}
