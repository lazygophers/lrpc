package db

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
	p.cond.Like(column, value)
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
	p.cond.whereRaw(quoteFieldName(column, p.clientType)+" BETWEEN ? AND ?", min, max)
	return p
}

func (p *Scoop) NotBetween(column string, min, max interface{}) *Scoop {
	p.cond.whereRaw(quoteFieldName(column, p.clientType)+" NOT BETWEEN ? AND ?", min, max)
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

// Join adds a JOIN clause to the query
// Example: Join("INNER", "orders", "users.id = orders.user_id")
func (p *Scoop) Join(joinType, table, condition string) *Scoop {
	p.joins = append(p.joins, joinClause{
		joinType:  joinType,
		table:     table,
		condition: condition,
	})
	return p
}

// InnerJoin adds an INNER JOIN clause to the query
// Example: InnerJoin("orders", "users.id = orders.user_id")
func (p *Scoop) InnerJoin(table, condition string) *Scoop {
	return p.Join("INNER", table, condition)
}

// LeftJoin adds a LEFT JOIN clause to the query
// Example: LeftJoin("orders", "users.id = orders.user_id")
func (p *Scoop) LeftJoin(table, condition string) *Scoop {
	return p.Join("LEFT", table, condition)
}

// RightJoin adds a RIGHT JOIN clause to the query
// Example: RightJoin("orders", "users.id = orders.user_id")
func (p *Scoop) RightJoin(table, condition string) *Scoop {
	return p.Join("RIGHT", table, condition)
}

// FullJoin adds a FULL OUTER JOIN clause to the query
// Example: FullJoin("orders", "users.id = orders.user_id")
func (p *Scoop) FullJoin(table, condition string) *Scoop {
	return p.Join("FULL OUTER", table, condition)
}

// CrossJoin adds a CROSS JOIN clause to the query
// Example: CrossJoin("orders")
func (p *Scoop) CrossJoin(table string) *Scoop {
	return p.Join("CROSS", table, "")
}

// Having adds a HAVING clause to the query (used with GROUP BY)
// Example: Having("COUNT(*) > ?", 5)
func (p *Scoop) Having(args ...interface{}) *Scoop {
	if p.havingCond == nil {
		p.havingCond = &Cond{clientType: p.clientType}
	}
	p.havingCond.Where(args...)
	return p
}
