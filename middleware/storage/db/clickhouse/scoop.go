package clickhouse

import (
	"database/sql"
	"fmt"
	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/lazygophers/utils/stringx"
	"gorm.io/gorm"
	"reflect"
	"strconv"
	"time"
)

type Scoop struct {
	_db *sql.DB

	notFoundError error

	hasDeletedAt bool
	hasId        bool
	table        string

	cond          Cond
	limit, offset uint64
	selects       []string
	groups        []string
	orders        []string
	unscoped      bool

	ignore bool

	depth int
}

func NewScoop(db *sql.DB) *Scoop {
	return &Scoop{
		depth: 1,
		_db:   db,
	}
}

func (p *Scoop) getNotFoundError() error {
	if p.notFoundError != nil {
		return p.notFoundError
	}

	return gorm.ErrRecordNotFound
}

func (p *Scoop) IsNotFound(err error) bool {
	return err == p.getNotFoundError() || err == gorm.ErrRecordNotFound
}

func (p *Scoop) inc() {
	p.depth++
}

func (p *Scoop) dec() {
	p.depth--
}

func (p *Scoop) Table(table string) *Scoop {
	p.table = table
	return p
}

func (p *Scoop) From(table string) *Scoop {
	p.table = table
	return p
}

//func (p *Scoop) Session(config ...*gorm.Session) *Scoop {
//	if len(config) == 0 {
//		return NewScoop(p._db.Session(&gorm.Session{
//			NewDB: true,
//		}))
//	}
//	return NewScoop(p._db.Session(config[0]))
//}

func (p *Scoop) Model(m any) *Scoop {
	rt := reflect.ValueOf(m).Type()
	p.table = getTableName(rt)
	p.hasDeletedAt = hasDeleted(rt)
	p.hasId = hasId(rt)

	return p
}

// ——————————条件——————————

func (p *Scoop) Select(fields ...string) *Scoop {
	p.selects = append(p.selects, fields...)
	return p
}

func (p *Scoop) Where(args ...interface{}) *Scoop {
	p.cond.Where(args...)
	return p
}

func (p *Scoop) Equal(column string, value interface{}) *Scoop {
	p.cond.where(column, value)
	return p
}

func (p *Scoop) NotEqual(column string, value interface{}) *Scoop {
	p.cond.where(column, "!= ", value)
	return p
}

func (p *Scoop) In(column string, values interface{}) *Scoop {
	vo := EnsureIsSliceOrArray(values)
	if vo.Len() == 0 {
		p.cond.where(false)
		return p
	}
	p.cond.where(column, "IN", UniqueSlice(vo.Interface()))
	return p
}

func (p *Scoop) NotIn(column string, values interface{}) *Scoop {
	vo := EnsureIsSliceOrArray(values)
	if vo.Len() == 0 {
		return p
	}
	p.cond.where(column, "NOT IN", UniqueSlice(vo.Interface()))
	return p
}

func (p *Scoop) Like(column string, value string) *Scoop {
	p.cond.where(column, "LIKE", "%"+value+"%")
	return p
}

func (p *Scoop) LeftLike(column string, value string) *Scoop {
	p.cond.where(column, "LIKE", "%"+value)
	return p
}

func (p *Scoop) RightLike(column string, value string) *Scoop {
	p.cond.where(column, "LIKE", value+"%")
	return p
}

func (p *Scoop) NotLike(column string, value string) *Scoop {
	p.cond.where(column, "NOT LIKE", "%"+value+"%")
	return p
}

func (p *Scoop) NotLeftLike(column string, value string) *Scoop {
	p.cond.where(column, "NOT LIKE", "%"+value)
	return p
}

func (p *Scoop) NotRightLike(column string, value string) *Scoop {
	p.cond.where(column, "NOT LIKE", value+"%")
	return p
}

func (p *Scoop) Between(column string, min, max interface{}) *Scoop {
	p.cond.whereRaw(quoteFieldName(column)+" BETWEEN ? AND ?", min, max)
	return p
}

func (p *Scoop) NotBetween(column string, min, max interface{}) *Scoop {
	p.cond.whereRaw(quoteFieldName(column)+" NOT BETWEEN ? AND ?", min, max)
	return p
}

func (p *Scoop) Unscoped(b ...bool) *Scoop {
	if len(b) == 0 {
		p.unscoped = true
		return p
	}
	p.unscoped = b[0]
	return p
}

func (p *Scoop) Limit(limit uint64) *Scoop {
	p.limit = limit
	return p
}

func (p *Scoop) Offset(offset uint64) *Scoop {
	p.offset = offset
	return p
}

func (p *Scoop) Group(fields ...string) *Scoop {
	p.groups = append(p.groups, fields...)
	return p
}

func (p *Scoop) Order(fields ...string) *Scoop {
	p.orders = append(p.orders, fields...)
	return p
}

func (p *Scoop) Ignore(b ...bool) *Scoop {
	if len(b) == 0 {
		p.ignore = true
		return p
	}
	p.ignore = b[0]
	return p
}

// ——————————操作——————————

func (p *Scoop) findSql() string {
	b := log.GetBuffer()
	defer log.PutBuffer(b)

	b.WriteString("SELECT ")
	if len(p.selects) > 0 {
		b.WriteString(p.selects[0])

		for _, s := range p.selects[1:] {
			b.WriteString(", ")
			b.WriteString(s)
		}
	} else {
		b.WriteString("*")
	}

	b.WriteString(" FROM ")
	b.WriteString(p.table)

	if len(p.cond.conds) > 0 {
		b.WriteString(" WHERE ")
		b.WriteString(p.cond.conds[0])
		for _, c := range p.cond.conds[1:] {
			b.WriteString(" AND ")
			b.WriteString(c)
		}
	}

	if len(p.groups) > 0 {
		b.WriteString(" GROUP BY ")
		b.WriteString(p.groups[0])
		for _, g := range p.groups[1:] {
			b.WriteString(", ")
			b.WriteString(g)
		}
	}

	if len(p.orders) > 0 {
		b.WriteString(" ORDER BY ")
		b.WriteString(p.orders[0])
		for _, o := range p.orders[1:] {
			b.WriteString(", ")
			b.WriteString(o)
		}
	}

	if p.limit > 0 {
		b.WriteString(" LIMIT ")
		b.WriteString(strconv.FormatUint(p.limit, 10))
	}

	if p.offset > 0 {
		b.WriteString(" OFFSET ")
		b.WriteString(strconv.FormatUint(p.offset, 10))
	}

	return b.String()
}

type FindResult struct {
	RowsAffected int64
	Error        error
}

func (p *Scoop) Find(out interface{}) *FindResult {
	if p.cond.skip {
		return &FindResult{}
	}

	vv := reflect.ValueOf(out)
	if vv.Type().Kind() != reflect.Ptr {
		panic("invalid out type, not ptr")
	}
	vv = vv.Elem()
	if vv.Type().Kind() != reflect.Slice {
		panic("invalid out type, not slice")
	}

	elem := vv.Type().Elem()

	if p.table == "" {
		p.table = getTableName(elem.Elem())
	}

	if !p.unscoped && (p.hasDeletedAt || hasDeleted(elem)) {
		p.cond.whereRaw("deleted_at = 0")
	}

	p.inc()
	defer p.dec()

	logBuf := log.GetBuffer()
	defer log.PutBuffer(logBuf)

	sqlRaw := p.findSql()
	start := Now()

	rows, err := p._db.Query(sqlRaw)
	if err != nil {
		return &FindResult{
			Error: err,
		}
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		db.GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
			return sqlRaw, -1
		}, err)
		return &FindResult{
			Error: err,
		}
	}

	values := make([]sql.RawBytes, len(cols))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var rawsAffected int64
	// 把数据写回到out
	for rows.Next() {
		rawsAffected++

		err = rows.Scan(scanArgs...)
		if err != nil {
			db.GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
				return sqlRaw, rawsAffected
			}, err)
			return &FindResult{
				Error: err,
			}
		}
		var v reflect.Value
		if elem.Elem().Kind() == reflect.Ptr {
			v = reflect.New(elem.Elem().Elem())
		} else {
			v = reflect.New(elem.Elem())
		}

		for i, col := range values {
			if col == nil {
				continue
			}
			field := v.Elem().FieldByName(stringx.Snake2Camel(cols[i]))
			if !field.IsValid() {
				log.Warnf("invalid field: %s", stringx.Snake2Camel(cols[i]))
				continue
			}

			err = decode(field, col)
			if err != nil {
				db.GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
					return sqlRaw, rawsAffected
				}, err)
				return &FindResult{
					Error: err,
				}
			}
		}

		vv.Set(reflect.Append(vv, v))
	}

	db.GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
		return sqlRaw, rawsAffected
	}, nil)
	return &FindResult{
		RowsAffected: rawsAffected,
	}
}

//type ChunkResult struct {
//	Error error
//}
//
//func (p *Scoop) Chunk(dest interface{}, size uint64, fc func(tx *Scoop, offset uint64) error) *ChunkResult {
//	p.offset = 0
//	p.limit = size
//
//	vv := reflect.ValueOf(dest)
//	if vv.Type().Kind() != reflect.Ptr {
//		panic("invalid out type, not ptr")
//	}
//
//	vv = vv.Elem()
//	if vv.Type().Kind() != reflect.Slice {
//		panic("invalid out type, not slice")
//	}
//
//	elem := vv.Type().Elem().Elem()
//
//	if p.table == "" {
//		p.table = getTableName(elem)
//	}
//
//	p.hasDeletedAt = hasDeleted(elem)
//
//	p.inc()
//	defer p.dec()
//
//	for {
//		// 重置dest内容
//		vv.Set(reflect.MakeSlice(vv.Type(), 0, int(size)))
//
//		res := p.Find(dest)
//		if res.Error != nil {
//			return &ChunkResult{
//				Error: res.Error,
//			}
//		}
//
//		if res.RowsAffected == 0 {
//			break
//		}
//
//		err := fc(p, p.offset)
//		if err != nil {
//			return &ChunkResult{
//				Error: err,
//			}
//		}
//
//		p.offset += size
//	}
//
//	return &ChunkResult{}
//}
//
//type FirstResult struct {
//	Error error
//}

//func (p *Scoop) First(out interface{}) *FirstResult {
//	if p.cond.skip {
//		return &FirstResult{
//			Error: p.getNotFoundError(),
//		}
//	}
//
//	vv := reflect.ValueOf(out)
//	if vv.Type().Kind() != reflect.Ptr {
//		panic("invalid out type, not ptr")
//	}
//
//	if p.table == "" {
//		p.table = getTableName(vv.Type())
//	}
//
//	if !p.unscoped && (p.hasDeletedAt || hasDeleted(vv.Type())) {
//		p.cond.whereRaw("deleted_at = 0")
//	}
//
//	p.offset = 0
//	p.limit = 1
//
//	p.inc()
//	defer p.dec()
//
//	sqlRaw := p.findSql()
//	start := Now()
//
//	scope := p._db.Raw(sqlRaw)
//	rows, err := scope.Rows()
//	if err != nil {
//		GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
//			return sqlRaw, -1
//		}, err)
//		return &FirstResult{
//			Error: err,
//		}
//	}
//	defer rows.Close()
//
//	cols, err := rows.Columns()
//	if err != nil {
//		GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
//			return sqlRaw, -1
//		}, err)
//		return &FirstResult{
//			Error: err,
//		}
//	}
//
//	values := make([]sql.RawBytes, len(cols))
//	scanArgs := make([]interface{}, len(values))
//	for i := range values {
//		scanArgs[i] = &values[i]
//	}
//
//	// 把数据写回到out
//	var rowAffected int64
//	for rows.Next() {
//		rowAffected++
//		err = rows.Scan(scanArgs...)
//		if err != nil {
//			GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
//				return sqlRaw, 1
//			}, err)
//			return &FirstResult{
//				Error: err,
//			}
//		}
//
//		if rowAffected != 1 {
//			continue
//		}
//
//		for i, col := range values {
//			if col == nil {
//				continue
//			}
//			field := vv.Elem().FieldByName(stringx.Snake2Camel(cols[i]))
//			if !field.IsValid() {
//				log.Warnf("invalid field: %s", stringx.Snake2Camel(cols[i]))
//				continue
//			}
//			err = decode(field, col)
//			if err != nil {
//				GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
//					return sqlRaw, 1
//				}, err)
//				return &FirstResult{
//					Error: err,
//				}
//			}
//		}
//	}
//
//	if rowAffected == 0 {
//		GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
//			return sqlRaw, 0
//		}, p.getNotFoundError())
//		return &FirstResult{
//			Error: p.getNotFoundError(),
//		}
//	}
//
//	GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
//		return sqlRaw, rowAffected
//	}, nil)
//	return &FirstResult{}
//}
//
//type DeleteResult struct {
//	RowsAffected int64
//	Error        error
//}
//
//func (p *Scoop) Delete() *DeleteResult {
//	if p.cond.skip {
//		return &DeleteResult{}
//	}
//
//	if p.table == "" {
//		panic("table name is empty")
//	}
//
//	if !p.unscoped && p.hasDeletedAt {
//		p.cond.whereRaw("deleted_at = 0")
//	}
//
//	p.inc()
//	defer p.dec()
//
//	sqlRaw := log.GetBuffer()
//	defer log.PutBuffer(sqlRaw)
//
//	// 软删除
//	if !p.unscoped && p.hasDeletedAt {
//		sqlRaw.WriteString("UPDATE")
//		sqlRaw.WriteString(" ")
//		sqlRaw.WriteString(p.table)
//		sqlRaw.WriteString(" SET deleted_at = ")
//		sqlRaw.WriteString(strconv.FormatInt(Now().Unix(), 10))
//	} else {
//		sqlRaw.WriteString("DELETE FROM")
//		sqlRaw.WriteString(" ")
//		sqlRaw.WriteString(p.table)
//	}
//
//	if len(p.cond.conds) > 0 {
//		sqlRaw.WriteString(" WHERE ")
//		sqlRaw.WriteString(p.cond.conds[0])
//		for _, c := range p.cond.conds[1:] {
//			sqlRaw.WriteString(" AND ")
//			sqlRaw.WriteString(c)
//		}
//	}
//
//	start := Now()
//	res, err := p._db.Exec(sqlRaw.String())
//	GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
//		return sqlRaw.String(), res.RowsAffected
//	}, res.Error)
//	return &DeleteResult{
//		RowsAffected: res.RowsAffected,
//		Error:        res.Error,
//	}
//}

//func (p *Scoop) Count() (uint64, error) {
//	if p.cond.skip {
//		return 0, nil
//	}
//
//	if p.table == "" {
//		panic("table name is empty")
//	}
//
//	if !p.unscoped && p.hasDeletedAt {
//		p.cond.whereRaw("deleted_at = 0")
//	}
//
//	p.inc()
//	defer p.dec()
//
//	sqlRaw := log.GetBuffer()
//	defer log.PutBuffer(sqlRaw)
//
//	sqlRaw.WriteString("SELECT COUNT(*) FROM ")
//	sqlRaw.WriteString(p.table)
//
//	if len(p.cond.conds) > 0 {
//		sqlRaw.WriteString(" WHERE ")
//		sqlRaw.WriteString(p.cond.conds[0])
//		for _, c := range p.cond.conds[1:] {
//			sqlRaw.WriteString(" AND ")
//			sqlRaw.WriteString(c)
//		}
//	}
//
//	start := Now()
//	var count uint64
//	err := p._db.Raw(sqlRaw.String()).Scan(&count).Error
//	GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
//		return sqlRaw.String(), int64(count)
//	}, err)
//
//	return count, err
//}
//
//func (p *Scoop) Exist() (bool, error) {
//	if p.cond.skip {
//		return false, nil
//	}
//
//	if p.table == "" {
//		panic("table name is empty")
//	}
//
//	if !p.unscoped && p.hasDeletedAt {
//		p.cond.whereRaw("deleted_at = 0")
//	}
//
//	p.limit = 1
//	p.offset = 0
//
//	p.inc()
//	defer p.dec()
//
//	sqlRaw := log.GetBuffer()
//	defer log.PutBuffer(sqlRaw)
//
//	sqlRaw.WriteString("SELECT ")
//	if p.hasId {
//		sqlRaw.WriteString("id")
//	} else {
//		sqlRaw.WriteString("count(*)")
//	}
//
//	sqlRaw.WriteString(" FROM ")
//	sqlRaw.WriteString(p.table)
//
//	if len(p.cond.conds) > 0 {
//		sqlRaw.WriteString(" WHERE ")
//		sqlRaw.WriteString(p.cond.conds[0])
//		for _, c := range p.cond.conds[1:] {
//			sqlRaw.WriteString(" AND ")
//			sqlRaw.WriteString(c)
//		}
//	}
//
//	sqlRaw.WriteString(" LIMIT 1 OFFSET 0")
//
//	start := Now()
//	var count int64
//	err := p._db.QueryRow(sqlRaw.String()).Scan(&count).Error
//	db.GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
//		return sqlRaw.String(), 0
//	}, err)
//	return count > 0, err
//}

//func (p *Scoop) FindByPage(opt *core.ListOption, values any) (*core.Paginate, error) {
//	p.Offset(opt.Offset).Limit(opt.Limit)
//
//	page := &core.Paginate{
//		Offset: opt.Offset,
//		Limit:  opt.Limit,
//	}
//
//	p.inc()
//	defer p.dec()
//
//	err := p.Find(values).Error
//	if err != nil {
//		log.Errorf("err:%v", err)
//		return nil, err
//	}
//
//	if opt.ShowTotal {
//		page.Total, err = p.Count()
//		if err != nil {
//			log.Errorf("err:%v", err)
//			return nil, err
//		}
//	}
//
//	return page, nil
//}
