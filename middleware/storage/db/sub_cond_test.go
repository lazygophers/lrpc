package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestSubCond_Like 测试 Like 条件
func TestSubCond_Like(t *testing.T) {
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

	t.Run("like condition", func(t *testing.T) {
		mockDB.Mock.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIKE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(1, "John Doe"))

		var user TestUser
		result := client.NewScoop().Model(TestUser{}).Where(db.Like("name", "%John%")).First(&user)
		assert.NoError(t, result.Error)
		assert.Equal(t, "John Doe", user.Name)
	})

	t.Run("left like condition", func(t *testing.T) {
		mockDB.Mock.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIKE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(2, "Alice Smith"))

		var user TestUser
		result := client.NewScoop().Model(TestUser{}).Where(db.LeftLike("name", "John")).First(&user)
		assert.NoError(t, result.Error)
		assert.Equal(t, "Alice Smith", user.Name)
	})

	t.Run("right like condition", func(t *testing.T) {
		mockDB.Mock.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIKE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(3, "Bob Johnson"))

		var user TestUser
		result := client.NewScoop().Model(TestUser{}).Where(db.RightLike("name", "Smith")).First(&user)
		assert.NoError(t, result.Error)
		assert.Equal(t, "Bob Johnson", user.Name)
	})

	t.Run("not like condition", func(t *testing.T) {
		mockDB.Mock.ExpectQuery("SELECT \\* FROM test_users WHERE .* NOT LIKE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(4, "Charlie Brown"))

		var user TestUser
		result := client.NewScoop().Model(TestUser{}).Where(db.NotLike("name", "%John%")).First(&user)
		assert.NoError(t, result.Error)
		assert.Equal(t, "Charlie Brown", user.Name)
	})

	t.Run("not left like condition", func(t *testing.T) {
		mockDB.Mock.ExpectQuery("SELECT \\* FROM test_users WHERE .* NOT LIKE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(5, "David Lee"))

		var user TestUser
		result := client.NewScoop().Model(TestUser{}).Where(db.NotLeftLike("name", "John")).First(&user)
		assert.NoError(t, result.Error)
		assert.Equal(t, "David Lee", user.Name)
	})

	t.Run("not right like condition", func(t *testing.T) {
		mockDB.Mock.ExpectQuery("SELECT \\* FROM test_users WHERE .* NOT LIKE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(6, "Emma Wilson"))

		var user TestUser
		result := client.NewScoop().Model(TestUser{}).Where(db.NotRightLike("name", "Smith")).First(&user)
		assert.NoError(t, result.Error)
		assert.Equal(t, "Emma Wilson", user.Name)
	})
}

// TestSubCond_Between 测试 Between 条件
func TestSubCond_Between(t *testing.T) {
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

	t.Run("between condition", func(t *testing.T) {
		mockDB.Mock.ExpectQuery("SELECT \\* FROM test_users WHERE .* BETWEEN .* AND .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}).
				AddRow(1, "User 1", 25))

		var user TestUser
		result := client.NewScoop().Model(TestUser{}).Where(db.Between("age", 20, 30)).First(&user)
		assert.NoError(t, result.Error)
		assert.Equal(t, 25, user.Age)
	})

	t.Run("not between condition", func(t *testing.T) {
		mockDB.Mock.ExpectQuery("SELECT \\* FROM test_users WHERE .* NOT BETWEEN .* AND .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}).
				AddRow(2, "User 2", 40))

		var user TestUser
		result := client.NewScoop().Model(TestUser{}).Where(db.NotBetween("age", 20, 30)).First(&user)
		assert.NoError(t, result.Error)
		assert.Equal(t, 40, user.Age)
	})
}

// TestSubCond_Or 测试 Or 条件
func TestSubCond_Or(t *testing.T) {
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

	t.Run("or condition", func(t *testing.T) {
		mockDB.Mock.ExpectQuery("SELECT \\* FROM test_users WHERE .* OR .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(1, "Test User"))

		var user TestUser
		result := client.NewScoop().Model(TestUser{}).Where(
			db.Or(
				db.Where("name", "John"),
				db.Where("email", "john@example.com"),
			),
		).First(&user)
		assert.NoError(t, result.Error)
	})
}
