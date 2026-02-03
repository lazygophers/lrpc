package db

import (
	"fmt"
	"strconv"
	"time"

	"github.com/lazygophers/log"
)

type DeleteResult struct {
	RowsAffected int64
	Error        error
}

// Delete performs a soft or hard delete on matching rows.
// If the table has a DeletedAt field, performs soft delete by setting timestamp.
// Otherwise performs hard delete by removing rows from the database.
// Use Unscoped() to force hard delete even with DeletedAt field.
// Returns DeleteResult with RowsAffected count.
//
// Example:
//
//	// Soft delete
//	result := scoop.Table("users").Where("id = ?", 1).Delete()
//	// Hard delete
//	result := scoop.Table("users").Unscoped().Where("id = ?", 1).Delete()
func (p *Scoop) Delete() *DeleteResult {
	if p.cond.skip {
		return &DeleteResult{}
	}

	if p.table == "" {
		return &DeleteResult{
			Error: fmt.Errorf("Delete failed: table name is empty, use Table() to specify table name"),
		}
	}

	if !p.unscoped && p.hasDeletedAt {
		p.cond.whereRaw("deleted_at = 0")
	}

	p.inc()
	defer p.dec()

	sqlRaw := log.GetBuffer()
	defer log.PutBuffer(sqlRaw)

	// 软删除
	if !p.unscoped && p.hasDeletedAt {
		sqlRaw.WriteString("UPDATE")
		sqlRaw.WriteString(" ")
		sqlRaw.WriteString(p.table)
		sqlRaw.WriteString(" SET deleted_at = ")
		sqlRaw.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
		p.cond.addCond("deleted_at", "=", 0)
	} else {
		sqlRaw.WriteString("DELETE FROM")
		sqlRaw.WriteString(" ")
		sqlRaw.WriteString(p.table)
	}

	if len(p.cond.conds) > 0 {
		sqlRaw.WriteString(" WHERE ")
		sqlRaw.WriteString(p.cond.conds[0])
		for _, c := range p.cond.conds[1:] {
			sqlRaw.WriteString(" AND ")
			sqlRaw.WriteString(c)
		}
	}

	start := time.Now()
	res := p._db.Exec(sqlRaw.String())
	GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
		return sqlRaw.String(), res.RowsAffected
	}, res.Error)

	if res.Error != nil {
		log.Errorf("err:%v", res.Error)
	}

	return &DeleteResult{
		RowsAffected: res.RowsAffected,
		Error:        res.Error,
	}
}
