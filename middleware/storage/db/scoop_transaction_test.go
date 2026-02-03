package db

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestScoop_TransactionCommit 测试事务提交
func TestScoop_TransactionCommit(t *testing.T) {
	tmpDir := t.TempDir()
	config := &Config{
		Type:    Sqlite,
		Address: "file:" + tmpDir,
		Name:    "test_tx_commit",
	}

	client, err := New(config, TestUser{})
	assert.NoError(t, err)
	defer client.Close()

	// 开始事务
	tx := client.NewScoop().Begin()
	assert.NotNil(t, tx)

	// 在事务中创建记录
	result := tx.Model(TestUser{}).Create(&TestUser{
		Name: "Transaction User",
	})
	assert.NoError(t, result.Error)

	// 提交事务
	tx.Commit()

	// 验证记录已保存
	var user TestUser
	firstResult := client.NewScoop().Model(TestUser{}).Where("name", "Transaction User").First(&user)
	assert.NoError(t, firstResult.Error)
	assert.Equal(t, "Transaction User", user.Name)
}

// TestScoop_TransactionRollback 测试事务回滚
func TestScoop_TransactionRollback(t *testing.T) {
	tmpDir := t.TempDir()
	config := &Config{
		Type:    Sqlite,
		Address: "file:" + tmpDir,
		Name:    "test_tx_rollback",
	}

	client, err := New(config, TestUser{})
	assert.NoError(t, err)
	defer client.Close()

	// 开始事务
	tx := client.NewScoop().Begin()
	assert.NotNil(t, tx)

	// 在事务中创建记录
	result := tx.Model(TestUser{}).Create(&TestUser{
		Name: "Rollback User",
	})
	assert.NoError(t, result.Error)

	// 回滚事务
	tx.Rollback()

	// 验证记录未保存
	count, err := client.NewScoop().Model(TestUser{}).Where("name", "Rollback User").Count()
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), count)
}

// TestScoop_TransactionCommitOrRollback 测试 CommitOrRollback
func TestScoop_TransactionCommitOrRollback(t *testing.T) {
	tmpDir := t.TempDir()
	config := &Config{
		Type:    Sqlite,
		Address: "file:" + tmpDir,
		Name:    "test_tx_cor",
	}

	client, err := New(config, TestUser{})
	assert.NoError(t, err)
	defer client.Close()

	t.Run("commit on success", func(t *testing.T) {
		tx := client.NewScoop().Begin()

		err := tx.CommitOrRollback(tx, func(tx *Scoop) error {
			result := tx.Model(TestUser{}).Create(&TestUser{
				Name: "Success User",
			})
			return result.Error
		})
		assert.NoError(t, err)

		// 验证记录已保存
		var user TestUser
		firstResult := client.NewScoop().Model(TestUser{}).Where("name", "Success User").First(&user)
		assert.NoError(t, firstResult.Error)
		assert.Equal(t, "Success User", user.Name)
	})

	t.Run("rollback on error", func(t *testing.T) {
		tx := client.NewScoop().Begin()

		testErr := errors.New("test error")
		err := tx.CommitOrRollback(tx, func(tx *Scoop) error {
			result := tx.Model(TestUser{}).Create(&TestUser{
				Name: "Error User",
			})
			if result.Error != nil {
				return result.Error
			}
			return testErr // 模拟错误
		})
		assert.Error(t, err)
		assert.Equal(t, testErr, err)

		// 验证记录未保存
		count, err := client.NewScoop().Model(TestUser{}).Where("name", "Error User").Count()
		assert.NoError(t, err)
		assert.Equal(t, uint64(0), count)
	})
}

// TestScoop_NestedTransaction 测试嵌套事务
func TestScoop_NestedTransaction(t *testing.T) {
	tmpDir := t.TempDir()
	config := &Config{
		Type:    Sqlite,
		Address: "file:" + tmpDir,
		Name:    "test_nested_tx",
	}

	client, err := New(config, TestUser{})
	assert.NoError(t, err)
	defer client.Close()

	// 外层事务
	tx1 := client.NewScoop().Begin()
	assert.NotNil(t, tx1)

	result := tx1.Model(TestUser{}).Create(&TestUser{
		Name: "Outer Transaction",
	})
	assert.NoError(t, result.Error)

	// 内层事务（使用同一个事务对象）
	tx2 := tx1.Begin()
	assert.NotNil(t, tx2)

	result = tx2.Model(TestUser{}).Create(&TestUser{
		Name: "Inner Transaction",
	})
	assert.NoError(t, result.Error)

	// 提交内层事务
	tx2.Commit()

	// 提交外层事务
	tx1.Commit()

	// 验证两条记录都已保存
	count, err := client.NewScoop().Model(TestUser{}).Count()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, uint64(2))
}
