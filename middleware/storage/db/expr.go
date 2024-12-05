package db

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DefaultDriver = "mysql"

func Expr(expression string, args ...interface{}) map[string]any {
	return map[string]any{
		"expr": expression,
		"args": args,
	}
}

func If(expr string, ok, nok interface{}) clause.Expr {
	switch DefaultDriver {
	case "sqlite", "sqlite3":
		return gorm.Expr(fmt.Sprintf("IIF(%s, ?, ?)", expr), ok, nok)
	default:
		return gorm.Expr(fmt.Sprintf("IF(%s, ?, ?)", expr), ok, nok)
	}
}
