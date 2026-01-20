package mongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestMockClient_Basic 测试 MockClient 基础功能
func TestMockClient_Basic(t *testing.T) {
	mockClient := NewMockClient()

	// 设置返回值
	mockClient.
		SetupPingSuccess().
		SetupDatabase("testdb")

	// 执行操作
	err := mockClient.Ping()
	db := mockClient.GetDatabase()

	// 验证
	require.NoError(t, err)
	require.Equal(t, "testdb", db)

	// 验证调用
	require.Equal(t, 1, mockClient.GetCallCount("Ping"))
	require.Equal(t, 1, mockClient.GetCallCount("GetDatabase"))
}

// TestMockClient_CallTracking 测试调用跟踪
func TestMockClient_CallTracking(t *testing.T) {
	mockClient := NewMockClient()

	mockClient.
		SetupPingSuccess().
		SetupCloseSuccess()

	// 执行多个操作
	mockClient.Ping()
	mockClient.Ping()
	mockClient.Close()

	// 验证调用
	require.True(t, mockClient.AssertCalled("Ping"))
	require.True(t, mockClient.AssertCalled("Close"))
	require.Equal(t, 2, mockClient.GetCallCount("Ping"))
	require.Equal(t, 1, mockClient.GetCallCount("Close"))

	// 验证未调用的方法
	require.True(t, mockClient.AssertNotCalled("Health"))
}

// TestMockModel_SetData 测试 MockModel 数据存储
func TestMockModel_SetData(t *testing.T) {
	mockModel := NewMockModel[interface{}]()

	// 设置数据
	data := map[string]interface{}{
		"id":   "123",
		"name": "John",
		"age":  30,
	}

	mockModel.SetData("user:123", data)

	// 获取数据
	retrievedData, ok := mockModel.GetData("user:123")
	require.True(t, ok)
	require.NotNil(t, retrievedData)

	// 验证数据内容
	dataMap := retrievedData.(map[string]interface{})
	require.Equal(t, "123", dataMap["id"])
	require.Equal(t, "John", dataMap["name"])
	require.Equal(t, 30, dataMap["age"])
}

// TestMockModel_GetData_NotFound 测试获取不存在的数据
func TestMockModel_GetData_NotFound(t *testing.T) {
	mockModel := NewMockModel[interface{}]()

	_, ok := mockModel.GetData("nonexistent")
	require.False(t, ok)
}

// TestMockModel_SetData_Multiple 测试存储多条数据
func TestMockModel_SetData_Multiple(t *testing.T) {
	mockModel := NewMockModel[interface{}]()

	data1 := map[string]interface{}{"id": "1"}
	data2 := map[string]interface{}{"id": "2"}

	mockModel.SetData("item:1", data1)
	mockModel.SetData("item:2", data2)

	// 验证都能获取到
	d1, ok1 := mockModel.GetData("item:1")
	d2, ok2 := mockModel.GetData("item:2")

	require.True(t, ok1)
	require.True(t, ok2)
	require.Equal(t, "1", d1.(map[string]interface{})["id"])
	require.Equal(t, "2", d2.(map[string]interface{})["id"])
}

// TestMockModel_SetData_Chainable 测试链式 API
func TestMockModel_SetData_Chainable(t *testing.T) {
	mockModel := NewMockModel[interface{}]()

	data := map[string]interface{}{"name": "test"}

	// 使用链式 API
	mockModel.
		SetupCollectionName("test_collection").
		SetupIsNotFound(false).
		SetData("key", data)

	// 执行方法调用
	collName := mockModel.CollectionName()
	isNotFound := mockModel.IsNotFound(nil)

	// 验证返回值
	require.Equal(t, "test_collection", collName)
	require.False(t, isNotFound)

	// 验证调用记录
	require.Equal(t, 1, mockModel.GetCallCount("CollectionName"))
	require.Equal(t, 1, mockModel.GetCallCount("IsNotFound"))

	// 验证数据存储
	retrievedData, ok := mockModel.GetData("key")
	require.True(t, ok)
	require.Equal(t, "test", retrievedData.(map[string]interface{})["name"])
}

// TestMockModel_WithCallTracking 测试与调用跟踪集成
func TestMockModel_WithCallTracking(t *testing.T) {
	mockModel := NewMockModel[interface{}]()

	data := map[string]interface{}{"id": "123"}
	mockModel.SetData("data:123", data)

	// 模拟调用
	mockModel.CollectionName()
	mockModel.CollectionName()
	mockModel.IsNotFound(nil)

	// 验证调用跟踪
	require.Equal(t, 2, mockModel.GetCallCount("CollectionName"))
	require.Equal(t, 1, mockModel.GetCallCount("IsNotFound"))

	// 验证调用记录
	calls := mockModel.GetCalls()
	require.Equal(t, 3, len(calls))
	require.Equal(t, "CollectionName", calls[0].Method)
	require.Equal(t, "IsNotFound", calls[2].Method)
}

// TestIntegration_DataStorageAndTracking 集成测试
func TestIntegration_DataStorageAndTracking(t *testing.T) {
	mockModel := NewMockModel[interface{}]()

	// 设置集合和数据
	mockModel.
		SetupCollectionName("users").
		SetupIsNotFound(false).
		SetData("all_users", []map[string]interface{}{
			{"id": "1", "name": "Alice"},
			{"id": "2", "name": "Bob"},
		})

	// 执行操作
	collName := mockModel.CollectionName()
	isNotFound := mockModel.IsNotFound(nil)

	// 验证返回值
	require.Equal(t, "users", collName)
	require.False(t, isNotFound)

	// 验证存储的数据
	data, ok := mockModel.GetData("all_users")
	require.True(t, ok)

	userList := data.([]map[string]interface{})
	require.Equal(t, 2, len(userList))
	require.Equal(t, "Alice", userList[0]["name"])
	require.Equal(t, "Bob", userList[1]["name"])

	// 验证调用
	require.Equal(t, 1, mockModel.GetCallCount("CollectionName"))
	require.Equal(t, 1, mockModel.GetCallCount("IsNotFound"))
}
