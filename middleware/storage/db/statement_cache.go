package db

import (
	"gorm.io/gorm"
)

// getStatement gets a Statement from the pool and initializes it
func getStatement(db *gorm.DB, table string, model interface{}) *gorm.Statement {
	stmt := statementPool.Get().(*gorm.Statement)
	stmt.DB = db
	stmt.Table = table
	stmt.Model = model
	if db.Statement != nil {
		stmt.TableExpr = db.Statement.TableExpr
	}
	return stmt
}

// putStatement resets and returns a Statement to the pool
func putStatement(stmt *gorm.Statement) {
	// Reset statement fields to avoid memory leaks
	stmt.DB = nil
	stmt.Table = ""
	stmt.Model = nil
	stmt.Schema = nil
	stmt.TableExpr = nil
	statementPool.Put(stmt)
}
