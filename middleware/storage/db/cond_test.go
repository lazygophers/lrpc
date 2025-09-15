package db_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
)

func TestCond(t *testing.T) {
	assert.Equal(t, db.OrWhere(map[string]any{
		"a": 1,
	}, map[string]any{
		"a": 2,
	}, map[string]any{
		"a": 3,
	}).ToString(), "((`a` = 1) OR (`a` = 2) OR (`a` = 3))")

	assert.Equal(t, db.Where("a", 1).ToString(), "(`a` = 1)")

	assert.Equal(t, db.Or(db.Where("a", 1), db.Where("a", 2)).ToString(), "((`a` = 1) OR (`a` = 2))")

	// Test OrWhere with multiple conditions - order may vary due to map iteration
	result := db.OrWhere(db.Where(map[string]any{
		"a": 1,
		"b": 2,
	}), db.Where(map[string]any{
		"a": 2,
		"b": 3,
	})).ToString()
	
	// Check that result contains both expected condition groups in any order
	if !(strings.Contains(result, "(`a` = 1)") && strings.Contains(result, "(`b` = 2)")) {
		t.Errorf("Result should contain a=1 and b=2 conditions: %s", result)
	}
	if !(strings.Contains(result, "(`a` = 2)") && strings.Contains(result, "(`b` = 3)")) {
		t.Errorf("Result should contain a=2 and b=3 conditions: %s", result)
	}
	if !strings.Contains(result, " OR ") {
		t.Errorf("Result should contain OR operator: %s", result)
	}
}

func TestLike(t *testing.T) {
	t.Log(db.Where("name", "like", "%a%").ToString())
}

func TestIn(t *testing.T) {
	t.Log(db.Where("id", "in", []int{1, 2, 3}).ToString())
}

func TestQuote(t *testing.T) {
	t.Log(strconv.Quote("a"))
}

func TestGormTag(t *testing.T) {
	//tag := "column:id;primaryKey;autoIncrement;not null"
	//tag := "primaryKey;autoIncrement;not null"
	tag := "primaryKey"

	idx := strings.Index(tag, "primaryKey")
	t.Log(idx)
}
