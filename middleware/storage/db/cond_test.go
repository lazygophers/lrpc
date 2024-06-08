package db_test

import (
	"github.com/lazygophers/utils/db"
	"strconv"
	"strings"
	"testing"
)

func TestCond(t *testing.T) {
	// t.Log(cond.Clean().Where("a", 1).ToString())

	t.Log(db.OrWhere(map[string]any{
		"a": 1,
	}, map[string]any{
		"a": 2,
	}, map[string]any{
		"a": 3,
	}).ToString())

}

// (a = 1 and b =2 ) or (a = 2 and b = 1)
func TestSubCond(t *testing.T) {
	t.Log(db.OrWhere(db.Where(map[string]any{
		"a": 1,
		"b": 2,
	}), db.Where(map[string]any{
		"a": 2,
		"b": 3,
	})).ToString())
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
