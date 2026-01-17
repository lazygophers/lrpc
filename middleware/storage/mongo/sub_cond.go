package mongo

// Where creates a new Cond with AND conditions
func Where(args ...interface{}) *Cond {
	cond := &Cond{}
	return cond.Where(args...)
}

// OrWhere creates a new Cond with OR conditions
func OrWhere(args ...interface{}) *Cond {
	cond := &Cond{}
	return cond.OrWhere(args...)
}

// Or is an alias for OrWhere
func Or(args ...interface{}) *Cond {
	cond := &Cond{}
	return cond.OrWhere(args...)
}

// And creates a new Cond with AND conditions
func And(args ...interface{}) *Cond {
	cond := &Cond{}
	return cond.Where(args...)
}

// Equal creates a new Cond with equality condition
func Equal(column string, value interface{}) *Cond {
	cond := &Cond{}
	return cond.Equal(column, value)
}

// Ne creates a new Cond with != condition using $ne operator
func Ne(column string, value interface{}) *Cond {
	cond := &Cond{}
	return cond.Ne(column, value)
}

// Gt creates a new Cond with > condition using $gt operator
func Gt(column string, value interface{}) *Cond {
	cond := &Cond{}
	return cond.Gt(column, value)
}

// Lt creates a new Cond with < condition using $lt operator
func Lt(column string, value interface{}) *Cond {
	cond := &Cond{}
	return cond.Lt(column, value)
}

// Gte creates a new Cond with >= condition using $gte operator
func Gte(column string, value interface{}) *Cond {
	cond := &Cond{}
	return cond.Gte(column, value)
}

// Lte creates a new Cond with <= condition using $lte operator
func Lte(column string, value interface{}) *Cond {
	cond := &Cond{}
	return cond.Lte(column, value)
}

// In creates a new Cond with $in condition
func In(column string, values ...interface{}) *Cond {
	cond := &Cond{}
	return cond.In(column, values...)
}

// NotIn creates a new Cond with $nin condition
func NotIn(column string, values ...interface{}) *Cond {
	cond := &Cond{}
	return cond.NotIn(column, values...)
}

// Like creates a new Cond with LIKE condition
func Like(column string, value string) *Cond {
	cond := &Cond{}
	return cond.Like(column, value)
}

// LeftLike creates a new Cond with LEFT LIKE condition (pattern%)
func LeftLike(column string, value string) *Cond {
	cond := &Cond{}
	return cond.LeftLike(column, value)
}

// RightLike creates a new Cond with RIGHT LIKE condition (%pattern)
func RightLike(column string, value string) *Cond {
	cond := &Cond{}
	return cond.RightLike(column, value)
}

// NotLike creates a new Cond with NOT LIKE condition
func NotLike(column string, value string) *Cond {
	cond := &Cond{}
	return cond.NotLike(column, value)
}

// NotLeftLike creates a new Cond with NOT LEFT LIKE condition
func NotLeftLike(column string, value string) *Cond {
	cond := &Cond{}
	return cond.NotLeftLike(column, value)
}

// NotRightLike creates a new Cond with NOT RIGHT LIKE condition
func NotRightLike(column string, value string) *Cond {
	cond := &Cond{}
	return cond.NotRightLike(column, value)
}

// Between creates a new Cond with BETWEEN condition
func Between(column string, min, max interface{}) *Cond {
	cond := &Cond{}
	return cond.Between(column, min, max)
}

// NotBetween creates a new Cond with NOT BETWEEN condition
func NotBetween(column string, min, max interface{}) *Cond {
	cond := &Cond{}
	return cond.NotBetween(column, min, max)
}
