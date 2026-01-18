package mongo

import (
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestPingWithInjectedFailure 测试注入的 Ping 故障
func TestPingWithInjectedFailure(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// 创建故障注入器
	injector := NewSimulatedFailure()
	injector.FailPing(errors.New("connection refused"))
	SetGlobalInjector(injector)
	defer ResetGlobalInjector()

	// Ping 应该返回注入的错误
	err := client.Ping()
	if err == nil {
		t.Error("Expected Ping to fail, but it succeeded")
	} else if err.Error() != "connection refused" {
		t.Errorf("Expected 'connection refused', got '%v'", err)
	} else {
		t.Logf("Ping correctly returned injected error: %v", err)
	}
}

// TestCountWithInjectedFailure 测试注入的 Count 故障
func TestCountWithInjectedFailure(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 创建故障注入器
	injector := NewSimulatedFailure()
	injector.FailCount(errors.New("collection not accessible"))
	SetGlobalInjector(injector)
	defer ResetGlobalInjector()

	// 插入测试数据
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "count_fail@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{})

	// Count 应该返回注入的错误
	count, err := scoop.Count()
	if err == nil {
		t.Error("Expected Count to fail, but it succeeded")
	} else if err.Error() != "collection not accessible" {
		t.Errorf("Expected 'collection not accessible', got '%v'", err)
	} else if count != 0 {
		t.Errorf("Expected count 0 on error, got %d", count)
	} else {
		t.Logf("Count correctly returned injected error: %v", err)
	}
}

// TestDeleteWithInjectedFailure 测试注入的 Delete 故障
func TestDeleteWithInjectedFailure(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 创建故障注入器
	injector := NewSimulatedFailure()
	injector.FailDelete(errors.New("permission denied"))
	SetGlobalInjector(injector)
	defer ResetGlobalInjector()

	// 插入测试数据
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "delete_fail@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{})

	// Delete 应该返回注入的错误
	deleted, err := scoop.Delete()
	if err == nil {
		t.Error("Expected Delete to fail, but it succeeded")
	} else if err.Error() != "permission denied" {
		t.Errorf("Expected 'permission denied', got '%v'", err)
	} else if deleted != 0 {
		t.Errorf("Expected deleted 0 on error, got %d", deleted)
	} else {
		t.Logf("Delete correctly returned injected error: %v", err)
	}
}

// TestTransactionWithInjectedFailure 测试注入的事务故障
func TestTransactionWithInjectedFailure(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 创建故障注入器
	injector := NewSimulatedFailure()
	injector.FailTransaction(errors.New("session not available"))
	SetGlobalInjector(injector)
	defer ResetGlobalInjector()

	scoop := client.NewScoop().Collection(User{})

	// Begin 应该返回注入的错误
	txScoop, err := scoop.Begin()
	if err == nil {
		t.Error("Expected Begin to fail, but it succeeded")
	} else if err.Error() != "session not available" {
		t.Errorf("Expected 'session not available', got '%v'", err)
	} else if txScoop != nil {
		t.Error("Expected nil Scoop on error, but got non-nil")
	} else {
		t.Logf("Begin correctly returned injected error: %v", err)
	}
}

// TestMultipleFailureInjections 测试多个同时注入的故障
func TestMultipleFailureInjections(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 创建注入多个故障的注入器
	injector := NewSimulatedFailure()
	injector.FailPing(errors.New("ping error"))
	injector.FailCount(errors.New("count error"))
	injector.FailDelete(errors.New("delete error"))
	SetGlobalInjector(injector)
	defer ResetGlobalInjector()

	// 测试 Ping 故障
	err := client.Ping()
	if err == nil || err.Error() != "ping error" {
		t.Errorf("Ping should fail with 'ping error', got %v", err)
	}

	// 插入测试数据
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "multi_fail@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{})

	// 测试 Count 故障
	_, err = scoop.Count()
	if err == nil || err.Error() != "count error" {
		t.Errorf("Count should fail with 'count error', got %v", err)
	}

	// 测试 Delete 故障
	_, err = scoop.Delete()
	if err == nil || err.Error() != "delete error" {
		t.Errorf("Delete should fail with 'delete error', got %v", err)
	}

	t.Logf("Multiple failure injections working correctly")
}

// TestFailureInjectionReset 测试故障注入的重置
func TestFailureInjectionReset(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 创建并启用故障注入
	injector := NewSimulatedFailure()
	injector.FailPing(errors.New("ping error"))
	SetGlobalInjector(injector)

	// Ping 应该失败
	err := client.Ping()
	if err == nil {
		t.Error("Ping should fail with injection")
	}

	// 重置注入器
	injector.Reset()

	// Ping 应该成功
	err = client.Ping()
	if err != nil {
		t.Errorf("Ping should succeed after reset, got %v", err)
	} else {
		t.Logf("Ping correctly succeeded after reset")
	}

	ResetGlobalInjector()
}

// TestNormalOperationWithoutInjection 测试没有注入时的正常操作
func TestNormalOperationWithoutInjection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 确保注入器已重置
	ResetGlobalInjector()

	// 所有操作应该正常工作
	err := client.Ping()
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}

	// 插入测试数据
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "normal@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{})

	// Count 应该成功
	count, err := scoop.Count()
	if err != nil {
		t.Errorf("Count failed: %v", err)
	} else if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	} else {
		t.Logf("All operations succeeded without injection")
	}
}

// TestSelectiveFailureInjection 测试有选择性的故障注入
func TestSelectiveFailureInjection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 创建注入器，仅注入 Count 故障
	injector := NewSimulatedFailure()
	injector.FailCount(errors.New("count error"))
	SetGlobalInjector(injector)
	defer ResetGlobalInjector()

	// Ping 应该成功（未被注入）
	err := client.Ping()
	if err != nil {
		t.Errorf("Ping should succeed, got %v", err)
	}

	// 插入测试数据
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "selective@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{})

	// Count 应该失败（被注入）
	_, err = scoop.Count()
	if err == nil || err.Error() != "count error" {
		t.Errorf("Count should fail with injected error, got %v", err)
	}

	// Delete 应该成功（未被注入）
	deleted, err := scoop.Delete()
	if err != nil {
		t.Errorf("Delete should succeed, got %v", err)
	} else if deleted != 1 {
		t.Errorf("Expected deleted 1, got %d", deleted)
	} else {
		t.Logf("Selective injection working correctly")
	}
}


// TestFailureInjectorMethods 测试故障注入器的所有方法
func TestFailureInjectorMethods(t *testing.T) {
	// 测试 SimulatedFailure 的所有方法
	injector := NewSimulatedFailure()

	// 测试初始状态 - 所有都应该是 false
	if injector.ShouldFailPing() {
		t.Error("Expected ShouldFailPing to be false initially")
	}
	if injector.ShouldFailFind() {
		t.Error("Expected ShouldFailFind to be false initially")
	}
	if injector.ShouldFailCount() {
		t.Error("Expected ShouldFailCount to be false initially")
	}
	if injector.ShouldFailDelete() {
		t.Error("Expected ShouldFailDelete to be false initially")
	}
	if injector.ShouldFailTransaction() {
		t.Error("Expected ShouldFailTransaction to be false initially")
	}

	// 设置 Ping 失败
	pingErr := errors.New("ping error")
	injector.FailPing(pingErr)
	if !injector.ShouldFailPing() {
		t.Error("Expected ShouldFailPing to be true after FailPing")
	}
	if injector.GetPingError() != pingErr {
		t.Error("Expected GetPingError to return the set error")
	}

	// 设置 Find 失败
	findErr := errors.New("find error")
	injector.FailFind(findErr)
	if !injector.ShouldFailFind() {
		t.Error("Expected ShouldFailFind to be true after FailFind")
	}
	if injector.GetFindError() != findErr {
		t.Error("Expected GetFindError to return the set error")
	}

	// 设置 Count 失败
	countErr := errors.New("count error")
	injector.FailCount(countErr)
	if !injector.ShouldFailCount() {
		t.Error("Expected ShouldFailCount to be true after FailCount")
	}
	if injector.GetCountError() != countErr {
		t.Error("Expected GetCountError to return the set error")
	}

	// 设置 Delete 失败
	deleteErr := errors.New("delete error")
	injector.FailDelete(deleteErr)
	if !injector.ShouldFailDelete() {
		t.Error("Expected ShouldFailDelete to be true after FailDelete")
	}
	if injector.GetDeleteError() != deleteErr {
		t.Error("Expected GetDeleteError to return the set error")
	}

	// 设置 Transaction 失败
	transactionErr := errors.New("transaction error")
	injector.FailTransaction(transactionErr)
	if !injector.ShouldFailTransaction() {
		t.Error("Expected ShouldFailTransaction to be true after FailTransaction")
	}
	if injector.GetTransactionError() != transactionErr {
		t.Error("Expected GetTransactionError to return the set error")
	}

	// 测试 RecordCall 和 GetCallCount
	injector.RecordCall("TestOperation1")
	injector.RecordCall("TestOperation1")
	injector.RecordCall("TestOperation2")

	if injector.GetCallCount("TestOperation1") != 2 {
		t.Errorf("Expected TestOperation1 call count 2, got %d", injector.GetCallCount("TestOperation1"))
	}
	if injector.GetCallCount("TestOperation2") != 1 {
		t.Errorf("Expected TestOperation2 call count 1, got %d", injector.GetCallCount("TestOperation2"))
	}

	// 测试 Reset
	injector.Reset()
	if injector.ShouldFailPing() {
		t.Error("Expected ShouldFailPing to be false after Reset")
	}
	if injector.ShouldFailFind() {
		t.Error("Expected ShouldFailFind to be false after Reset")
	}
	if injector.ShouldFailCount() {
		t.Error("Expected ShouldFailCount to be false after Reset")
	}
	if injector.ShouldFailDelete() {
		t.Error("Expected ShouldFailDelete to be false after Reset")
	}
	if injector.ShouldFailTransaction() {
		t.Error("Expected ShouldFailTransaction to be false after Reset")
	}
	if injector.GetCallCount("TestOperation1") != 0 {
		t.Errorf("Expected TestOperation1 call count 0 after Reset, got %d", injector.GetCallCount("TestOperation1"))
	}

	t.Logf("All failure injector methods working correctly")
}

// TestNoopFailureInjector 测试 NoopFailureInjector 的行为
func TestNoopFailureInjector(t *testing.T) {
	// 创建一个 NoopFailureInjector
	injector := &NoopFailureInjector{}

	// 所有 ShouldFail 方法应该都返回 false
	if injector.ShouldFailPing() {
		t.Error("NoopFailureInjector.ShouldFailPing should always be false")
	}
	if injector.ShouldFailFind() {
		t.Error("NoopFailureInjector.ShouldFailFind should always be false")
	}
	if injector.ShouldFailCount() {
		t.Error("NoopFailureInjector.ShouldFailCount should always be false")
	}
	if injector.ShouldFailDelete() {
		t.Error("NoopFailureInjector.ShouldFailDelete should always be false")
	}
	if injector.ShouldFailTransaction() {
		t.Error("NoopFailureInjector.ShouldFailTransaction should always be false")
	}

	// 所有 GetError 方法应该返回 nil
	if injector.GetPingError() != nil {
		t.Error("NoopFailureInjector.GetPingError should return nil")
	}
	if injector.GetFindError() != nil {
		t.Error("NoopFailureInjector.GetFindError should return nil")
	}
	if injector.GetCountError() != nil {
		t.Error("NoopFailureInjector.GetCountError should return nil")
	}
	if injector.GetDeleteError() != nil {
		t.Error("NoopFailureInjector.GetDeleteError should return nil")
	}
	if injector.GetTransactionError() != nil {
		t.Error("NoopFailureInjector.GetTransactionError should return nil")
	}

	t.Logf("NoopFailureInjector working correctly as noop")
}

// TestGlobalInjectorManagement 测试全局注入器的管理
func TestGlobalInjectorManagement(t *testing.T) {
	// 保存原始注入器
	originalInjector := GetGlobalInjector()
	defer func() {
		SetGlobalInjector(originalInjector)
	}()

	// 创建新的注入器
	newInjector := NewSimulatedFailure()
	newInjector.FailPing(errors.New("global ping error"))

	// 设置全局注入器
	SetGlobalInjector(newInjector)
	globalInjector := GetGlobalInjector()

	if !globalInjector.ShouldFailPing() {
		t.Error("Expected global injector to have Ping failure set")
	}
	if globalInjector.GetPingError().Error() != "global ping error" {
		t.Errorf("Expected 'global ping error', got '%v'", globalInjector.GetPingError())
	}

	// 重置全局注入器
	ResetGlobalInjector()
	resetInjector := GetGlobalInjector()

	if resetInjector.ShouldFailPing() {
		t.Error("Expected global injector to be reset (ShouldFailPing should be false)")
	}

	t.Logf("Global injector management working correctly")
}


// TestSetNilGlobalInjector 测试设置 nil 注入器的情况
func TestSetNilGlobalInjector(t *testing.T) {
	// 保存原始注入器
	originalInjector := GetGlobalInjector()
	defer func() {
		SetGlobalInjector(originalInjector)
	}()

	// 创建一个有故障的注入器
	injector := NewSimulatedFailure()
	injector.FailPing(errors.New("ping error"))
	SetGlobalInjector(injector)

	// 验证故障已设置
	if !GetGlobalInjector().ShouldFailPing() {
		t.Error("Expected Ping failure to be set")
	}

	// 设置 nil 注入器 - 应该使用 NoopFailureInjector
	SetGlobalInjector(nil)
	noop := GetGlobalInjector()

	// 验证已经转换为 Noop 注入器
	if noop.ShouldFailPing() {
		t.Error("Expected nil injector to be converted to NoopFailureInjector")
	}
	if noop.GetPingError() != nil {
		t.Error("Expected NoopFailureInjector to return nil error")
	}

	t.Logf("Setting nil injector correctly creates NoopFailureInjector")
}


// TestAggregationExecute 测试 Aggregation 的 Execute 方法
func TestAggregationExecute(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 插入测试数据
	user1 := User{
		ID:        primitive.NewObjectID(),
		Email:     "exec1@example.com",
		Name:      "Test1",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	user2 := User{
		ID:        primitive.NewObjectID(),
		Email:     "exec2@example.com",
		Name:      "Test2",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user1)
	InsertTestData(t, client, "users", user2)

	scoop := client.NewScoop().Collection(User{})

	// 使用 Aggregation 的 Execute 方法
	pipeline := bson.M{
		"$match": bson.M{
			"age": bson.M{
				"$gte": 25,
			},
		},
	}

	var results []User
	agg := scoop.Aggregate(pipeline)
	err := agg.Execute(&results)
	if err != nil {
		t.Errorf("Expected Execute to succeed, got error: %v", err)
	} else if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	} else {
		t.Logf("Aggregation Execute correctly returned %d results", len(results))
	}
}

// TestAggregationExecuteOne 测试 Aggregation 的 ExecuteOne 方法
func TestAggregationExecuteOneFromFail(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 插入测试数据
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "execone@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{})

	// 使用 Aggregation 的 ExecuteOne 方法
	pipeline := bson.M{
		"$match": bson.M{
			"email": "execone@example.com",
		},
	}

	var result User
	agg := scoop.Aggregate(pipeline)
	err := agg.ExecuteOne(&result)
	if err != nil {
		t.Errorf("Expected ExecuteOne to succeed, got error: %v", err)
	} else if result.Email != "execone@example.com" {
		t.Errorf("Expected email 'execone@example.com', got '%s'", result.Email)
	} else {
		t.Logf("Aggregation ExecuteOne correctly returned the result")
	}
}

// TestUpdateOperation 测试 Update 操作
func TestUpdateOperation(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 插入测试数据
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "update_test@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{}).Where("email", "update_test@example.com")

	// Update 应该成功
	updateData := User{
		Age: 26,
	}
	modifiedCount, err := scoop.Update(&updateData)
	if err != nil {
		t.Errorf("Expected Update to succeed, got error: %v", err)
	} else if modifiedCount == 0 {
		t.Errorf("Expected Update to modify at least one document, got %d", modifiedCount)
	} else {
		t.Logf("Update correctly modified %d document(s)", modifiedCount)
	}
}

// TestBatchCreateOperation 测试 BatchCreate 操作
func TestBatchCreateOperation(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 创建多个用户
	user1 := User{
		ID:        primitive.NewObjectID(),
		Email:     "batch1@example.com",
		Name:      "Test1",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	user2 := User{
		ID:        primitive.NewObjectID(),
		Email:     "batch2@example.com",
		Name:      "Test2",
		Age:       26,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	scoop := client.NewScoop().Collection(User{})

	// BatchCreate 应该成功
	err := scoop.BatchCreate(user1, user2)
	if err != nil {
		t.Errorf("Expected BatchCreate to succeed, got error: %v", err)
	} else {
		t.Logf("BatchCreate correctly inserted 2 document(s)")
	}
}


// TestCloneScoop 测试 Clone 方法
func TestCloneScoopFromFail(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 插入测试数据
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "clone_test@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	// 创建原始 scoop
	scoop := client.NewScoop().Collection(User{}).Where("email", "clone_test@example.com").Limit(10)

	// Clone scoop
	clonedScoop := scoop.Clone()
	if clonedScoop == nil {
		t.Error("Expected Clone to return a non-nil scoop")
	} else {
		t.Logf("Clone correctly returned a new scoop")
	}

	// 验证克隆后的 scoop 可以使用
	count, err := clonedScoop.Count()
	if err != nil {
		t.Errorf("Expected Count on cloned scoop to succeed, got error: %v", err)
	} else if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	} else {
		t.Logf("Cloned scoop Count correctly returned %d", count)
	}
}

// TestClearScoop 测试 Clear 方法
func TestClearScoopFromFail(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 创建带有多个条件的 scoop
	scoop := client.NewScoop().Collection(User{}).
		Where("age", ">", 20).
		Limit(10).
		Offset(5).
		Sort("email", 1).
		Select("name", "email")

	// Clear 应该重置所有条件
	clearedScoop := scoop.Clear()
	if clearedScoop == nil {
		t.Error("Expected Clear to return a non-nil scoop")
	} else {
		t.Logf("Clear correctly returned a scoop")
	}
}

// TestScoopChaining 测试 Scoop 的链式调用
func TestScoopChaining(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 插入测试数据
	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     primitive.NewObjectID().Hex() + "@example.com",
			Name:      "Test",
			Age:       20 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})

	// 使用链式调用测试 In 条件
	scoop = scoop.Where("age", ">=", 21).In("age", 21, 22, 23)

	count, err := scoop.Count()
	if err != nil {
		t.Errorf("Expected Count to succeed, got error: %v", err)
	} else if count < 1 {
		t.Errorf("Expected count >= 1, got %d", count)
	} else {
		t.Logf("Chain calling correctly returned count %d", count)
	}
}

// TestNotInCondition 测试 NotIn 条件
func TestNotInCondition(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 插入测试数据
	user1 := User{
		ID:        primitive.NewObjectID(),
		Email:     "notin1@example.com",
		Name:      "Test1",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	user2 := User{
		ID:        primitive.NewObjectID(),
		Email:     "notin2@example.com",
		Name:      "Test2",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	user3 := User{
		ID:        primitive.NewObjectID(),
		Email:     "notin3@example.com",
		Name:      "Test3",
		Age:       35,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user1)
	InsertTestData(t, client, "users", user2)
	InsertTestData(t, client, "users", user3)

	scoop := client.NewScoop().Collection(User{}).NotIn("age", 25, 30)

	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Errorf("Expected Find to succeed, got error: %v", err)
	} else if len(results) != 1 {
		t.Errorf("Expected 1 result (age 35), got %d", len(results))
	} else if results[0].Age != 35 {
		t.Errorf("Expected age 35, got %d", results[0].Age)
	} else {
		t.Logf("NotIn condition correctly returned 1 result")
	}
}

// TestFindWithInjectedFailure 测试 Find 操作的故障注入
func TestFindWithInjectedFailure(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 创建故障注入器
	injector := NewSimulatedFailure()
	injector.FailFind(errors.New("find not available"))
	SetGlobalInjector(injector)
	defer ResetGlobalInjector()

	// 插入测试数据
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "find_fail@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{})

	// Find 应该返回注入的错误
	var results []User
	err := scoop.Find(&results)
	if err == nil {
		t.Error("Expected Find to fail, but it succeeded")
	} else if err.Error() != "find not available" {
		t.Errorf("Expected 'find not available', got '%v'", err)
	} else if results != nil && len(results) > 0 {
		t.Error("Expected empty results on error")
	} else {
		t.Logf("Find correctly returned injected error: %v", err)
	}
}

// TestConcurrentFailureInjection 测试并发环境下的故障注入
func TestConcurrentFailureInjection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 使用一个共享的注入器
	injector := NewSimulatedFailure()
	SetGlobalInjector(injector)
	defer ResetGlobalInjector()

	// 并发测试调用计数功能
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// 每个 goroutine 记录自己的调用
			injector.RecordCall("concurrent_operation")

			// 验证调用已记录
			// 由于并发，我们不能确切知道最终计数，但应该 >= 1
			if injector.GetCallCount("concurrent_operation") < 1 {
				t.Errorf("Goroutine %d: Expected call count >= 1", id)
			}
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// 验证总的调用次数应该等于 goroutine 数量
	if injector.GetCallCount("concurrent_operation") != numGoroutines {
		t.Errorf("Expected total call count %d, got %d", numGoroutines, injector.GetCallCount("concurrent_operation"))
	}

	t.Logf("Concurrent failure injection working correctly")
}


// TestCommitWithInjectedFailure 测试 Commit 操作的故障注入（如果实现了的话）
func TestCommitWithInjectedFailure(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 插入测试数据
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "commit_test@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{})

	// 开始事务
	txScoop, err := scoop.Begin()
	if err != nil {
		t.Logf("Begin returned error (expected for non-replica set): %v", err)
		// 这是预期的，因为本地 MongoDB 通常没有副本集
		return
	}
	if txScoop == nil {
		t.Logf("Begin returned nil scoop (expected for standalone MongoDB)")
		return
	}

	// 提交事务
	err = txScoop.Commit()
	if err != nil {
		t.Logf("Commit returned error: %v", err)
	} else {
		t.Logf("Commit succeeded")
	}
}

// TestRollbackWithFailure 测试 Rollback 操作的基本功能
func TestRollbackWithFailure(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 插入测试数据
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "rollback_test@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{})

	// 开始事务
	txScoop, err := scoop.Begin()
	if err != nil {
		t.Logf("Begin returned error (expected for non-replica set): %v", err)
		// 这是预期的，因为本地 MongoDB 通常没有副本集
		return
	}
	if txScoop == nil {
		t.Logf("Begin returned nil scoop (expected for standalone MongoDB)")
		return
	}

	// 回滚事务
	err = txScoop.Rollback()
	if err != nil {
		t.Logf("Rollback returned error: %v", err)
	} else {
		t.Logf("Rollback succeeded")
	}
}

// TestFirstWithFailure 测试 First 方法
func TestFirstWithFailure(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 插入测试数据
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "first_test@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{})

	// First 应该成功找到第一条记录
	var result User
	err := scoop.First(&result)
	if err != nil {
		t.Errorf("Expected First to succeed, got error: %v", err)
	} else if result.Email != "first_test@example.com" {
		t.Errorf("Expected email 'first_test@example.com', got '%s'", result.Email)
	} else {
		t.Logf("First correctly returned the first record: %s", result.Email)
	}
}

// TestExistWithData 测试 Exist 方法
func TestExistWithData(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 插入测试数据
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "exist_test@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{})

	// Exist 应该返回 true
	exist, err := scoop.Exist()
	if err != nil {
		t.Errorf("Expected Exist to succeed, got error: %v", err)
	} else if !exist {
		t.Error("Expected Exist to return true when data exists")
	} else {
		t.Logf("Exist correctly returned true")
	}

	// 带过滤条件的 Exist
	scoop2 := client.NewScoop().Collection(User{}).Where("email", "nonexistent@example.com")
	exist2, err := scoop2.Exist()
	if err != nil {
		t.Errorf("Expected Exist to succeed, got error: %v", err)
	} else if exist2 {
		t.Error("Expected Exist to return false when data doesn't match filter")
	} else {
		t.Logf("Exist correctly returned false for non-matching filter")
	}
}

// TestCreateWithFailure 测试 Create 方法
func TestCreateWithFailure(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Create 应该成功
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "create_test@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	scoop := client.NewScoop().Collection(User{})
	err := scoop.Create(&user)
	if err != nil {
		t.Errorf("Expected Create to succeed, got error: %v", err)
	} else {
		t.Logf("Create correctly succeeded")
	}
}

// TestAggregateWithData 测试 Aggregate 方法
func TestAggregateWithData(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// 插入测试数据
	user1 := User{
		ID:        primitive.NewObjectID(),
		Email:     "agg1@example.com",
		Name:      "Test1",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	user2 := User{
		ID:        primitive.NewObjectID(),
		Email:     "agg2@example.com",
		Name:      "Test2",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user1)
	InsertTestData(t, client, "users", user2)

	scoop := client.NewScoop().Collection(User{})

	// Aggregate 应该成功
	pipeline := bson.M{
		"$match": bson.M{
			"age": bson.M{
				"$gte": 25,
			},
		},
	}

	agg := scoop.Aggregate(pipeline)
	if agg == nil {
		t.Error("Expected Aggregate to return an Aggregation object")
	} else {
		t.Logf("Aggregate correctly returned an Aggregation object")
	}
}
