package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestScoop_Delete 测试 Delete 方法
func TestScoop_Delete(t *testing.T) {
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

	t.Run("delete with soft delete", func(t *testing.T) {
		mockDB.Mock.ExpectExec("UPDATE test_users SET deleted_at = .* WHERE .* AND deleted_at = 0").
			WillReturnResult(sqlmock.NewResult(0, 1))

		result := client.NewScoop().Model(TestUser{}).Where("id", 1).Delete()
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(1), result.RowsAffected)
	})

	t.Run("delete unscoped (hard delete)", func(t *testing.T) {
		mockDB.Mock.ExpectExec("DELETE FROM test_users WHERE .*").
			WillReturnResult(sqlmock.NewResult(0, 1))

		result := client.NewScoop().Model(TestUser{}).Unscoped().Where("id", 1).Delete()
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(1), result.RowsAffected)
	})

	t.Run("delete all records", func(t *testing.T) {
		mockDB.Mock.ExpectExec("UPDATE test_users SET deleted_at = .* WHERE deleted_at = 0").
			WillReturnResult(sqlmock.NewResult(0, 10))

		result := client.NewScoop().Model(TestUser{}).Delete()
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(10), result.RowsAffected)
	})

	t.Run("delete with error", func(t *testing.T) {
		mockDB.Mock.ExpectExec("UPDATE test_users SET deleted_at = .*").
			WillReturnError(assert.AnError)

		result := client.NewScoop().Model(TestUser{}).Where("id", 999).Delete()
		assert.Error(t, result.Error)
	})
}
