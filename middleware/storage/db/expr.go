package db

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DefaultDriver = "mysql"

func Expr(expression string, args ...interface{}) clause.Expr {
	return gorm.Expr(expression, args...)
}

func ExprInc(field string) clause.Expr {
	return ExprIncBy(field, 1)
}

func ExprIncBy(field string, cnt int64) clause.Expr {
	return gorm.Expr(fmt.Sprintf("%s + %d", field, cnt))
}

func ExprIf(expr string, ok, nok interface{}) clause.Expr {
	switch DefaultDriver {
	case "sqlite", "sqlite3":
		return gorm.Expr(fmt.Sprintf("IIF(%s, ?, ?)", expr), ok, nok)
	default:
		return gorm.Expr(fmt.Sprintf("IF(%s, ?, ?)", expr), ok, nok)
	}
}
