package db

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Expr(expression string, args ...interface{}) clause.Expr {
	return gorm.Expr(expression, args...)
}

func ExprInc(clientType, field string) clause.Expr {
	return ExprIncBy(clientType, field, 1)
}

func ExprIncBy(clientType, field string, cnt int64) clause.Expr {
	quotedField := quoteFieldWithTable(field, clientType)
	return gorm.Expr(fmt.Sprintf("%s + %d", quotedField, cnt))
}

func ExprIf(clientType, expr string, ok, nok interface{}) clause.Expr {
	switch clientType {
	case Sqlite:
		return gorm.Expr(fmt.Sprintf("IIF(%s, ?, ?)", expr), ok, nok)
	default:
		return gorm.Expr(fmt.Sprintf("IF(%s, ?, ?)", expr), ok, nok)
	}
}

// quoteFieldWithTable quotes field name and handles table prefix (e.g., "users.id" -> "`users`.`id`")
func quoteFieldWithTable(field string, clientType string) string {
	// Check if already fully quoted
	if (strings.HasPrefix(field, "`") && strings.Contains(field, "`.`")) ||
		(strings.HasPrefix(field, "\"") && strings.Contains(field, "\".\"")) {
		return field
	}

	// Check if field has table prefix
	if strings.Contains(field, ".") {
		parts := strings.SplitN(field, ".", 2)
		if len(parts) == 2 {
			tableName := parts[0]
			fieldName := parts[1]
			return quoteFieldName(tableName, clientType) + "." + quoteFieldName(fieldName, clientType)
		}
	}

	// No table prefix, just quote the field
	return quoteFieldName(field, clientType)
}
