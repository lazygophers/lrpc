package db

import (
	"github.com/lazygophers/utils/candy"
)

// ——————————条件——————————

func (p *ModelScoop[M]) Select(fields ...string) *ModelScoop[M] {
	p.selects = append(p.selects, fields...)
	return p
}

func (p *ModelScoop[M]) Where(args ...interface{}) *ModelScoop[M] {
	p.cond.Where(args...)
	return p
}

func (p *ModelScoop[M]) Or(args ...interface{}) *ModelScoop[M] {
	p.cond.OrWhere(args...)
	return p
}

func (p *ModelScoop[M]) Equal(column string, value interface{}) *ModelScoop[M] {
	p.cond.where(column, value)
	return p
}

func (p *ModelScoop[M]) NotEqual(column string, value interface{}) *ModelScoop[M] {
	p.cond.where(column, " != ", value)
	return p
}

func (p *ModelScoop[M]) In(column string, values interface{}) *ModelScoop[M] {
	vo := EnsureIsSliceOrArray(values)
	if vo.Len() == 0 {
		p.cond.where(false)
		return p
	}
	p.cond.where(column, "IN", UniqueSlice(vo.Interface()))
	return p
}

func (p *ModelScoop[M]) NotIn(column string, values interface{}) *ModelScoop[M] {
	vo := EnsureIsSliceOrArray(values)
	if vo.Len() == 0 {
		return p
	}
	p.cond.where(column, "NOT IN", UniqueSlice(vo.Interface()))
	return p
}

func (p *ModelScoop[M]) Like(column string, value string) *ModelScoop[M] {
	if value == "" {
		return p
	}
	p.cond.where(column, "LIKE", "%"+value+"%")
	return p
}

func (p *ModelScoop[M]) LeftLike(column string, value string) *ModelScoop[M] {
	if value == "" {
		return p
	}
	p.cond.where(column, "LIKE", value+"%")
	return p
}

func (p *ModelScoop[M]) RightLike(column string, value string) *ModelScoop[M] {
	if value == "" {
		return p
	}
	p.cond.where(column, "LIKE", "%"+value)
	return p
}

func (p *ModelScoop[M]) NotLike(column string, value string) *ModelScoop[M] {
	if value == "" {
		return p
	}
	p.cond.where(column, "NOT LIKE", "%"+value+"%")
	return p
}

func (p *ModelScoop[M]) NotLeftLike(column string, value string) *ModelScoop[M] {
	if value == "" {
		return p
	}
	p.cond.where(column, "NOT LIKE", value+"%")
	return p
}

func (p *ModelScoop[M]) NotRightLike(column string, value string) *ModelScoop[M] {
	if value == "" {
		return p
	}
	p.cond.where(column, "NOT LIKE", "%"+value)
	return p
}

func (p *ModelScoop[M]) Between(column string, min, max interface{}) *ModelScoop[M] {
	p.cond.whereRaw(quoteFieldName(column, p.clientType)+" BETWEEN ? AND ?", min, max)
	return p
}

func (p *ModelScoop[M]) NotBetween(column string, min, max interface{}) *ModelScoop[M] {
	p.cond.whereRaw(quoteFieldName(column, p.clientType)+" NOT BETWEEN ? AND ?", min, max)
	return p
}

func (p *ModelScoop[M]) Unscoped(b ...bool) *ModelScoop[M] {
	if len(b) == 0 {
		p.unscoped = true
		return p
	}
	p.unscoped = b[0]
	return p
}

func (p *ModelScoop[M]) Limit(limit uint64) *ModelScoop[M] {
	p.limit = limit
	return p
}

func (p *ModelScoop[M]) Offset(offset uint64) *ModelScoop[M] {
	p.offset = offset
	return p
}

func (p *ModelScoop[M]) Group(fields ...string) *ModelScoop[M] {
	p.groups = append(p.groups, fields...)
	return p
}

func (p *ModelScoop[M]) Order(fields ...string) *ModelScoop[M] {
	p.orders = append(p.orders, fields...)
	return p
}

func (p *ModelScoop[M]) Desc(fields ...string) *ModelScoop[M] {
	p.orders = append(p.orders, candy.Map(fields, func(s string) string {
		return s + " DESC"
	})...)
	return p
}

func (p *ModelScoop[M]) Asc(fields ...string) *ModelScoop[M] {
	p.orders = append(p.orders, candy.Map(fields, func(s string) string {
		return s + " ASC"
	})...)
	return p
}

func (p *ModelScoop[M]) Ignore(b ...bool) *ModelScoop[M] {
	if len(b) == 0 {
		p.ignore = true
		return p
	}

	p.ignore = b[0]

	return p
}
