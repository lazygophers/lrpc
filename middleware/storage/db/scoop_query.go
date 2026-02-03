package db

import (
	"fmt"
	"reflect"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
)

type FindResult struct {
	RowsAffected int64
	Error        error
}

// Find executes a SELECT query and scans all matching rows into out.
// The out parameter must be a pointer to a slice of structs.
// Returns FindResult containing any error that occurred.
//
// Example:
//
//	var users []User
//	result := scoop.Where("age > ?", 18).Find(&users)
func (p *Scoop) Find(out interface{}) *FindResult {
	if p.cond.skip {
		return &FindResult{}
	}

	vv := reflect.ValueOf(out)
	if vv.Type().Kind() != reflect.Ptr {
		return &FindResult{
			Error: fmt.Errorf("Find failed: out parameter must be a pointer, got %v", vv.Type().Kind()),
		}
	}
	vv = vv.Elem()
	if vv.Type().Kind() != reflect.Slice {
		return &FindResult{
			Error: fmt.Errorf("Find failed: out parameter must be a pointer to slice, got pointer to %v", vv.Type().Kind()),
		}
	}

	elem := vv.Type().Elem()

	if p.table == "" {
		p.table = getTableName(elem.Elem())
	}

	if !p.unscoped && (p.hasDeletedAt || hasDeletedAt(elem)) {
		p.cond.whereRaw(condNotDeleted)
	}

	p.inc()
	defer p.dec()

	logBuf := log.GetBuffer()
	defer log.PutBuffer(logBuf)

	sqlRaw := p.findSql()
	start := time.Now()

	// Use GORM's Rows and ScanRows to properly handle serializers
	rows, err := p._db.Raw(sqlRaw).Rows()
	if err != nil {
		GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
			return sqlRaw, 0
		}, err)
		return &FindResult{
			Error: err,
		}
	}
	defer rows.Close()

	// Clear the slice first
	vv.Set(reflect.MakeSlice(vv.Type(), 0, 0))

	var rawsAffected int64
	for rows.Next() {
		rawsAffected++
		// Create a new element for each row
		elemPtr := reflect.New(elem.Elem())
		err = p._db.ScanRows(rows, elemPtr.Interface())
		if err != nil {
			GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
				return sqlRaw, rawsAffected
			}, err)
			return &FindResult{
				Error: err,
			}
		}
		// Append the scanned element to the slice
		vv.Set(reflect.Append(vv, elemPtr))
	}

	GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
		return sqlRaw, rawsAffected
	}, nil)
	return &FindResult{
		RowsAffected: rawsAffected,
	}
}

type ChunkResult struct {
	Error error
}

func (p *Scoop) Chunk(dest interface{}, size uint64, fc func(tx *Scoop, offset uint64) error) *ChunkResult {
	// Optimize: Perform reflection checks only once instead of every iteration
	// This significantly improves performance in loops with many chunks
	p.offset = 0
	p.limit = size

	vv := reflect.ValueOf(dest)
	if vv.Type().Kind() != reflect.Ptr {
		return &ChunkResult{
			Error: fmt.Errorf("Chunk failed: dest parameter must be a pointer, got %v", vv.Type().Kind()),
		}
	}

	vv = vv.Elem()
	if vv.Type().Kind() != reflect.Slice {
		return &ChunkResult{
			Error: fmt.Errorf("Chunk failed: dest parameter must be a pointer to slice, got pointer to %v", vv.Type().Kind()),
		}
	}

	elem := vv.Type().Elem().Elem()

	if p.table == "" {
		p.table = getTableName(elem)
	}

	// Optimize: Check DeletedAt once and add condition before loop
	p.hasDeletedAt = hasDeletedAt(elem)
	if !p.unscoped && p.hasDeletedAt {
		p.cond.whereRaw(condNotDeleted)
	}

	p.inc()
	defer p.dec()

	// Optimize: Inline query logic to avoid repeated reflection checks in Find()
	// This eliminates function call overhead and redundant type validations
	for {
		// Reset dest content
		vv.Set(reflect.MakeSlice(vv.Type(), 0, int(size)))

		// Inline Find logic to avoid repeated reflection and validation
		if p.cond.skip {
			break
		}

		// Build SQL once per iteration (necessary as offset changes)
		sqlRaw := p.findSql()
		start := time.Now()

		// Execute query using GORM's Scan to handle serializers
		scope := p._db.Raw(sqlRaw)
		err := scope.Scan(dest).Error

		// Log the query
		GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
			return sqlRaw, scope.RowsAffected
		}, err)

		if err != nil {
			return &ChunkResult{
				Error: err,
			}
		}

		// Break if no more rows
		if scope.RowsAffected == 0 {
			break
		}

		// Execute user callback
		err = fc(p, p.offset)
		if err != nil {
			return &ChunkResult{
				Error: err,
			}
		}

		// Move to next chunk
		p.offset += size
	}

	return &ChunkResult{}
}

type FirstResult struct {
	Error error
}

// First executes a SELECT query with LIMIT 1 and scans the first row into out.
// The out parameter must be a pointer to a struct.
// Returns FirstResult with Error set to ErrRecordNotFound if no row is found.
//
// Example:
//
//	var user User
//	result := scoop.Table("users").Where("id = ?", 1).First(&user)
func (p *Scoop) First(out interface{}) *FirstResult {
	if p.cond.skip {
		return &FirstResult{
			Error: p.getNotFoundError(),
		}
	}

	vv := reflect.ValueOf(out)
	if vv.Type().Kind() != reflect.Ptr {
		return &FirstResult{
			Error: fmt.Errorf("First failed: out parameter must be a pointer, got %v", vv.Type().Kind()),
		}
	}

	if p.table == "" {
		p.table = getTableName(vv.Type())
	}

	if !p.unscoped && (p.hasDeletedAt || hasDeletedAt(vv.Type())) {
		p.cond.whereRaw("deleted_at = 0")
	}

	p.offset = 0
	p.limit = 1

	p.inc()
	defer p.dec()

	sqlRaw := p.findSql()
	start := time.Now()

	// Use GORM's Rows and ScanRows to properly handle serializers
	rows, err := p._db.Raw(sqlRaw).Rows()
	if err != nil {
		GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
			return sqlRaw, 0
		}, err)
		return &FirstResult{
			Error: err,
		}
	}
	defer rows.Close()

	var rowAffected int64
	if rows.Next() {
		rowAffected = 1
		err = p._db.ScanRows(rows, out)
		if err != nil {
			GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
				return sqlRaw, rowAffected
			}, err)
			return &FirstResult{
				Error: err,
			}
		}
	}

	if rowAffected == 0 {
		GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
			return sqlRaw, 0
		}, p.getNotFoundError())
		return &FirstResult{
			Error: p.getNotFoundError(),
		}
	}

	GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
		return sqlRaw, rowAffected
	}, nil)
	return &FirstResult{}
}

// Count returns the number of rows matching the query conditions.
// Respects WHERE clauses and soft delete (DeletedAt) filtering.
// Returns the count and any error that occurred.
//
// Example:
//
//	count, err := scoop.Table("users").Where("age > ?", 18).Count()
func (p *Scoop) Count() (uint64, error) {
	if p.cond.skip {
		return 0, nil
	}

	if p.table == "" {
		return 0, fmt.Errorf("Count failed: table name is empty, use Table() to specify table name")
	}

	if !p.unscoped && p.hasDeletedAt {
		p.cond.whereRaw("deleted_at = 0")
	}

	p.inc()
	defer p.dec()

	sqlRaw := log.GetBuffer()
	defer log.PutBuffer(sqlRaw)

	sqlRaw.WriteString("SELECT COUNT(*) FROM ")
	sqlRaw.WriteString(p.table)

	if len(p.cond.conds) > 0 {
		sqlRaw.WriteString(" WHERE ")
		sqlRaw.WriteString(p.cond.conds[0])
		for _, c := range p.cond.conds[1:] {
			sqlRaw.WriteString(" AND ")
			sqlRaw.WriteString(c)
		}
	}

	start := time.Now()
	var count uint64
	err := p._db.Raw(sqlRaw.String()).Scan(&count).Error
	GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
		return sqlRaw.String(), int64(count)
	}, err)

	return count, err
}

func (p *Scoop) Exist() (bool, error) {
	if p.cond.skip {
		return false, nil
	}

	if p.table == "" {
		return false, fmt.Errorf("Exists failed: table name is empty, use Table() to specify table name")
	}

	if !p.unscoped && p.hasDeletedAt {
		p.cond.whereRaw("deleted_at = 0")
	}

	p.limit = 1
	p.offset = 0

	p.inc()
	defer p.dec()

	sqlRaw := log.GetBuffer()
	defer log.PutBuffer(sqlRaw)

	sqlRaw.WriteString("SELECT ")
	if p.hasId {
		sqlRaw.WriteString("id")
	} else {
		sqlRaw.WriteString("count(*)")
	}

	sqlRaw.WriteString(" FROM ")
	sqlRaw.WriteString(p.table)

	if len(p.cond.conds) > 0 {
		sqlRaw.WriteString(" WHERE ")
		sqlRaw.WriteString(p.cond.conds[0])
		for _, c := range p.cond.conds[1:] {
			sqlRaw.WriteString(" AND ")
			sqlRaw.WriteString(c)
		}
	}

	sqlRaw.WriteString(" LIMIT 1 OFFSET 0")

	start := time.Now()
	var count int64
	err := p._db.Raw(sqlRaw.String()).Scan(&count).Error
	GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
		return sqlRaw.String(), 0
	}, err)
	return count > 0, err
}

func (p *Scoop) FindByPage(opt *core.ListOption, values any) (*core.Paginate, error) {
	if opt == nil {
		opt = &core.ListOption{
			Offset: core.DefaultOffset,
			Limit:  core.DefaultLimit,
		}
	}

	p.Offset(opt.Offset).Limit(opt.Limit)

	page := &core.Paginate{
		Offset: opt.Offset,
		Limit:  opt.Limit,
	}

	p.Offset(opt.Offset).Limit(opt.Limit)

	p.inc()
	defer p.dec()

	err := p.Find(values).Error
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	if opt.ShowTotal {
		page.Total, err = p.Count()
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}
	}

	return page, nil
}

type ScanResult struct {
	RowsAffected int64
	Error        error
}

// Scan executes the query and scans the result into dest.
// The dest parameter can be:
//   - A pointer to a basic type (int, string, etc.) for single value queries
//   - A pointer to a slice for multiple values
//   - A pointer to a struct for single row queries
//   - A pointer to a slice of structs for multiple rows
//
// Returns ScanResult containing any error that occurred.
//
// Example:
//
//	var count int64
//	result := scoop.Table("users").Select("COUNT(*)").Where("age > ?", 18).Scan(&count)
//
//	var names []string
//	result := scoop.Table("users").Select("name").Where("age > ?", 18).Scan(&names)
func (p *Scoop) Scan(dest interface{}) *ScanResult {
	if p.cond.skip {
		return &ScanResult{}
	}

	if p.table == "" {
		return &ScanResult{
			Error: fmt.Errorf("Scan failed: table name is empty, use Table() to specify table name"),
		}
	}

	if !p.unscoped && p.hasDeletedAt {
		p.cond.whereRaw("deleted_at = 0")
	}

	p.inc()
	defer p.dec()

	sqlRaw := p.findSql()
	start := time.Now()

	res := p._db.Raw(sqlRaw).Scan(dest)

	GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
		return sqlRaw, res.RowsAffected
	}, res.Error)

	if res.Error != nil {
		log.Errorf("err:%v", res.Error)
		return &ScanResult{
			RowsAffected: res.RowsAffected,
			Error:        res.Error,
		}
	}

	return &ScanResult{
		RowsAffected: res.RowsAffected,
		Error:        nil,
	}
}
