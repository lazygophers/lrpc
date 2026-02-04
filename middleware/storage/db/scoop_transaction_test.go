package db_test

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestScoop_TransactionCommit 测试事务提交
func TestScoop_TransactionCommit(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer func() {
		mockDB.Mock.ExpectClose()
		mockDB.Close()
	}()

	t.Run("commit successful transaction", func(t *testing.T) {
		mockDB.Mock.ExpectBegin()
		mockDB.Mock.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mockDB.Mock.ExpectCommit()

		tx := client.NewScoop().Begin()
		assert.NotNil(t, tx)

		result := tx.Model(TestUser{}).Create(&TestUser{
			Name:  "Transaction User",
			Email: "tx@example.com",
		})
		assert.NoError(t, result.Error)

		tx.Commit()
	})

	t.Run("commit with error in transaction", func(t *testing.T) {
		mockDB.Mock.ExpectBegin()
		mockDB.Mock.ExpectExec("INSERT INTO test_users").
			WillReturnError(errors.New("transaction error"))
		mockDB.Mock.ExpectRollback()

		tx := client.NewScoop().Begin()
		assert.NotNil(t, tx)

		result := tx.Model(TestUser{}).Create(&TestUser{
			Name:  "Error User",
			Email: "error@example.com",
		})
		assert.Error(t, result.Error)

		tx.Rollback()
	})
}

// TestScoop_TransactionRollback 测试事务回滚
func TestScoop_TransactionRollback(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer func() {
		mockDB.Mock.ExpectClose()
		mockDB.Close()
	}()

	t.Run("rollback transaction", func(t *testing.T) {
		mockDB.Mock.ExpectBegin()
		mockDB.Mock.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mockDB.Mock.ExpectRollback()

		tx := client.NewScoop().Begin()
		assert.NotNil(t, tx)

		result := tx.Model(TestUser{}).Create(&TestUser{
			Name: "Rollback User",
		})
		assert.NoError(t, result.Error)

		tx.Rollback()
	})
}

// TestScoop_TransactionCommitOrRollback 测试 CommitOrRollback
func TestScoop_TransactionCommitOrRollback(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer func() {
		mockDB.Mock.ExpectClose()
		mockDB.Close()
	}()

	t.Run("commit on success", func(t *testing.T) {
		mockDB.Mock.ExpectBegin()
		mockDB.Mock.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mockDB.Mock.ExpectCommit()

		tx := client.NewScoop().Begin()

		err := tx.CommitOrRollback(tx, func(tx *db.Scoop) error {
			result := tx.Model(TestUser{}).Create(&TestUser{
				Name:  "Success User",
				Email: "success@example.com",
			})
			return result.Error
		})
		assert.NoError(t, err)
	})

	t.Run("rollback on error", func(t *testing.T) {
		mockDB.Mock.ExpectBegin()
		mockDB.Mock.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mockDB.Mock.ExpectRollback()

		tx := client.NewScoop().Begin()

		testErr := errors.New("test error")
		err := tx.CommitOrRollback(tx, func(tx *db.Scoop) error {
			result := tx.Model(TestUser{}).Create(&TestUser{
				Name:  "Error User",
				Email: "error@example.com",
			})
			if result.Error != nil {
				return result.Error
			}
			return testErr // 模拟错误
		})
		assert.Error(t, err)
		assert.Equal(t, testErr, err)
	})
}

// TestScoop_NestedTransaction 测试嵌套事务
func TestScoop_NestedTransaction(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer func() {
		mockDB.Mock.ExpectClose()
		mockDB.Close()
	}()

	t.Run("nested operations in same transaction", func(t *testing.T) {
		// 嵌套事务实际上是同一个事务，不创建新的事务连接
		mockDB.Mock.ExpectBegin()
		mockDB.Mock.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mockDB.Mock.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(2, 1))
		mockDB.Mock.ExpectCommit()

		tx1 := client.NewScoop().Begin()
		assert.NotNil(t, tx1)

		result := tx1.Model(TestUser{}).Create(&TestUser{
			Name:  "Outer Operation",
			Email: "outer@example.com",
		})
		assert.NoError(t, result.Error)

		// 在同一个事务中继续操作
		result = tx1.Model(TestUser{}).Create(&TestUser{
			Name:  "Inner Operation",
			Email: "inner@example.com",
		})
		assert.NoError(t, result.Error)

		tx1.Commit()
	})
}
