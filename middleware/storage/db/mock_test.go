package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestMockClient_Setup 测试 MockClient 设置
func TestMockClient_Setup(t *testing.T) {
	mockClient := NewMockClient()

	// 设置返回值
	mockClient.
		SetupAutoMigrateSuccess().
		SetupPingSuccess()

	// 执行操作
	err1 := mockClient.AutoMigrate(nil)
	err2 := mockClient.Ping()

	// 验证
	require.NoError(t, err1)
	require.NoError(t, err2)
}

// TestMockClient_CallTracking 测试调用跟踪
func TestMockClient_CallTracking(t *testing.T) {
	mockClient := NewMockClient()

	mockClient.
		SetupAutoMigrateSuccess().
		SetupPingSuccess().
		SetupCloseSuccess()

	// 执行多个操作
	mockClient.AutoMigrate(nil)
	mockClient.AutoMigrate(nil)
	mockClient.Ping()
	mockClient.Close()

	// 验证调用计数
	require.True(t, mockClient.AssertCalled("AutoMigrate"))
	require.True(t, mockClient.AssertCalled("Ping"))
	require.True(t, mockClient.AssertCalled("Close"))
	require.Equal(t, 2, mockClient.GetCallCount("AutoMigrate"))
	require.Equal(t, 1, mockClient.GetCallCount("Ping"))
	require.Equal(t, 1, mockClient.GetCallCount("Close"))

	// 验证未调用的方法
	require.True(t, mockClient.AssertNotCalled("SqlDB"))
}

// TestMockClient_ResetCalls 测试重置调用记录
func TestMockClient_ResetCalls(t *testing.T) {
	mockClient := NewMockClient()

	mockClient.SetupPingSuccess()

	// 执行操作
	mockClient.Ping()
	mockClient.Ping()

	// 验证
	require.Equal(t, 2, mockClient.GetCallCount("Ping"))

	// 重置
	mockClient.ResetCalls()

	// 验证重置
	require.Equal(t, 0, mockClient.GetCallCount("Ping"))
	require.True(t, mockClient.AssertNotCalled("Ping"))
}

// TestMockClient_ErrorHandling 测试错误处理
func TestMockClient_ErrorHandling(t *testing.T) {
	mockClient := NewMockClient()

	// 设置成功
	mockClient.SetupPingSuccess()
	err := mockClient.Ping()
	require.NoError(t, err)

	// 设置错误
	mockClient.SetupPingError(fmt.Errorf("connection failed"))
	err = mockClient.Ping()
	require.Error(t, err)
}

// TestMockModel_SetData 测试 MockModel 数据存储
func TestMockModel_SetData(t *testing.T) {
	mockModel := NewMockModel[interface{}]()

	// 设置数据
	data := map[string]interface{}{
		"id":    uint(1),
		"name":  "Laptop",
		"price": 999.99,
	}

	mockModel.SetData("product:1", data)

	// 获取数据
	retrievedData, ok := mockModel.GetData("product:1")
	require.True(t, ok)
	require.NotNil(t, retrievedData)

	// 验证数据内容
	dataMap := retrievedData.(map[string]interface{})
	require.Equal(t, uint(1), dataMap["id"])
	require.Equal(t, "Laptop", dataMap["name"])
	require.Equal(t, 999.99, dataMap["price"])
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

	data1 := map[string]interface{}{"id": uint(1)}
	data2 := map[string]interface{}{"id": uint(2)}

	mockModel.SetData("prod:1", data1)
	mockModel.SetData("prod:2", data2)

	// 验证都能获取到
	d1, ok1 := mockModel.GetData("prod:1")
	d2, ok2 := mockModel.GetData("prod:2")

	require.True(t, ok1)
	require.True(t, ok2)
	require.Equal(t, uint(1), d1.(map[string]interface{})["id"])
	require.Equal(t, uint(2), d2.(map[string]interface{})["id"])
}

// TestMockModel_SetData_Chainable 测试链式 API
func TestMockModel_SetData_Chainable(t *testing.T) {
	mockModel := NewMockModel[interface{}]()

	data := map[string]interface{}{"name": "Product"}

	// 使用链式 API
	mockModel.
		SetupSaveSuccess().
		SetupCount(int64(1)).
		SetData("product:1", data)

	// 验证数据存储
	retrievedData, ok := mockModel.GetData("product:1")
	require.True(t, ok)
	require.Equal(t, "Product", retrievedData.(map[string]interface{})["name"])
}

// TestMockModel_WithCallTracking 测试与调用跟踪集成
func TestMockModel_WithCallTracking(t *testing.T) {
	mockModel := NewMockModel[interface{}]()

	data := map[string]interface{}{"id": "123"}
	mockModel.SetData("data:123", data)

	// 模拟操作调用
	mockModel.SetupSaveSuccess()
	mockModel.SetupSaveSuccess()

	// 验证调用跟踪
	require.Equal(t, 0, mockModel.GetCallCount("Save"))

	// 验证数据存储
	retrievedData, ok := mockModel.GetData("data:123")
	require.True(t, ok)
	require.NotNil(t, retrievedData)
}

// TestIntegration_DatabaseQueryMock 集成测试 - 模拟数据库查询
func TestIntegration_DatabaseQueryMock(t *testing.T) {
	mockModel := NewMockModel[interface{}]()

	// 设置查询结果
	products := []map[string]interface{}{
		{"id": uint(1), "name": "Laptop", "price": 999.99},
		{"id": uint(2), "name": "Mouse", "price": 29.99},
		{"id": uint(3), "name": "Keyboard", "price": 79.99},
	}

	mockModel.
		SetupFindSuccess().
		SetupCount(int64(len(products))).
		SetupExists(true).
		SetData("all_products", &products)

	// 获取存储的数据
	data, ok := mockModel.GetData("all_products")
	require.True(t, ok)

	productList := data.(*[]map[string]interface{})
	require.Equal(t, 3, len(*productList))
	require.Equal(t, "Laptop", (*productList)[0]["name"])
	require.Equal(t, 999.99, (*productList)[0]["price"])
}

// TestIntegration_ClientAndModel_Combined 集成测试 - 组合使用 MockClient 和 MockModel
func TestIntegration_ClientAndModel_Combined(t *testing.T) {
	// 创建 Mock 客户端
	mockClient := NewMockClient()
	mockClient.
		SetupAutoMigrateSuccess().
		SetupPingSuccess()

	// 创建 Mock 模型
	mockModel := NewMockModel[interface{}]()
	product := map[string]interface{}{"id": uint(1), "name": "Test"}
	mockModel.
		SetupSaveSuccess().
		SetData("test_product", product)

	// 模拟数据库初始化
	err := mockClient.AutoMigrate(nil)
	require.NoError(t, err)

	// 健康检查
	err = mockClient.Ping()
	require.NoError(t, err)

	// 获取模型数据
	data, ok := mockModel.GetData("test_product")
	require.True(t, ok)
	require.Equal(t, "Test", data.(map[string]interface{})["name"])

	// 验证调用
	require.Equal(t, 1, mockClient.GetCallCount("AutoMigrate"))
	require.Equal(t, 1, mockClient.GetCallCount("Ping"))
}
