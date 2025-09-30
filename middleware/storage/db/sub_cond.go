package db

// Where creates a new Cond with AND conditions
func Where(args ...interface{}) *Cond {
	cond := &Cond{isTopLevel: true}
	return cond.Where(args...)
}

// OrWhere creates a new Cond with OR conditions
func OrWhere(args ...interface{}) *Cond {
	cond := &Cond{isTopLevel: true}
	return cond.OrWhere(args...)
}

// Or is an alias for OrWhere
func Or(args ...interface{}) *Cond {
	cond := &Cond{isTopLevel: true}
	return cond.OrWhere(args...)
}

// Like creates a new Cond with LIKE condition
func Like(column string, value string) *Cond {
	cond := &Cond{isTopLevel: true}
	return cond.Like(column, value)
}

// LeftLike creates a new Cond with LEFT LIKE condition (value%)
func LeftLike(column string, value string) *Cond {
	cond := &Cond{isTopLevel: true}
	return cond.LeftLike(column, value)
}

// RightLike creates a new Cond with RIGHT LIKE condition (%value)
func RightLike(column string, value string) *Cond {
	cond := &Cond{isTopLevel: true}
	return cond.RightLike(column, value)
}

// NotLike creates a new Cond with NOT LIKE condition
func NotLike(column string, value string) *Cond {
	cond := &Cond{isTopLevel: true}
	return cond.NotLike(column, value)
}

// NotLeftLike creates a new Cond with NOT LEFT LIKE condition
func NotLeftLike(column string, value string) *Cond {
	cond := &Cond{isTopLevel: true}
	return cond.NotLeftLike(column, value)
}

// NotRightLike creates a new Cond with NOT RIGHT LIKE condition
func NotRightLike(column string, value string) *Cond {
	cond := &Cond{isTopLevel: true}
	return cond.NotRightLike(column, value)
}

// Between creates a new Cond with BETWEEN condition
func Between(column string, min, max interface{}) *Cond {
	cond := &Cond{isTopLevel: true}
	return cond.Between(column, min, max)
}

// NotBetween creates a new Cond with NOT BETWEEN condition
func NotBetween(column string, min, max interface{}) *Cond {
	cond := &Cond{isTopLevel: true}
	return cond.NotBetween(column, min, max)
}
