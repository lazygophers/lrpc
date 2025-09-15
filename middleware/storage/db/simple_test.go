package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
)

// Simple working tests to improve coverage

func TestSimpleExpr(t *testing.T) {
	expr := db.Expr("COUNT(*)")
	assert.Equal(t, "COUNT(*)", expr.SQL)
}

func TestSimpleExprInc(t *testing.T) {
	expr := db.ExprInc("counter")
	assert.Equal(t, "counter + 1", expr.SQL)
}

func TestSimpleExprIncBy(t *testing.T) {
	expr := db.ExprIncBy("counter", 5)
	assert.Equal(t, "counter + 5", expr.SQL)
}

func TestSimpleExprIf(t *testing.T) {
	expr := db.ExprIf("status = 1", "active", "inactive")
	assert.Assert(t, expr.SQL != "")
}

func TestSimpleUtilFunctions(t *testing.T) {
	// Test EscapeMysqlString
	result := db.EscapeMysqlString("test'quote")
	assert.Equal(t, "test\\'quote", result)
	
	// Test UniqueSlice
	input := []int{1, 2, 3, 2, 1}
	unique := db.UniqueSlice(input)
	uniqueSlice := unique.([]int)
	assert.Assert(t, len(uniqueSlice) < len(input))
	
	// Test Camel2UnderScore
	result2 := db.Camel2UnderScore("CamelCase")
	assert.Equal(t, "camel_case", result2)
	
	// Test FormatSql
	sql := db.FormatSql("SELECT * FROM users WHERE id = ?", 123)
	assert.Assert(t, len(sql) > 0)
	
	// Test IsUniqueIndexConflictErr
	err := &testError{"Error 1062: Duplicate entry"}
	isConflict := db.IsUniqueIndexConflictErr(err)
	assert.Assert(t, isConflict)
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestSimpleCond(t *testing.T) {
	// Test String and GoString methods
	cond := db.Where("id", 1)
	str := cond.String()
	goStr := cond.GoString()
	assert.Equal(t, str, goStr)
	
	// Test Reset
	cond.Reset()
	assert.Equal(t, "", cond.ToString())
}