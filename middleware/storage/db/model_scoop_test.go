package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestModelScoop_Asc 测试 Asc 方法
func TestModelScoop_Asc(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)
	scoop := model.NewScoop().Asc("name", "age")
	assert.NotNil(t, scoop)
	assert.Contains(t, scoop.orders, "name ASC")
	assert.Contains(t, scoop.orders, "age ASC")
}

// TestModelScoop_Desc 测试 Desc 方法
func TestModelScoop_Desc(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)
	scoop := model.NewScoop().Desc("created_at", "id")
	assert.NotNil(t, scoop)
	assert.Contains(t, scoop.orders, "created_at DESC")
	assert.Contains(t, scoop.orders, "id DESC")
}

// TestModelScoop_Or 测试 Or 方法
func TestModelScoop_Or(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)
	scoop := model.NewScoop().
		Where("name", "John").
		Or("name", "Jane")
	assert.NotNil(t, scoop)
}

// TestModelScoop_NotEqual 测试 NotEqual 方法
func TestModelScoop_NotEqual(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)
	scoop := model.NewScoop().NotEqual("status", 0)
	assert.NotNil(t, scoop)
}

// TestModelScoop_In 测试 In 方法
func TestModelScoop_In(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)

	t.Run("in with values", func(t *testing.T) {
		scoop := model.NewScoop().In("id", []int{1, 2, 3})
		assert.NotNil(t, scoop)
	})

	t.Run("in with empty slice", func(t *testing.T) {
		scoop := model.NewScoop().In("id", []int{})
		assert.NotNil(t, scoop)
	})
}

// TestModelScoop_NotIn 测试 NotIn 方法
func TestModelScoop_NotIn(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)

	t.Run("not in with values", func(t *testing.T) {
		scoop := model.NewScoop().NotIn("id", []int{1, 2, 3})
		assert.NotNil(t, scoop)
	})

	t.Run("not in with empty slice", func(t *testing.T) {
		scoop := model.NewScoop().NotIn("id", []int{})
		assert.NotNil(t, scoop)
	})
}

// TestModelScoop_Like 测试 Like 方法
func TestModelScoop_Like(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)

	t.Run("like with value", func(t *testing.T) {
		scoop := model.NewScoop().Like("name", "John")
		assert.NotNil(t, scoop)
	})

	t.Run("like with empty value", func(t *testing.T) {
		scoop := model.NewScoop().Like("name", "")
		assert.NotNil(t, scoop)
	})
}

// TestModelScoop_LeftLike 测试 LeftLike 方法
func TestModelScoop_LeftLike(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)

	t.Run("left like with value", func(t *testing.T) {
		scoop := model.NewScoop().LeftLike("name", "John")
		assert.NotNil(t, scoop)
	})

	t.Run("left like with empty value", func(t *testing.T) {
		scoop := model.NewScoop().LeftLike("name", "")
		assert.NotNil(t, scoop)
	})
}

// TestModelScoop_RightLike 测试 RightLike 方法
func TestModelScoop_RightLike(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)

	t.Run("right like with value", func(t *testing.T) {
		scoop := model.NewScoop().RightLike("name", "John")
		assert.NotNil(t, scoop)
	})

	t.Run("right like with empty value", func(t *testing.T) {
		scoop := model.NewScoop().RightLike("name", "")
		assert.NotNil(t, scoop)
	})
}

// TestModelScoop_NotLike 测试 NotLike 方法
func TestModelScoop_NotLike(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)

	t.Run("not like with value", func(t *testing.T) {
		scoop := model.NewScoop().NotLike("name", "John")
		assert.NotNil(t, scoop)
	})

	t.Run("not like with empty value", func(t *testing.T) {
		scoop := model.NewScoop().NotLike("name", "")
		assert.NotNil(t, scoop)
	})
}

// TestModelScoop_NotLeftLike 测试 NotLeftLike 方法
func TestModelScoop_NotLeftLike(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)

	t.Run("not left like with value", func(t *testing.T) {
		scoop := model.NewScoop().NotLeftLike("name", "John")
		assert.NotNil(t, scoop)
	})

	t.Run("not left like with empty value", func(t *testing.T) {
		scoop := model.NewScoop().NotLeftLike("name", "")
		assert.NotNil(t, scoop)
	})
}

// TestModelScoop_NotRightLike 测试 NotRightLike 方法
func TestModelScoop_NotRightLike(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)

	t.Run("not right like with value", func(t *testing.T) {
		scoop := model.NewScoop().NotRightLike("name", "John")
		assert.NotNil(t, scoop)
	})

	t.Run("not right like with empty value", func(t *testing.T) {
		scoop := model.NewScoop().NotRightLike("name", "")
		assert.NotNil(t, scoop)
	})
}

// TestModelScoop_Between 测试 Between 方法
func TestModelScoop_Between(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)
	scoop := model.NewScoop().Between("age", 18, 65)
	assert.NotNil(t, scoop)
}

// TestModelScoop_NotBetween 测试 NotBetween 方法
func TestModelScoop_NotBetween(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)
	scoop := model.NewScoop().NotBetween("age", 0, 10)
	assert.NotNil(t, scoop)
}

// TestModelScoop_Unscoped 测试 Unscoped 方法
func TestModelScoop_Unscoped(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)

	t.Run("unscoped without args", func(t *testing.T) {
		scoop := model.NewScoop().Unscoped()
		assert.NotNil(t, scoop)
		assert.True(t, scoop.unscoped)
	})

	t.Run("unscoped with true", func(t *testing.T) {
		scoop := model.NewScoop().Unscoped(true)
		assert.NotNil(t, scoop)
		assert.True(t, scoop.unscoped)
	})

	t.Run("unscoped with false", func(t *testing.T) {
		scoop := model.NewScoop().Unscoped(false)
		assert.NotNil(t, scoop)
		assert.False(t, scoop.unscoped)
	})
}

// TestModelScoop_Limit 测试 Limit 方法
func TestModelScoop_Limit(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)
	scoop := model.NewScoop().Limit(10)
	assert.NotNil(t, scoop)
	assert.Equal(t, uint64(10), scoop.limit)
}

// TestModelScoop_Offset 测试 Offset 方法
func TestModelScoop_Offset(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)
	scoop := model.NewScoop().Offset(5)
	assert.NotNil(t, scoop)
	assert.Equal(t, uint64(5), scoop.offset)
}

// TestModelScoop_Group 测试 Group 方法
func TestModelScoop_Group(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)
	scoop := model.NewScoop().Group("age", "name")
	assert.NotNil(t, scoop)
	assert.Contains(t, scoop.groups, "age")
	assert.Contains(t, scoop.groups, "name")
}

// TestModelScoop_Order 测试 Order 方法
func TestModelScoop_Order(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)
	scoop := model.NewScoop().Order("id DESC", "name ASC")
	assert.NotNil(t, scoop)
	assert.Contains(t, scoop.orders, "id DESC")
	assert.Contains(t, scoop.orders, "name ASC")
}

// TestModelScoop_Ignore 测试 Ignore 方法
func TestModelScoop_Ignore(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)

	t.Run("ignore without args", func(t *testing.T) {
		scoop := model.NewScoop().Ignore()
		assert.NotNil(t, scoop)
		assert.True(t, scoop.ignore)
	})

	t.Run("ignore with true", func(t *testing.T) {
		scoop := model.NewScoop().Ignore(true)
		assert.NotNil(t, scoop)
		assert.True(t, scoop.ignore)
	})

	t.Run("ignore with false", func(t *testing.T) {
		scoop := model.NewScoop().Ignore(false)
		assert.NotNil(t, scoop)
		assert.False(t, scoop.ignore)
	})
}

// TestModelScoop_ChainedOperations 测试链式调用
func TestModelScoop_ChainedOperations(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)

	scoop := model.NewScoop().
		Select("id", "name", "age").
		Where("age", ">", 18).
		NotEqual("status", 0).
		In("role", []string{"admin", "user"}).
		Like("name", "John").
		Between("age", 18, 65).
		Order("id DESC").
		Limit(10).
		Offset(5)

	assert.NotNil(t, scoop)
	assert.Equal(t, uint64(10), scoop.limit)
	assert.Equal(t, uint64(5), scoop.offset)
	assert.Contains(t, scoop.selects, "id")
	assert.Contains(t, scoop.selects, "name")
	assert.Contains(t, scoop.selects, "age")
}
