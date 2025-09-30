package db

import (
	"reflect"
	"testing"

	"gotest.tools/v3/assert"
)

// Test wrapper functions that are one-line calls to hasField

func TestHasDeletedAt(t *testing.T) {
	type TestStruct struct {
		DeletedAt int64
	}

	t.Run("has DeletedAt", func(t *testing.T) {
		typ := reflect.TypeOf(TestStruct{})
		result := hasDeletedAt(typ)
		assert.Assert(t, result)
	})

	t.Run("no DeletedAt", func(t *testing.T) {
		type NoDeletedAt struct {
			ID int64
		}
		typ := reflect.TypeOf(NoDeletedAt{})
		result := hasDeletedAt(typ)
		assert.Assert(t, !result)
	})
}

func TestHasCreatedAt(t *testing.T) {
	type TestStruct struct {
		CreatedAt int64
	}

	t.Run("has CreatedAt", func(t *testing.T) {
		typ := reflect.TypeOf(TestStruct{})
		result := hasCreatedAt(typ)
		assert.Assert(t, result)
	})

	t.Run("no CreatedAt", func(t *testing.T) {
		type NoCreatedAt struct {
			ID int64
		}
		typ := reflect.TypeOf(NoCreatedAt{})
		result := hasCreatedAt(typ)
		assert.Assert(t, !result)
	})
}

func TestHasUpdatedAt(t *testing.T) {
	type TestStruct struct {
		UpdatedAt int64
	}

	t.Run("has UpdatedAt", func(t *testing.T) {
		typ := reflect.TypeOf(TestStruct{})
		result := hasUpdatedAt(typ)
		assert.Assert(t, result)
	})

	t.Run("no UpdatedAt", func(t *testing.T) {
		type NoUpdatedAt struct {
			ID int64
		}
		typ := reflect.TypeOf(NoUpdatedAt{})
		result := hasUpdatedAt(typ)
		assert.Assert(t, !result)
	})
}

func TestHasId(t *testing.T) {
	type TestStruct struct {
		Id int64
	}

	t.Run("has Id", func(t *testing.T) {
		typ := reflect.TypeOf(TestStruct{})
		result := hasId(typ)
		assert.Assert(t, result)
	})

	t.Run("no Id", func(t *testing.T) {
		type NoId struct {
			Name string
		}
		typ := reflect.TypeOf(NoId{})
		result := hasId(typ)
		assert.Assert(t, !result)
	})
}